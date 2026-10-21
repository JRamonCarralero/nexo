// Package apperror defines the application's error type. It is transport-agnostic:
// the HTTP layer maps a Kind to a status code and renders it as Problem Details.
package apperror

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Kind classifies an error. Each kind maps to exactly one HTTP status.
type Kind string

const (
	KindValidation      Kind = "VALIDATION"
	KindNotFound        Kind = "NOT_FOUND"
	KindConflict        Kind = "CONFLICT"
	KindUnauthenticated Kind = "UNAUTHENTICATED"
	KindForbidden       Kind = "FORBIDDEN"
	KindUnavailable     Kind = "UNAVAILABLE"
	KindInternal        Kind = "INTERNAL"
)

// Generic codes, used when no more specific code applies.
const (
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeUnauthenticated = "UNAUTHENTICATED"
	CodeForbidden       = "FORBIDDEN"
	CodeUnavailable     = "SERVICE_UNAVAILABLE"
	CodeInternal        = "INTERNAL_ERROR"
)

// FieldError describes a validation problem on a single input field.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// Error is the application error.
//
// Code is stable and machine-readable (e.g. TASK_NOT_FOUND): the frontend
// translates messages from it. Message is meant for developers and logs.
type Error struct {
	Kind     Kind
	Code     string
	Message  string
	Details  []FieldError
	Location string // file:line where the error was created
	Err      error  // underlying cause, never exposed to clients
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", e.Code, e.Message)
	if e.Err != nil {
		fmt.Fprintf(&b, ": %v", e.Err)
	}
	return b.String()
}

func (e *Error) Unwrap() error { return e.Err }

// Is matches two *Error values by Code, so callers can do
// errors.Is(err, apperror.NotFound("task", nil)) style comparisons on sentinels.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Code == e.Code
}

// New creates an error of the given kind. Prefer the named constructors.
func New(kind Kind, code, message string) *Error {
	return newError(kind, code, message, nil, 2)
}

// Wrap creates an error of the given kind with an underlying cause.
func Wrap(err error, kind Kind, code, message string) *Error {
	return newError(kind, code, message, err, 2)
}

// NotFound reports that a resource does not exist (code: <RESOURCE>_NOT_FOUND).
func NotFound(resource string, id any) *Error {
	code := strings.ToUpper(resource) + "_NOT_FOUND"
	msg := resource + " not found"
	if id != nil {
		msg = fmt.Sprintf("%s %v not found", resource, id)
	}
	return newError(KindNotFound, code, msg, nil, 2)
}

// Validation reports invalid input, optionally with per-field details.
func Validation(code, message string, details ...FieldError) *Error {
	e := newError(KindValidation, code, message, nil, 2)
	e.Details = details
	return e
}

// Conflict reports that the operation clashes with the current state.
func Conflict(code, message string) *Error {
	return newError(KindConflict, code, message, nil, 2)
}

// Unauthenticated reports a missing or invalid identity.
func Unauthenticated(message string) *Error {
	return newError(KindUnauthenticated, CodeUnauthenticated, message, nil, 2)
}

// Forbidden reports that the identity lacks permission.
func Forbidden(code, message string) *Error {
	return newError(KindForbidden, code, message, nil, 2)
}

// Unavailable reports that a dependency is temporarily down.
func Unavailable(err error, message string) *Error {
	return newError(KindUnavailable, CodeUnavailable, message, err, 2)
}

// Internal wraps an unexpected error.
func Internal(err error, message string) *Error {
	return newError(KindInternal, CodeInternal, message, err, 2)
}

// From returns err as an *Error. Errors that are not application errors
// become Internal, keeping the original as the cause.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return newError(KindInternal, CodeInternal, "unexpected error", err, 2)
}

func newError(kind Kind, code, message string, err error, skip int) *Error {
	e := &Error{Kind: kind, Code: code, Message: message, Err: err}
	if _, file, line, ok := runtime.Caller(skip); ok {
		e.Location = fmt.Sprintf("%s:%d", trimPath(file), line)
	}
	return e
}

// trimPath keeps the path from "internal/" or "cmd/" onwards for readable logs.
func trimPath(file string) string {
	for _, marker := range []string{"/internal/", "/cmd/"} {
		if i := strings.LastIndex(file, marker); i >= 0 {
			return file[i+1:]
		}
	}
	return file
}
