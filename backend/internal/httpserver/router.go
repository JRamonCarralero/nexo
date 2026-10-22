// Package httpserver wires HTTP routes and middlewares.
package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/middleware"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
)

// BasePath is where the API described in api/openapi.yaml is served.
const BasePath = "/api/v1"

// Handlers groups every module handler. Each new module adds its field here and
// is embedded in server so that, together, they implement the generated interface.
type Handlers struct {
	System *system.Handler
	User   *user.Handler
}

// Every module names its type Handler; the aliases give each embedded field a
// distinct name so their methods are promoted together.
type (
	systemHandler = system.Handler
	userHandler   = user.Handler
)

type server struct {
	*systemHandler
	*userHandler
}

// Deps are the cross-cutting services the HTTP layer needs.
type Deps struct {
	Logger        *slog.Logger
	Authenticator auth.Authenticator
	Users         middleware.IdentityResolver
}

var _ api.StrictServerInterface = server{}

// NewRouter builds the HTTP handler. Middleware order matters: the outer ones
// (trace ID, logging) see the result of the inner ones (errors, recovery, auth).
func NewRouter(d Deps, h Handlers) (*gin.Engine, error) {
	isPublic, err := publicRoutes()
	if err != nil {
		return nil, err
	}

	r := gin.New()
	r.ContextWithFallback = true // handlers read the request context (trace ID, deadlines) through *gin.Context

	r.Use(
		middleware.TraceID(),
		middleware.Logging(d.Logger, BasePath+"/health"),
		middleware.Errors(),
		middleware.Recovery(),
		middleware.Auth(d.Authenticator, d.Users, isPublic),
	)

	r.NoRoute(func(c *gin.Context) {
		_ = c.Error(apperror.NotFound("route", c.Request.URL.Path))
	})

	strict := api.NewStrictHandlerWithOptions(server{h.System, h.User}, nil, api.StrictGinServerOptions{
		RequestErrorHandlerFunc: func(c *gin.Context, err error) {
			_ = c.Error(apperror.Wrap(err, apperror.KindValidation, apperror.CodeInvalidRequest, "invalid request"))
		},
		HandlerErrorFunc: func(c *gin.Context, err error) {
			_ = c.Error(err)
		},
		ResponseErrorHandlerFunc: func(c *gin.Context, err error) {
			_ = c.Error(apperror.Internal(err, "writing response"))
		},
	})
	api.RegisterHandlersWithOptions(r, strict, api.GinServerOptions{BaseURL: BasePath})

	return r, nil
}
