// Package config loads the application configuration from environment variables.
// A .env file in the working directory is loaded first when present (local development).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
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

	// FirebaseProjectID is the Firebase project whose ID tokens are accepted.
	FirebaseProjectID string
	// FirebaseAuthEmulatorHost points the Admin SDK to the Auth emulator (local only).
	FirebaseAuthEmulatorHost string
	// BootstrapSuperadminEmail, when set, makes sure the platform has a superadmin.
	BootstrapSuperadminEmail string

	// AppBaseURL is the web app's public URL, used in links sent to people
	// (e.g. invitations).
	AppBaseURL string
	// SMTP delivers email: Mailpit locally, Resend's SMTP relay in production.
	SMTP SMTPConfig
}

// SMTPConfig configures outgoing email.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	// TLS is none (local only), starttls or tls.
	TLS  string
	From string
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

		FirebaseProjectID:        os.Getenv("FIREBASE_PROJECT_ID"),
		FirebaseAuthEmulatorHost: os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"),
		BootstrapSuperadminEmail: os.Getenv("BOOTSTRAP_SUPERADMIN_EMAIL"),

		AppBaseURL: getenv("APP_BASE_URL", "http://localhost:5173"),
		SMTP: SMTPConfig{
			Host:     getenv("SMTP_HOST", "localhost"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			TLS:      getenv("SMTP_TLS", "none"),
			From:     getenv("EMAIL_FROM", "Nexo <no-reply@nexo.local>"),
		},
	}

	port, err := strconv.Atoi(getenv("SMTP_PORT", "1025"))
	if err != nil {
		return Config{}, fmt.Errorf("SMTP_PORT: %w", err)
	}
	cfg.SMTP.Port = port
	if u, err := url.Parse(cfg.AppBaseURL); err != nil || u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("APP_BASE_URL: invalid URL %q", cfg.AppBaseURL)
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
	if cfg.FirebaseProjectID == "" {
		return Config{}, errors.New("FIREBASE_PROJECT_ID is required")
	}
	if cfg.IsProduction() && cfg.FirebaseAuthEmulatorHost != "" {
		// The emulator accepts unsigned tokens: it must never reach production.
		return Config{}, errors.New("FIREBASE_AUTH_EMULATOR_HOST must not be set in production")
	}
	if cfg.IsProduction() {
		// Invitation links carry tokens: they must travel over HTTPS, and so must email.
		if u, _ := url.Parse(cfg.AppBaseURL); u.Scheme != "https" {
			return Config{}, errors.New("APP_BASE_URL must use https in production")
		}
		if cfg.SMTP.TLS == "none" {
			return Config{}, errors.New("SMTP_TLS must be starttls or tls in production")
		}
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
