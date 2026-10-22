// Package httpserver wires HTTP routes and middlewares.
package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/middleware"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// BasePath is where the API described in api/openapi.yaml is served.
const BasePath = "/api/v1"

// Handlers groups every module handler. Each new module adds its field here and
// is embedded in server so that, together, they implement the generated interface.
type Handlers struct {
	System *system.Handler
}

type server struct {
	*system.Handler
}

var _ api.StrictServerInterface = server{}

// NewRouter builds the HTTP handler. Middleware order matters: the outer ones
// (trace ID, logging) see the result of the inner ones (errors, recovery).
func NewRouter(log *slog.Logger, h Handlers) *gin.Engine {
	r := gin.New()
	r.ContextWithFallback = true // handlers read the request context (trace ID, deadlines) through *gin.Context

	r.Use(
		middleware.TraceID(),
		middleware.Logging(log, BasePath+"/health"),
		middleware.Errors(),
		middleware.Recovery(),
	)

	r.NoRoute(func(c *gin.Context) {
		_ = c.Error(apperror.NotFound("route", c.Request.URL.Path))
	})

	strict := api.NewStrictHandlerWithOptions(server{h.System}, nil, api.StrictGinServerOptions{
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

	return r
}
