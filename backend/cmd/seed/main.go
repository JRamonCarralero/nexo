// Command seed fills the local database with example data (make db-seed).
// It refuses to run in production or without the Firebase Auth emulator.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/JRamonCarralero/nexo/backend/internal/app"
	"github.com/JRamonCarralero/nexo/backend/internal/devseed"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth/firebase"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/config"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg.Env != config.EnvDevelopment {
		return fmt.Errorf("refusing to seed: APP_ENV is %q, example data is for development only", cfg.Env)
	}
	if cfg.FirebaseAuthEmulatorHost == "" {
		return errors.New("refusing to seed: " + firebase.EmulatorHostEnv + " is not set (run `make up`)")
	}
	gin.SetMode(gin.ReleaseMode) // the router is built but never served: keep the output clean
	log := logger.New(os.Stdout, cfg.Env, cfg.LogLevel)
	slog.SetDefault(log)

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	authn, err := firebase.New(ctx, cfg.FirebaseProjectID)
	if err != nil {
		return err
	}
	accounts, err := firebase.NewEmulatorAccounts(cfg.FirebaseAuthEmulatorHost, cfg.FirebaseProjectID)
	if err != nil {
		return err
	}

	// The example invitations are not delivered: nobody is waiting for them.
	application, err := app.New(app.Deps{
		Config: cfg, Logger: log, DB: db, Authenticator: authn, Mailer: email.Discard{}, Version: "seed",
	})
	if err != nil {
		return err
	}
	if err := application.Seed(ctx, accounts); err != nil {
		return err
	}

	fmt.Printf("\nExample data ready. Sign in at %s with any of:\n", cfg.AppBaseURL)
	for _, e := range devseed.People() {
		fmt.Printf("  %-20s password: %s\n", e, devseed.Password)
	}
	fmt.Println(strings.Repeat("-", 8))
	return nil
}
