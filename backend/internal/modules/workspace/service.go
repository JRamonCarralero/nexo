package workspace

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// Error codes specific to workspaces.
const (
	CodeInvalidName   = "INVALID_WORKSPACE_NAME"
	CodeInvalidRole   = "INVALID_ROLE"
	CodeAdminRequired = "WORKSPACE_ADMIN_REQUIRED"
	CodeLastAdmin     = "LAST_ADMIN"
)

type Service struct {
	repo  Repository
	users UserDirectory
	newID func() (uuid.UUID, error)
}

func NewService(repo Repository, users UserDirectory) *Service {
	return &Service{repo: repo, users: users, newID: uuid.NewV7}
}

// Create makes a new workspace with the acting user as its first admin.
func (s *Service) Create(ctx context.Context, name string) (WithRole, error) {
	name, err := validName(name)
	if err != nil {
		return WithRole{}, err
	}
	if _, ok := scope.UserID(ctx); !ok {
		return WithRole{}, apperror.Unauthenticated("no authenticated user")
	}
	wsID, err := s.newID()
	if err != nil {
		return WithRole{}, apperror.Internal(err, "generating workspace id")
	}
	membershipID, err := s.newID()
	if err != nil {
		return WithRole{}, apperror.Internal(err, "generating membership id")
	}

	w := Workspace{Name: name}
	w.ID = wsID
	if err := s.repo.Create(ctx, w, membershipID); err != nil {
		return WithRole{}, err
	}
	created, err := s.repo.GetByID(ctx, wsID)
	if err != nil {
		return WithRole{}, err
	}
	return WithRole{Workspace: created, Role: RoleAdmin}, nil
}

// ListMine returns the workspaces of the acting user.
func (s *Service) ListMine(ctx context.Context) ([]WithRole, error) {
	userID, ok := scope.UserID(ctx)
	if !ok {
		return nil, apperror.Unauthenticated("no authenticated user")
	}
	return s.repo.ListForUser(ctx, userID)
}

// RoleOf returns the acting user's role in a workspace. Used by the workspace
// middleware; non-members get WORKSPACE_NOT_FOUND so existence is not revealed.
func (s *Service) RoleOf(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	userID, ok := scope.UserID(ctx)
	if !ok {
		return "", apperror.Unauthenticated("no authenticated user")
	}
	role, err := s.repo.RoleOf(ctx, workspaceID, userID)
	return string(role), err
}

// Current returns the workspace of the request scope with the acting user's role.
func (s *Service) Current(ctx context.Context) (WithRole, error) {
	id, err := currentWorkspace(ctx)
	if err != nil {
		return WithRole{}, err
	}
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return WithRole{}, err
	}
	return WithRole{Workspace: w, Role: Role(scope.WorkspaceRole(ctx))}, nil
}

// Rename changes the name of the current workspace. Admins only.
func (s *Service) Rename(ctx context.Context, name string) (WithRole, error) {
	id, err := currentWorkspace(ctx)
	if err != nil {
		return WithRole{}, err
	}
	if err := requireAdmin(ctx); err != nil {
		return WithRole{}, err
	}
	name, err = validName(name)
	if err != nil {
		return WithRole{}, err
	}
	w, err := s.repo.Rename(ctx, id, name)
	if err != nil {
		return WithRole{}, err
	}
	return WithRole{Workspace: w, Role: RoleAdmin}, nil
}

// Members returns the members of the current workspace with their profiles.
func (s *Service) Members(ctx context.Context) ([]Member, error) {
	if _, err := currentWorkspace(ctx); err != nil {
		return nil, err
	}
	memberships, err := s.repo.ListMemberships(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(memberships))
	for i, m := range memberships {
		ids[i] = m.UserID
	}
	profiles, err := s.users.Profiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	members := make([]Member, len(memberships))
	for i, m := range memberships {
		members[i] = Member{Membership: m, Profile: profiles[m.UserID]}
	}
	return members, nil
}

// ChangeRole changes a member's role. Admins only; a workspace keeps at least one admin.
func (s *Service) ChangeRole(ctx context.Context, userID uuid.UUID, role Role) (Member, error) {
	if _, err := currentWorkspace(ctx); err != nil {
		return Member{}, err
	}
	if err := requireAdmin(ctx); err != nil {
		return Member{}, err
	}
	if !role.Valid() {
		return Member{}, apperror.Validation(CodeInvalidRole, "unknown role: "+string(role),
			apperror.FieldError{Field: "role", Code: CodeInvalidRole})
	}
	m, err := s.repo.GetMembership(ctx, userID)
	if err != nil {
		return Member{}, err
	}
	if m.Role != role {
		if m, err = s.repo.UpdateRole(ctx, m.ID, role); err != nil {
			return Member{}, err
		}
	}
	return s.withProfile(ctx, m)
}

// RemoveMember removes someone from the current workspace. Admins can remove
// anyone; anyone can remove themselves (leave). A workspace keeps at least one admin.
func (s *Service) RemoveMember(ctx context.Context, userID uuid.UUID) error {
	if _, err := currentWorkspace(ctx); err != nil {
		return err
	}
	me, _ := scope.UserID(ctx)
	if userID != me {
		if err := requireAdmin(ctx); err != nil {
			return err
		}
	}
	m, err := s.repo.GetMembership(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.RemoveMembership(ctx, m.ID)
}

func (s *Service) withProfile(ctx context.Context, m Membership) (Member, error) {
	profiles, err := s.users.Profiles(ctx, []uuid.UUID{m.UserID})
	if err != nil {
		return Member{}, err
	}
	return Member{Membership: m, Profile: profiles[m.UserID]}, nil
}

func currentWorkspace(ctx context.Context) (uuid.UUID, error) {
	id, ok := scope.WorkspaceID(ctx)
	if !ok {
		return uuid.Nil, apperror.Internal(nil, "workspace operation without a workspace in scope")
	}
	return id, nil
}

func requireAdmin(ctx context.Context) error {
	if Role(scope.WorkspaceRole(ctx)) != RoleAdmin {
		return apperror.Forbidden(CodeAdminRequired, "only workspace admins can do this")
	}
	return nil
}

func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	switch n := utf8.RuneCountInString(name); {
	case n == 0:
		return "", apperror.Validation(CodeInvalidName, "workspace name is required",
			apperror.FieldError{Field: "name", Code: "REQUIRED"})
	case n > MaxNameLength:
		return "", apperror.Validation(CodeInvalidName, "workspace name is too long",
			apperror.FieldError{Field: "name", Code: "TOO_LONG"})
	}
	return name, nil
}
