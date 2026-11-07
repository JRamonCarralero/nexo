package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/config"
)

func setValid(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("FIREBASE_PROJECT_ID", "demo-nexo")
	t.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "")
	t.Setenv("APP_BASE_URL", "")
	t.Setenv("SMTP_TLS", "")
	t.Setenv("SMTP_PORT", "")
}

func TestDefaultsForLocalDevelopment(t *testing.T) {
	setValid(t)

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:5173", cfg.AppBaseURL)
	assert.Equal(t, "localhost", cfg.SMTP.Host)
	assert.Equal(t, 1025, cfg.SMTP.Port)
	assert.Equal(t, "none", cfg.SMTP.TLS)
}

func TestProductionSafetyChecks(t *testing.T) {
	tests := map[string]map[string]string{
		"emulator in production": {"FIREBASE_AUTH_EMULATOR_HOST": "localhost:9099"},
		"plain http app URL":     {"APP_BASE_URL": "http://nexo.example.com", "SMTP_TLS": "tls"},
		"email without TLS":      {"APP_BASE_URL": "https://nexo.example.com", "SMTP_TLS": "none"},
		"invalid app URL":        {"APP_BASE_URL": "nexo.example.com", "SMTP_TLS": "tls"},
		"non-numeric SMTP port":  {"APP_BASE_URL": "https://nexo.example.com", "SMTP_TLS": "tls", "SMTP_PORT": "abc"},
	}
	for name, env := range tests {
		t.Run(name, func(t *testing.T) {
			setValid(t)
			t.Setenv("APP_ENV", "production")
			for k, v := range env {
				t.Setenv(k, v)
			}
			_, err := config.Load()
			assert.Error(t, err)
		})
	}
}

func TestValidProductionConfig(t *testing.T) {
	setValid(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_BASE_URL", "https://nexo.example.com")
	t.Setenv("SMTP_TLS", "tls")
	t.Setenv("SMTP_PORT", "465")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, 465, cfg.SMTP.Port)
}
