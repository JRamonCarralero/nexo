package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
)

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name     string
		in       error
		wantKind apperror.Kind
		wantCode string
	}{
		{"no rows", fmt.Errorf("get: %w", pgx.ErrNoRows), apperror.KindNotFound, "RESOURCE_NOT_FOUND"},
		{"unique", &pgconn.PgError{Code: "23505"}, apperror.KindConflict, "ALREADY_EXISTS"},
		{"foreign key", &pgconn.PgError{Code: "23503"}, apperror.KindValidation, "REFERENCE_NOT_FOUND"},
		{"not null", &pgconn.PgError{Code: "23502"}, apperror.KindValidation, apperror.CodeInvalidRequest},
		{"rls", &pgconn.PgError{Code: "42501"}, apperror.KindForbidden, apperror.CodeForbidden},
		{"timeout", context.DeadlineExceeded, apperror.KindUnavailable, apperror.CodeUnavailable},
		{"unknown", errors.New("weird"), apperror.KindInternal, apperror.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *apperror.Error
			require.ErrorAs(t, postgres.TranslateError(tt.in), &got)
			assert.Equal(t, tt.wantKind, got.Kind)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.ErrorIs(t, got, tt.in, "the original error must be kept as the cause")
		})
	}
}

func TestTranslateErrorPassThrough(t *testing.T) {
	assert.NoError(t, postgres.TranslateError(nil))

	appErr := apperror.NotFound("task", 1)
	assert.Same(t, appErr, postgres.TranslateError(appErr))
}
