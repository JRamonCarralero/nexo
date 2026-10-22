// Package firebase implements auth.Authenticator with Firebase Authentication.
//
// When FIREBASE_AUTH_EMULATOR_HOST is set (local development and tests) the
// Admin SDK talks to the emulator and accepts its unsigned tokens; in
// production it verifies the token signature against Google's public keys.
package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"

	fb "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
)

// EmulatorHostEnv is read by the Admin SDK to switch to the emulator.
const EmulatorHostEnv = "FIREBASE_AUTH_EMULATOR_HOST"

type Authenticator struct {
	client *fbauth.Client
}

var _ auth.Authenticator = (*Authenticator)(nil)

// New creates the authenticator for the given Firebase project.
func New(ctx context.Context, projectID string) (*Authenticator, error) {
	if projectID == "" {
		return nil, errors.New("firebase: project id is required")
	}
	var opts []option.ClientOption
	if os.Getenv(EmulatorHostEnv) != "" {
		// The emulator needs no credentials (and "demo-" projects have none).
		opts = append(opts, option.WithoutAuthentication())
	}
	app, err := fb.NewApp(ctx, &fb.Config{ProjectID: projectID}, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase: creating app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: creating auth client: %w", err)
	}
	return &Authenticator{client: client}, nil
}

func (a *Authenticator) VerifyToken(ctx context.Context, token string) (auth.Identity, error) {
	t, err := a.client.VerifyIDToken(ctx, token)
	if err != nil {
		return auth.Identity{}, apperror.Wrap(err, apperror.KindUnauthenticated, "INVALID_TOKEN", "invalid or expired token")
	}
	id := auth.Identity{
		Subject:       t.UID,
		Email:         claim[string](t.Claims, "email"),
		EmailVerified: claim[bool](t.Claims, "email_verified"),
		Name:          claim[string](t.Claims, "name"),
		AvatarURL:     claim[string](t.Claims, "picture"),
		Provider:      t.Firebase.SignInProvider,
	}
	if id.Email == "" {
		// Nexo identifies people by email (invitations, superadmin…).
		return auth.Identity{}, apperror.New(apperror.KindUnauthenticated, "EMAIL_REQUIRED", "the account has no email address")
	}
	return id, nil
}

func claim[T any](claims map[string]any, key string) T {
	v, _ := claims[key].(T)
	return v
}
