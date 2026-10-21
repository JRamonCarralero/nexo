// Package config loads the application configuration from environment variables.
// A .env file in the working directory is loaded first when present (local development).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

func (c Config) IsProduction() bool { return c.Env == EnvProduction }

// Load reads the configuration. Missing optional values get sensible defaults.
func Load() (Config, error) {
	_ = godotenv.Load() // optional: only present in local development

	cfg := Config{
		Env:             getenv("APP_ENV", EnvDevelopment),
		HTTPAddr:        ":" + getenv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ShutdownTimeout: 15 * time.Second,
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	switch cfg.Env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		return Config{}, fmt.Errorf("APP_ENV: unknown environment %q", cfg.Env)
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
