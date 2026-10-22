// Package bootstrap runs the idempotent tasks the platform needs before
// serving requests, in every environment. It never deletes or demotes
// anything; running it again is always safe.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
)

// SuperadminEnsurer makes sure the platform has a superadmin.
type SuperadminEnsurer interface {
	EnsureSuperadmin(ctx context.Context, email string) (user.BootstrapResult, error)
}

// Config selects which tasks run.
type Config struct {
	// SuperadminEmail, when set, makes sure the platform has a superadmin.
	SuperadminEmail string
}

// Run executes the bootstrap tasks.
func Run(ctx context.Context, log *slog.Logger, cfg Config, users SuperadminEnsurer) error {
	if cfg.SuperadminEmail == "" {
		return nil
	}
	result, err := users.EnsureSuperadmin(ctx, cfg.SuperadminEmail)
	if err != nil {
		return fmt.Errorf("bootstrap superadmin: %w", err)
	}

	attrs := []any{slog.String("email", cfg.SuperadminEmail), slog.String("result", string(result))}
	switch result {
	case user.BootstrapCreated:
		log.InfoContext(ctx, "superadmin pre-provisioned: it is linked on the first login with a verified email", attrs...)
	case user.BootstrapPromoted:
		log.InfoContext(ctx, "existing user promoted to superadmin", attrs...)
	case user.BootstrapUnverified:
		log.WarnContext(ctx, "superadmin NOT granted: an account with that email exists but has not verified it", attrs...)
	default:
		log.DebugContext(ctx, "a superadmin already exists", attrs...)
	}
	return nil
}
