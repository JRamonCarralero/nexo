// Package app is the composition root: the only place that knows every concrete
// implementation and wires repositories, services and handlers together.
package app

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/config"
)

// Deps are the clients created by main.
type Deps struct {
	Config  config.Config
	Logger  *slog.Logger
	DB      *pgxpool.Pool
	Version string
}

type App struct {
	handler http.Handler
}

func New(d Deps) *App {
	handlers := httpserver.Handlers{
		System: system.NewHandler(d.DB, d.Version),
	}
	return &App{handler: httpserver.NewRouter(d.Logger, handlers)}
}

// Handler is the HTTP entry point.
func (a *App) Handler() http.Handler { return a.handler }
