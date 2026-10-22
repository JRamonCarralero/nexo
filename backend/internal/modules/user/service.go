package user

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

type Service struct {
	repo  Repository
	newID func() (uuid.UUID, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, newID: uuid.NewV7}
}

// Resolve returns the Nexo user for a verified identity, creating it on the
// first login. Two simultaneous first requests may race to create the user:
// the loser gets a conflict and simply resolves again.
func (s *Service) Resolve(ctx context.Context, identity auth.Identity) (uuid.UUID, error) {
	var lastErr error
	for range 2 {
		newID, err := s.newID()
		if err != nil {
			return uuid.Nil, apperror.Internal(err, "generating user id")
		}
		id, err := s.repo.Resolve(ctx, newID, identity)
		if err == nil {
			return id, nil
		}
		var appErr *apperror.Error
		if !errors.As(err, &appErr) || appErr.Kind != apperror.KindConflict {
			return uuid.Nil, err
		}
		lastErr = err
	}
	return uuid.Nil, lastErr
}

// Me returns the authenticated user.
func (s *Service) Me(ctx context.Context) (User, error) {
	id, ok := scope.UserID(ctx)
	if !ok {
		return User{}, apperror.Unauthenticated("no authenticated user")
	}
	return s.repo.GetByID(ctx, id)
}

// EnsureSuperadmin makes sure the platform has a superadmin with that email.
func (s *Service) EnsureSuperadmin(ctx context.Context, email string) (BootstrapResult, error) {
	email = strings.TrimSpace(email)
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		return "", apperror.Validation("INVALID_EMAIL", "invalid superadmin email: "+email)
	}
	newID, err := s.newID()
	if err != nil {
		return "", apperror.Internal(err, "generating user id")
	}
	return s.repo.EnsureSuperadmin(ctx, newID, email)
}
