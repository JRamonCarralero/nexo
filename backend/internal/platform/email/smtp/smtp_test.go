package smtp_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email/smtp"
)

func TestNewValidatesTheConfig(t *testing.T) {
	ok := smtp.Config{Host: "localhost", Port: 1025, TLS: smtp.TLSNone, From: "Nexo <no-reply@nexo.local>"}

	_, err := smtp.New(ok)
	require.NoError(t, err)

	for name, change := range map[string]func(*smtp.Config){
		"no host":        func(c *smtp.Config) { c.Host = "" },
		"no port":        func(c *smtp.Config) { c.Port = 0 },
		"bad TLS mode":   func(c *smtp.Config) { c.TLS = "ssl" },
		"invalid sender": func(c *smtp.Config) { c.From = "not an address" },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := ok
			change(&cfg)
			_, err := smtp.New(cfg)
			assert.Error(t, err)
		})
	}
}

func TestInvalidRecipientIsAValidationError(t *testing.T) {
	m, err := smtp.New(smtp.Config{Host: "localhost", Port: 1025, TLS: smtp.TLSNone, From: "no-reply@nexo.local"})
	require.NoError(t, err)

	err = m.Send(context.Background(), email.Message{To: "nope", Subject: "x", Text: "x"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.KindValidation, appErr.Kind)
}
