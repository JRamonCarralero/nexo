// Package workspace manages workspaces and their memberships. It owns the
// workspaces and memberships tables.
package workspace

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
)

// Role is a member's role in a workspace.
type Role string

const (
	RoleAdmin  Role = "admin"  // manages the workspace, its members and invitations
	RoleMember Role = "member" // works in the workspace
	RoleGuest  Role = "guest"  // external collaborator (restrictions arrive with projects)
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleMember, RoleGuest:
		return true
	}
	return false
}

// MaxNameLength is the maximum workspace name length, in characters.
const MaxNameLength = 100

// Workspace groups people and their projects. Its id is the tenant key of
// every workspace-scoped table.
type Workspace struct {
	database.Model
	Name      string     `db:"name"`
	CreatedBy *uuid.UUID `db:"created_by"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (Workspace) TableName() string    { return "workspaces" }
func (Workspace) ResourceName() string { return "workspace" }

// WithRole is a workspace together with the acting user's role in it.
type WithRole struct {
	Workspace
	Role Role `db:"role"`
}

// Membership says that a user belongs to a workspace with a role.
type Membership struct {
	database.TenantModel
	UserID    uuid.UUID  `db:"user_id"`
	Role      Role       `db:"role"`
	CreatedBy *uuid.UUID `db:"created_by"`
}

func (Membership) TableName() string    { return "memberships" }
func (Membership) ResourceName() string { return "member" }

// Member is a membership with the person's public profile.
type Member struct {
	Membership
	Profile UserProfile
}

// Repository is the persistence the workspace service needs.
type Repository interface {
	// Create inserts the workspace and makes the acting user its first admin.
	Create(ctx context.Context, w Workspace, membershipID uuid.UUID) error
	// ListForUser returns the non-deleted workspaces the user belongs to, by name.
	ListForUser(ctx context.Context, userID uuid.UUID) ([]WithRole, error)
	// RoleOf returns the user's role in the workspace, or WORKSPACE_NOT_FOUND.
	RoleOf(ctx context.Context, workspaceID, userID uuid.UUID) (Role, error)
	// GetByID returns a workspace visible to the acting user.
	GetByID(ctx context.Context, id uuid.UUID) (Workspace, error)
	// Rename changes the name of the current workspace.
	Rename(ctx context.Context, id uuid.UUID, name string) (Workspace, error)
	// ListMemberships returns every membership of the current workspace, oldest first.
	ListMemberships(ctx context.Context) ([]Membership, error)
	// GetMembership returns the membership of a user in the current workspace, or MEMBER_NOT_FOUND.
	GetMembership(ctx context.Context, userID uuid.UUID) (Membership, error)
	// UpdateRole changes a membership's role. Removing the last admin returns LAST_ADMIN.
	UpdateRole(ctx context.Context, membershipID uuid.UUID, role Role) (Membership, error)
	// RemoveMembership deletes a membership. Removing the last admin returns LAST_ADMIN.
	RemoveMembership(ctx context.Context, membershipID uuid.UUID) error
}
