package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
)

// TraceHeader is returned on every response so clients can report it.
const TraceHeader = "X-Trace-Id"

var traceIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// TraceID takes the trace ID from the incoming request (W3C traceparent or
// Cloud Run's X-Cloud-Trace-Context) or generates one, stores it in the
// request context for logging and echoes it in the response.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := incomingTraceID(c)
		if id == "" {
			id = newTraceID()
		}
		c.Request = c.Request.WithContext(logger.WithTraceID(c.Request.Context(), id))
		c.Header(TraceHeader, id)
		c.Next()
	}
}

func incomingTraceID(c *gin.Context) string {
	// traceparent: 00-<trace-id>-<span-id>-<flags>
	if parts := strings.Split(c.GetHeader("traceparent"), "-"); len(parts) == 4 && traceIDPattern.MatchString(parts[1]) {
		return parts[1]
	}
	// X-Cloud-Trace-Context: <trace-id>/<span-id>;o=<options>
	if v := c.GetHeader("X-Cloud-Trace-Context"); v != "" {
		id, _, _ := strings.Cut(v, "/")
		if id = strings.ToLower(id); traceIDPattern.MatchString(id) {
			return id
		}
	}
	return ""
}

func newTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
