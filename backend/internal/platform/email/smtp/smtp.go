// Package smtp implements email.Mailer over SMTP. The same adapter serves
// local development (Mailpit, no TLS) and production (Resend's SMTP relay
// with implicit TLS and an API key as password).
package smtp

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	gosmtp "net/smtp"
	"strings"
	"time"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
)

// Security modes.
const (
	TLSNone     = "none"     // plain connection: local development only
	TLSStartTLS = "starttls" // upgrade a plain connection (usually port 587)
	TLSImplicit = "tls"      // TLS from the start (usually port 465)
)

// Config configures the SMTP connection.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      string
	// From is the sender, e.g. "Nexo <no-reply@example.com>".
	From string
	// Timeout bounds the whole delivery (connect + send).
	Timeout time.Duration
}

type Mailer struct {
	cfg  Config
	from *mail.Address
}

var _ email.Mailer = (*Mailer)(nil)

// New validates the configuration and returns a mailer.
func New(cfg Config) (*Mailer, error) {
	if cfg.Host == "" || cfg.Port == 0 {
		return nil, errors.New("smtp: host and port are required")
	}
	switch cfg.TLS {
	case TLSNone, TLSStartTLS, TLSImplicit:
	default:
		return nil, fmt.Errorf("smtp: unknown TLS mode %q", cfg.TLS)
	}
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("smtp: invalid sender %q: %w", cfg.From, err)
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &Mailer{cfg: cfg, from: from}, nil
}

func (m *Mailer) Send(ctx context.Context, msg email.Message) error {
	to, err := mail.ParseAddress(msg.To)
	if err != nil {
		return apperror.Validation("INVALID_EMAIL", "invalid recipient: "+msg.To)
	}
	body, err := build(m.from, to, msg)
	if err != nil {
		return apperror.Internal(err, "building email")
	}
	ctx, cancel := context.WithTimeout(ctx, m.cfg.Timeout)
	defer cancel()
	if err := m.deliver(ctx, to.Address, body); err != nil {
		return apperror.Unavailable(err, "email delivery failed")
	}
	return nil
}

func (m *Mailer) deliver(ctx context.Context, to string, body []byte) error {
	addr := net.JoinHostPort(m.cfg.Host, fmt.Sprint(m.cfg.Port))
	tlsConfig := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	var err error
	dialer := &net.Dialer{}
	if m.cfg.TLS == TLSImplicit {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", addr, err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := gosmtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()

	if m.cfg.TLS == TLSStartTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}
	if m.cfg.Username != "" {
		if err := client.Auth(gosmtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}
	if err := client.Mail(m.from.Address); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// build writes a MIME message: plain text, or multipart/alternative with HTML.
func build(from, to *mail.Address, msg email.Message) ([]byte, error) {
	var b bytes.Buffer
	header := func(k, v string) { fmt.Fprintf(&b, "%s: %s\r\n", k, v) }

	header("From", from.String())
	header("To", to.String())
	header("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	header("Date", time.Now().UTC().Format(time.RFC1123Z))
	header("Message-ID", fmt.Sprintf("<%s@%s>", randomID(), domainOf(from.Address)))
	header("MIME-Version", "1.0")

	if msg.HTML == "" {
		header("Content-Type", `text/plain; charset="utf-8"`)
		header("Content-Transfer-Encoding", "quoted-printable")
		b.WriteString("\r\n")
		if err := writeQP(&b, msg.Text); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}

	boundary := "nexo-" + randomID()
	header("Content-Type", fmt.Sprintf(`multipart/alternative; boundary="%s"`, boundary))
	b.WriteString("\r\n")
	for _, part := range []struct{ typ, content string }{{"text/plain", msg.Text}, {"text/html", msg.HTML}} {
		fmt.Fprintf(&b, "--%s\r\n", boundary)
		fmt.Fprintf(&b, "Content-Type: %s; charset=\"utf-8\"\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", part.typ)
		if err := writeQP(&b, part.content); err != nil {
			return nil, err
		}
		b.WriteString("\r\n")
	}
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return b.Bytes(), nil
}

func writeQP(b *bytes.Buffer, s string) error {
	w := quotedprintable.NewWriter(b)
	if _, err := w.Write([]byte(strings.ReplaceAll(s, "\n", "\r\n"))); err != nil {
		return err
	}
	return w.Close()
}

func randomID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func domainOf(address string) string {
	if _, domain, ok := strings.Cut(address, "@"); ok {
		return domain
	}
	return "localhost"
}
