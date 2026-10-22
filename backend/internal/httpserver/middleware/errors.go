package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
)

const appErrorKey = "nexo.apperror"

// genericServerMessage is what clients see for 5xx errors: internal details stay in the logs.
const genericServerMessage = "An unexpected error occurred. Please report the trace ID."

// Errors is the single place where errors become HTTP responses. Handlers only
// return (or c.Error) errors; this middleware renders the last one as RFC 9457
// Problem Details and leaves it in the context for the logging middleware.
func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		appErr := apperror.From(c.Errors.Last().Err)
		c.Set(appErrorKey, appErr)

		if c.Writer.Written() {
			return
		}
		status := StatusFor(appErr.Kind)
		c.Header("Content-Type", "application/problem+json")
		c.AbortWithStatusJSON(status, problemFor(c, appErr, status))
	}
}

// StatusFor maps an error kind to its HTTP status.
func StatusFor(kind apperror.Kind) int {
	switch kind {
	case apperror.KindValidation:
		return http.StatusBadRequest
	case apperror.KindUnauthenticated:
		return http.StatusUnauthorized
	case apperror.KindForbidden:
		return http.StatusForbidden
	case apperror.KindNotFound:
		return http.StatusNotFound
	case apperror.KindConflict:
		return http.StatusConflict
	case apperror.KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func problemFor(c *gin.Context, e *apperror.Error, status int) api.Problem {
	detail := e.Message
	if status >= http.StatusInternalServerError {
		detail = genericServerMessage
	}
	typ := "about:blank"
	p := api.Problem{
		Type:    &typ,
		Title:   http.StatusText(status),
		Status:  status,
		Code:    e.Code,
		Detail:  &detail,
		TraceId: logger.TraceID(c.Request.Context()),
	}
	if len(e.Details) > 0 {
		fields := make([]api.FieldError, len(e.Details))
		for i, d := range e.Details {
			fields[i] = api.FieldError{Field: d.Field, Code: d.Code}
			if d.Message != "" {
				msg := d.Message
				fields[i].Message = &msg
			}
		}
		p.Errors = &fields
	}
	return p
}

// appErrorFrom returns the error the Errors middleware stored, if any.
func appErrorFrom(c *gin.Context) (*apperror.Error, bool) {
	v, ok := c.Get(appErrorKey)
	if !ok {
		return nil, false
	}
	var e *apperror.Error
	return e, errors.As(v.(error), &e)
}
