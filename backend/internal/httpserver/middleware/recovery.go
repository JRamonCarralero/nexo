package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// Recovery turns a panic into an Internal error (with stack trace) so it is
// rendered and logged like any other error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if p := recover(); p != nil {
				err := fmt.Errorf("panic: %v\n%s", p, debug.Stack())
				_ = c.Error(apperror.Internal(err, "panic recovered"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
