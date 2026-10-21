// Package dbtest starts a throwaway PostgreSQL container with all migrations
// applied, for integration tests. Use it from TestMain, once per package.
package dbtest

import (
	"context"
	"database/sql"
	"fmt"
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

// Database is a running, migrated test database.
type Database struct {
	Pool      *pgxpool.Pool
	URL       string
	container *tcpostgres.PostgresContainer
}

// Start launches the container and applies the migrations.
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

	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	if err := migrate(url); err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	return &Database{Pool: pool, URL: url, container: container}, nil
}

// Stop closes the pool and removes the container.
func (d *Database) Stop(ctx context.Context) {
	d.Pool.Close()
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
