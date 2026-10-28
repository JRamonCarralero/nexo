// Package scope carries who is acting and in which workspace through the
// request context. The auth middleware sets the user, the workspace middleware
// sets the workspace, and the database layer reads both to apply Row-Level
// Security and fill audit columns.
package scope

import (
	"context"

	"github.com/google/uuid"
)

type (
	userKey      struct{}
	workspaceKey struct{}
	roleKey      struct{}
)

// WithUser returns a context acting on behalf of the given user.
func WithUser(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey{}, userID)
}

// UserID returns the acting user, if any.
func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// WithWorkspace returns a context scoped to the given workspace.
func WithWorkspace(ctx context.Context, workspaceID uuid.UUID) context.Context {
	return context.WithValue(ctx, workspaceKey{}, workspaceID)
}

// WorkspaceID returns the current workspace, if any.
func WorkspaceID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(workspaceKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// WithWorkspaceRole records the acting user's role in the current workspace
// (set by the workspace middleware after checking the membership).
func WithWorkspaceRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey{}, role)
}

// WorkspaceRole returns the acting user's role in the current workspace, or "".
func WorkspaceRole(ctx context.Context) string {
	role, _ := ctx.Value(roleKey{}).(string)
	return role
}
