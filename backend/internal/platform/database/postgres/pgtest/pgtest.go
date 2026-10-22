// Package pgtest starts a throwaway PostgreSQL container with the roles and
// all migrations applied, for integration tests. Use it from TestMain, once
// per package.
package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/JRamonCarralero/nexo/backend/db"
)

// Image must match the one in docker-compose.yml.
const Image = "postgres:17-alpine"

// Credentials of the API login role, from db/local/roles.sql.
const (
	appUser     = "nexo_api"
	appPassword = "nexo_api"
)

// Database is a running, migrated test database.
type Database struct {
	// Pool connects as the API role: Row-Level Security applies, like in production.
	Pool *pgxpool.Pool
	// OwnerPool connects as the owner of the tables (bypasses RLS). Use it only
	// to prepare fixtures or to inspect rows regardless of policies.
	OwnerPool *pgxpool.Pool
	container *tcpostgres.PostgresContainer
}

// Start launches the container, creates the roles and applies the migrations.
func Start(ctx context.Context) (*Database, error) {
	container, err := tcpostgres.Run(ctx, Image,
		tcpostgres.WithDatabase("nexo_test"),
		tcpostgres.WithUsername("nexo"),
		tcpostgres.WithPassword("nexo"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres container: %w", err)
	}

	d, err := setup(ctx, container)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	return d, nil
}

func setup(ctx context.Context, container *tcpostgres.PostgresContainer) (*Database, error) {
	ownerURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}
	ownerPool, err := pgxpool.New(ctx, ownerURL)
	if err != nil {
		return nil, err
	}
	if _, err := ownerPool.Exec(ctx, db.LocalRoles); err != nil {
		ownerPool.Close()
		return nil, fmt.Errorf("creating roles: %w", err)
	}
	if err := migrate(ownerURL); err != nil {
		ownerPool.Close()
		return nil, err
	}

	appURL, err := url.Parse(ownerURL)
	if err != nil {
		ownerPool.Close()
		return nil, err
	}
	appURL.User = url.UserPassword(appUser, appPassword)
	pool, err := pgxpool.New(ctx, appURL.String())
	if err != nil {
		ownerPool.Close()
		return nil, err
	}
	return &Database{Pool: pool, OwnerPool: ownerPool, container: container}, nil
}

// Stop closes the pools and removes the container.
func (d *Database) Stop(ctx context.Context) {
	d.Pool.Close()
	d.OwnerPool.Close()
	_ = d.container.Terminate(ctx)
}

func migrate(url string) error {
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	goose.SetBaseFS(db.Migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}
