// Package app is the composition root: the only place that knows every concrete
// implementation and wires repositories, services and handlers together.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JRamonCarralero/nexo/backend/internal/bootstrap"
	"github.com/JRamonCarralero/nexo/backend/internal/devseed"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/config"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
)

// Deps are the clients created by main.
type Deps struct {
	Config        config.Config
	Logger        *slog.Logger
	DB            *pgxpool.Pool
	Authenticator auth.Authenticator
	Mailer        email.Mailer
	Version       string
}

type App struct {
	cfg     config.Config
	log     *slog.Logger
	users   *user.Service
	seed    devseed.Services
	handler http.Handler
}

func New(d Deps) (*App, error) {
	tm := postgres.NewTxManager(d.DB)

	users := user.NewService(user.NewPostgresRepository(tm))
	workspaceRepo := workspace.NewPostgresRepository(tm)
	workspaces := workspace.NewService(workspaceRepo, userDirectory{users})
	invitations := workspace.NewInvitationService(workspaceRepo, workspaceRepo, userDirectory{users}, d.Mailer, d.Config.AppBaseURL)

	router, err := httpserver.NewRouter(
		httpserver.Deps{Logger: d.Logger, Authenticator: d.Authenticator, Users: users, Workspaces: workspaces},
		httpserver.Handlers{
			System:    system.NewHandler(d.DB, d.Version),
			User:      user.NewHandler(users),
			Workspace: workspace.NewHandler(workspaces, invitations),
		},
	)
	if err != nil {
		return nil, err
	}
	return &App{
		cfg: d.Config, log: d.Logger, users: users, handler: router,
		seed: devseed.Services{Users: users, Workspaces: workspaces, Invitations: invitations},
	}, nil
}

// Bootstrap runs the idempotent startup tasks (see package bootstrap).
func (a *App) Bootstrap(ctx context.Context) error {
	return bootstrap.Run(ctx, a.log, bootstrap.Config{SuperadminEmail: a.cfg.BootstrapSuperadminEmail}, a.users)
}

// Seed fills the database with example data (local development only: the
// caller, cmd/seed, refuses to run it in production).
func (a *App) Seed(ctx context.Context, accounts devseed.Accounts) error {
	return devseed.Run(ctx, a.log, a.seed, accounts)
}

// Handler is the HTTP entry point.
func (a *App) Handler() http.Handler { return a.handler }
