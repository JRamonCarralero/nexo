//go:build integration

package smtp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email/smtp"
)

// MailpitImage must match docker-compose.yml.
const MailpitImage = "axllent/mailpit:v1.33.0"

var (
	smtpHost string
	smtpPort int
	apiURL   string
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        MailpitImage,
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForHTTP("/api/v1/info").WithPort("8025/tcp"),
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	smtpHost, _ = c.Host(ctx)
	port, _ := c.MappedPort(ctx, "1025/tcp")
	smtpPort, _ = strconv.Atoi(port.Port())
	api, _ := c.PortEndpoint(ctx, "8025/tcp", "http")
	apiURL = api

	code := m.Run()
	_ = c.Terminate(ctx)
	os.Exit(code)
}

func newMailer(t *testing.T) *smtp.Mailer {
	t.Helper()
	m, err := smtp.New(smtp.Config{Host: smtpHost, Port: smtpPort, TLS: smtp.TLSNone, From: "Nexo <no-reply@nexo.local>"})
	require.NoError(t, err)
	return m
}

type mailpitMessage struct {
	ID      string
	Subject string
	From    struct{ Name, Address string }
	To      []struct{ Address string }
	Text    string
	HTML    string
}

// lastMessageTo returns the latest message received for the address.
func lastMessageTo(t *testing.T, address string) mailpitMessage {
	t.Helper()
	var list struct {
		Messages []mailpitMessage `json:"messages"`
	}
	getJSON(t, apiURL+"/api/v1/search?query=to:"+address, &list)
	require.NotEmpty(t, list.Messages, "no email received for %s", address)
	var full mailpitMessage
	getJSON(t, apiURL+"/api/v1/message/"+list.Messages[0].ID, &full)
	return full
}

func getJSON(t *testing.T, url string, out any) {
	t.Helper()
	res, err := http.Get(url) //nolint:noctx // test helper
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(out))
}

func TestSendsTextAndHTML(t *testing.T) {
	err := newMailer(t).Send(context.Background(), email.Message{
		To:      "ana@example.com",
		Subject: "Invitación a «Diseño» ✨",
		Text:    "Hola Ana,\nte han invitado.",
		HTML:    "<p>Hola <strong>Ana</strong>, te han invitado.</p>",
	})
	require.NoError(t, err)

	got := lastMessageTo(t, "ana@example.com")
	assert.Equal(t, "Invitación a «Diseño» ✨", got.Subject, "non-ASCII subjects are encoded correctly")
	assert.Equal(t, "no-reply@nexo.local", got.From.Address)
	assert.Contains(t, got.Text, "te han invitado.")
	assert.Contains(t, got.HTML, "<strong>Ana</strong>")
}

func TestSendsPlainTextOnly(t *testing.T) {
	err := newMailer(t).Send(context.Background(), email.Message{To: "luis@example.com", Subject: "Hi", Text: "Only text"})
	require.NoError(t, err)

	got := lastMessageTo(t, "luis@example.com")
	assert.Contains(t, got.Text, "Only text")
	assert.Empty(t, got.HTML)
}

func TestUnreachableServerIsUnavailable(t *testing.T) {
	m, err := smtp.New(smtp.Config{Host: "127.0.0.1", Port: 1, TLS: smtp.TLSNone, From: "no-reply@nexo.local"})
	require.NoError(t, err)

	err = m.Send(context.Background(), email.Message{To: "ana@example.com", Subject: "x", Text: "x"})

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.KindUnavailable, appErr.Kind)
}
