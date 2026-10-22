package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// IdentityResolver maps a verified identity to a Nexo user id, creating the
// user on the first login.
type IdentityResolver interface {
	Resolve(ctx context.Context, identity auth.Identity) (uuid.UUID, error)
}

// IsPublic reports whether the route (method + gin route pattern) can be
// called without authentication.
type IsPublic func(method, route string) bool

// Auth requires a valid bearer token on every route that is not public, and
// puts the authenticated user in the request scope.
//
// Secure by default: a route is public only if IsPublic says so. Requests
// that match no route (FullPath empty) pass through and end as 404.
func Auth(authn auth.Authenticator, users IdentityResolver, isPublic IsPublic) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" || isPublic(c.Request.Method, route) {
			c.Next()
			return
		}

		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abort(c, apperror.New(apperror.KindUnauthenticated, "MISSING_TOKEN", "missing bearer token"))
			return
		}
		identity, err := authn.VerifyToken(c.Request.Context(), token)
		if err != nil {
			abort(c, err)
			return
		}
		userID, err := users.Resolve(c.Request.Context(), identity)
		if err != nil {
			abort(c, err)
			return
		}

		c.Request = c.Request.WithContext(scope.WithUser(c.Request.Context(), userID))
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}

func abort(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}
