package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
)

// EmulatorAccounts creates accounts in the Auth emulator, for local seed data
// and demos. It refuses to work without an emulator host: it must never be
// pointed at a real Firebase project.
type EmulatorAccounts struct {
	host      string
	projectID string
	client    *http.Client
}

// NewEmulatorAccounts targets the emulator at host (host:port).
func NewEmulatorAccounts(host, projectID string) (*EmulatorAccounts, error) {
	if host == "" {
		return nil, errors.New("firebase: " + EmulatorHostEnv + " is required to create accounts")
	}
	if projectID == "" {
		return nil, errors.New("firebase: project id is required")
	}
	return &EmulatorAccounts{host: host, projectID: projectID, client: &http.Client{Timeout: 10 * time.Second}}, nil
}

// Ensure makes sure an account with that email exists, with the given
// password and name and a verified email, and returns its identity. It can be
// run again: an existing account keeps its uid.
func (e *EmulatorAccounts) Ensure(ctx context.Context, email, password, name string) (auth.Identity, error) {
	var out struct {
		LocalID string `json:"localId"`
		IDToken string `json:"idToken"`
	}
	body := map[string]any{"email": email, "password": password, "returnSecureToken": true}
	err := e.post(ctx, "/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake", "", body, &out)
	if err != nil && strings.Contains(err.Error(), "EMAIL_EXISTS") {
		err = e.post(ctx, "/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=fake", "", body, &out)
	}
	if err != nil {
		return auth.Identity{}, fmt.Errorf("firebase emulator: account %s: %w", email, err)
	}

	update := map[string]any{"localId": out.LocalID, "displayName": name, "emailVerified": true}
	path := fmt.Sprintf("/identitytoolkit.googleapis.com/v1/projects/%s/accounts:update", e.projectID)
	if err := e.post(ctx, path, "owner", update, nil); err != nil {
		return auth.Identity{}, fmt.Errorf("firebase emulator: updating %s: %w", email, err)
	}
	return auth.Identity{
		Subject: out.LocalID, Email: email, EmailVerified: true, Name: name, Provider: "password",
	}, nil
}

func (e *EmulatorAccounts) post(ctx context.Context, path, bearer string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+e.host+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("%s: %s", res.Status, msg)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}
