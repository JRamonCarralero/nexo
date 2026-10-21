package apperror_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

func TestConstructors(t *testing.T) {
	cause := errors.New("boom")

	tests := []struct {
		name     string
		err      *apperror.Error
		wantKind apperror.Kind
		wantCode string
	}{
		{"not found", apperror.NotFound("task", 42), apperror.KindNotFound, "TASK_NOT_FOUND"},
		{"validation", apperror.Validation("TITLE_REQUIRED", "title is required"), apperror.KindValidation, "TITLE_REQUIRED"},
		{"conflict", apperror.Conflict("SLUG_TAKEN", "slug taken"), apperror.KindConflict, "SLUG_TAKEN"},
		{"unauthenticated", apperror.Unauthenticated("no token"), apperror.KindUnauthenticated, apperror.CodeUnauthenticated},
		{"forbidden", apperror.Forbidden("NOT_A_MEMBER", "not a member"), apperror.KindForbidden, "NOT_A_MEMBER"},
		{"unavailable", apperror.Unavailable(cause, "db down"), apperror.KindUnavailable, apperror.CodeUnavailable},
		{"internal", apperror.Internal(cause, "oops"), apperror.KindInternal, apperror.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKind, tt.err.Kind)
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.True(t, strings.HasPrefix(tt.err.Location, "internal/platform/apperror/apperror_test.go:"),
				"location should point to the caller, got %q", tt.err.Location)
		})
	}
}

func TestNotFoundMessage(t *testing.T) {
	assert.Equal(t, "task 42 not found", apperror.NotFound("task", 42).Message)
	assert.Equal(t, "task not found", apperror.NotFound("task", nil).Message)
}

func TestUnwrapAndAs(t *testing.T) {
	cause := errors.New("connection refused")
	wrapped := fmt.Errorf("loading project: %w", apperror.Unavailable(cause, "database unavailable"))

	var appErr *apperror.Error
	require.True(t, errors.As(wrapped, &appErr))
	assert.Equal(t, apperror.KindUnavailable, appErr.Kind)
	assert.ErrorIs(t, wrapped, cause)
}

func TestIsMatchesByCode(t *testing.T) {
	err := fmt.Errorf("ctx: %w", apperror.NotFound("task", 1))
	assert.ErrorIs(t, err, apperror.NotFound("task", nil))
	assert.NotErrorIs(t, err, apperror.NotFound("project", nil))
}

func TestFrom(t *testing.T) {
	assert.Nil(t, apperror.From(nil))

	original := apperror.Conflict("X", "x")
	assert.Same(t, original, apperror.From(fmt.Errorf("wrap: %w", original)))

	plain := errors.New("plain")
	got := apperror.From(plain)
	assert.Equal(t, apperror.KindInternal, got.Kind)
	assert.ErrorIs(t, got, plain)
}

func TestValidationDetails(t *testing.T) {
	err := apperror.Validation(apperror.CodeInvalidRequest, "invalid task",
		apperror.FieldError{Field: "title", Code: "REQUIRED"},
	)
	require.Len(t, err.Details, 1)
	assert.Equal(t, "title", err.Details[0].Field)
}

func TestErrorString(t *testing.T) {
	err := apperror.Internal(errors.New("boom"), "saving task")
	assert.Equal(t, "[INTERNAL_ERROR] saving task: boom", err.Error())
}
