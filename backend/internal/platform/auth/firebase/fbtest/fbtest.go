// Package fbtest runs the Firebase Auth emulator in a throwaway container for
// integration tests, and offers helpers to create users and obtain ID tokens
// like a real client would.
//
// It uses the image built from docker/firebase (`make firebase-image`, also run
// by `make test-integration` and in CI) instead of building it here: building
// through testcontainers trips over local Docker credential helpers.
package fbtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ProjectID is the demo project the emulator runs (see docker/firebase).
const ProjectID = "demo-nexo"

// Image is the emulator image tag, shared with docker-compose.yml and the Makefile.
const Image = "nexo-firebase-emulator"

// Emulator is a running Auth emulator.
type Emulator struct {
	// Host is host:port, the value for FIREBASE_AUTH_EMULATOR_HOST.
	Host      string
	container testcontainers.Container
}

// Start runs the Auth emulator from the prebuilt image.
func Start(ctx context.Context) (*Emulator, error) {
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        Image,
			ExposedPorts: []string{"9099/tcp"},
			WaitingFor: wait.ForHTTP("/").WithPort("9099/tcp").
				WithStartupTimeout(3 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting firebase emulator (build it with `make firebase-image`): %w", err)
	}
	host, err := c.PortEndpoint(ctx, "9099/tcp", "")
	if err != nil {
		_ = c.Terminate(ctx)
		return nil, err
	}
	return &Emulator{Host: host, container: c}, nil
}

// Stop removes the container.
func (e *Emulator) Stop(ctx context.Context) { _ = e.container.Terminate(ctx) }

// User is an account created in the emulator.
type User struct {
	UID     string
	IDToken string
}

// SignUp creates an email/password account and returns its ID token.
func (e *Emulator) SignUp(ctx context.Context, email, password string) (User, error) {
	var out struct {
		LocalID string `json:"localId"`
		IDToken string `json:"idToken"`
	}
	err := e.call(ctx, "/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake", "",
		map[string]any{"email": email, "password": password, "returnSecureToken": true}, &out)
	return User{UID: out.LocalID, IDToken: out.IDToken}, err
}

// SignIn returns a fresh ID token (e.g. after changing the account).
func (e *Emulator) SignIn(ctx context.Context, email, password string) (User, error) {
	var out struct {
		LocalID string `json:"localId"`
		IDToken string `json:"idToken"`
	}
	err := e.call(ctx, "/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=fake", "",
		map[string]any{"email": email, "password": password, "returnSecureToken": true}, &out)
	return User{UID: out.LocalID, IDToken: out.IDToken}, err
}

// VerifyEmail marks the account's email as verified (admin operation).
func (e *Emulator) VerifyEmail(ctx context.Context, uid string) error {
	path := fmt.Sprintf("/identitytoolkit.googleapis.com/v1/projects/%s/accounts:update", ProjectID)
	return e.call(ctx, path, "owner", map[string]any{"localId": uid, "emailVerified": true}, nil)
}

func (e *Emulator) call(ctx context.Context, path, bearer string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+e.Host+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var msg bytes.Buffer
		_, _ = msg.ReadFrom(res.Body)
		return fmt.Errorf("emulator %s: %s: %s", path, res.Status, msg.String())
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}
