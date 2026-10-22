package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// PostgreSQL error codes we translate. See https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgNotNullViolation    = "23502"
	pgCheckViolation      = "23514"
	pgInvalidTextRepr     = "22P02"
	pgInsufficientPriv    = "42501" // raised by Row-Level Security policies
)

// TranslateError converts database errors into application errors so modules
// never deal with PostgreSQL codes. Errors that are already application errors
// pass through unchanged.
func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.Wrap(err, apperror.KindNotFound, "RESOURCE_NOT_FOUND", "resource not found")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperror.Unavailable(err, "database timeout")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return apperror.Wrap(err, apperror.KindConflict, "ALREADY_EXISTS", "resource already exists")
		case pgForeignKeyViolation:
			return apperror.Wrap(err, apperror.KindValidation, "REFERENCE_NOT_FOUND", "referenced resource does not exist")
		case pgNotNullViolation, pgCheckViolation, pgInvalidTextRepr:
			return apperror.Wrap(err, apperror.KindValidation, apperror.CodeInvalidRequest, "invalid data")
		case pgInsufficientPriv:
			return apperror.Wrap(err, apperror.KindForbidden, apperror.CodeForbidden, "access denied")
		}
	}

	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		return apperror.Unavailable(err, "database unavailable")
	}

	return apperror.Internal(err, "database error")
}
