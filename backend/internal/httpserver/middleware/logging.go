package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging writes exactly one log line per request. Errors are logged here and
// nowhere else: 4xx as WARN without internals, 5xx as ERROR with code
// location and the full cause chain.
func Logging(log *slog.Logger, quietPaths ...string) gin.HandlerFunc {
	quiet := make(map[string]bool, len(quietPaths))
	for _, p := range quietPaths {
		quiet[p] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", c.FullPath()),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
		}

		level := slog.LevelInfo
		if quiet[c.FullPath()] {
			level = slog.LevelDebug
		}
		msg := "request"

		if appErr, ok := appErrorFrom(c); ok {
			attrs = append(attrs, slog.String("error_code", appErr.Code), slog.String("error_kind", string(appErr.Kind)))
			msg = appErr.Message
			level = slog.LevelWarn
			if status >= http.StatusInternalServerError {
				level = slog.LevelError
				attrs = append(attrs, slog.String("error_location", appErr.Location))
				if appErr.Err != nil {
					attrs = append(attrs, slog.String("error_cause", appErr.Err.Error()))
				}
			}
		}

		log.LogAttrs(c.Request.Context(), level, msg, attrs...)
	}
}
