package scope_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

func TestScope(t *testing.T) {
	ctx := context.Background()

	_, ok := scope.UserID(ctx)
	assert.False(t, ok)
	_, ok = scope.WorkspaceID(ctx)
	assert.False(t, ok)

	user, ws := uuid.New(), uuid.New()
	ctx = scope.WithWorkspace(scope.WithUser(ctx, user), ws)

	got, ok := scope.UserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, user, got)
	got, ok = scope.WorkspaceID(ctx)
	assert.True(t, ok)
	assert.Equal(t, ws, got)
}

func TestNilIDsAreAbsent(t *testing.T) {
	ctx := scope.WithWorkspace(scope.WithUser(context.Background(), uuid.Nil), uuid.Nil)

	_, ok := scope.UserID(ctx)
	assert.False(t, ok)
	_, ok = scope.WorkspaceID(ctx)
	assert.False(t, ok)
}
