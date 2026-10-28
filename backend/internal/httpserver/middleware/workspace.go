package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// WorkspaceParam is the path parameter that identifies the workspace.
const WorkspaceParam = "workspaceId"

// WorkspaceResolver returns the acting user's role in a workspace, or a
// WORKSPACE_NOT_FOUND error when they are not a member.
type WorkspaceResolver interface {
	RoleOf(ctx context.Context, workspaceID uuid.UUID) (string, error)
}

// Workspace runs after Auth on routes with a {workspaceId} path parameter: it
// checks that the user belongs to the workspace and puts the workspace and
// the user's role in the request scope, from where they reach Row-Level
// Security. Non-members get 404, so the existence of a workspace is not revealed.
func Workspace(workspaces WorkspaceResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Param(WorkspaceParam)
		if raw == "" {
			c.Next()
			return
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			abort(c, apperror.Validation("INVALID_WORKSPACE_ID", "invalid workspace id",
				apperror.FieldError{Field: WorkspaceParam, Code: "INVALID_UUID"}))
			return
		}
		role, err := workspaces.RoleOf(c.Request.Context(), id)
		if err != nil {
			abort(c, err)
			return
		}

		ctx := scope.WithWorkspaceRole(scope.WithWorkspace(c.Request.Context(), id), role)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
