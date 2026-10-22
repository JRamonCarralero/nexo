// Package app is the composition root: the only place that knows every concrete
// implementation and wires repositories, services and handlers together.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JRamonCarralero/nexo/backend/internal/bootstrap"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/config"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
)

// Deps are the clients created by main.
type Deps struct {
	Config        config.Config
	Logger        *slog.Logger
	DB            *pgxpool.Pool
	Authenticator auth.Authenticator
	Version       string
}

type App struct {
	cfg     config.Config
	log     *slog.Logger
	users   *user.Service
	handler http.Handler
}

func New(d Deps) (*App, error) {
	tm := postgres.NewTxManager(d.DB)

	users := user.NewService(user.NewPostgresRepository(tm))

	router, err := httpserver.NewRouter(
		httpserver.Deps{Logger: d.Logger, Authenticator: d.Authenticator, Users: users},
		httpserver.Handlers{
			System: system.NewHandler(d.DB, d.Version),
			User:   user.NewHandler(users),
		},
	)
	if err != nil {
		return nil, err
	}
	return &App{cfg: d.Config, log: d.Logger, users: users, handler: router}, nil
}

// Bootstrap runs the idempotent startup tasks (see package bootstrap).
func (a *App) Bootstrap(ctx context.Context) error {
	return bootstrap.Run(ctx, a.log, bootstrap.Config{SuperadminEmail: a.cfg.BootstrapSuperadminEmail}, a.users)
}

// Handler is the HTTP entry point.
func (a *App) Handler() http.Handler { return a.handler }
