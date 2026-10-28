// Package user manages Nexo users: mapping identities from the identity
// provider to users, the authenticated user's profile and the platform
// superadmin.
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
)

// User is a person using Nexo. Users are global: they belong to workspaces
// through memberships.
type User struct {
	database.Model
	AuthSubject   *string `db:"auth_subject"` // nil until the first login
	Email         string  `db:"email"`
	EmailVerified bool    `db:"email_verified"`
	Name          string  `db:"name"`
	AvatarURL     string  `db:"avatar_url"`
	IsSuperadmin  bool    `db:"is_superadmin"`
}

func (User) TableName() string    { return "users" }
func (User) ResourceName() string { return "user" }

// BootstrapResult says what EnsureSuperadmin did.
type BootstrapResult string

const (
	BootstrapExists     BootstrapResult = "exists"
	BootstrapPromoted   BootstrapResult = "promoted"
	BootstrapCreated    BootstrapResult = "created"
	BootstrapUnverified BootstrapResult = "unverified"
)

// Repository is the persistence the user service needs.
type Repository interface {
	// Resolve returns the user for a verified identity, creating it on first login.
	Resolve(ctx context.Context, newID uuid.UUID, id auth.Identity) (uuid.UUID, error)
	// GetByID returns a user visible to the acting user.
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	// GetByIDs returns the users visible to the acting user among ids (others are skipped).
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error)
	// EnsureSuperadmin makes sure the platform has a superadmin.
	EnsureSuperadmin(ctx context.Context, newID uuid.UUID, email string) (BootstrapResult, error)
}
