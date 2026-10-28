package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/httpserver"
	"github.com/JRamonCarralero/nexo/backend/internal/httpserver/api"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/system"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/logger"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

const validToken = "valid-token"

var testUserID = uuid.MustParse("0192f7a4-0000-7000-8000-000000000001")

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

// fakeAuthn accepts only validToken.
type fakeAuthn struct{}

func (fakeAuthn) VerifyToken(_ context.Context, token string) (auth.Identity, error) {
	if token != validToken {
		return auth.Identity{}, apperror.New(apperror.KindUnauthenticated, "INVALID_TOKEN", "invalid token")
	}
	return auth.Identity{Subject: "uid-1", Email: "ana@example.com"}, nil
}

// fakeUsers resolves every identity to testUserID and returns it as the profile
// only when the request scope carries it.
type fakeUsers struct{}

func (fakeUsers) Resolve(context.Context, uuid.UUID, auth.Identity) (uuid.UUID, error) {
	return testUserID, nil
}

func (fakeUsers) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	if scoped, ok := scope.UserID(ctx); !ok || scoped != id {
		return user.User{}, apperror.NotFound("user", id)
	}
	u := user.User{Email: "ana@example.com", Name: "Ana"}
	u.ID = id
	return u, nil
}

func (fakeUsers) GetByIDs(context.Context, []uuid.UUID) ([]user.User, error) { return nil, nil }

func (fakeUsers) EnsureSuperadmin(context.Context, uuid.UUID, string) (user.BootstrapResult, error) {
	return user.BootstrapExists, nil
}

// testWorkspaceID is the only workspace testUserID belongs to (as a member).
var testWorkspaceID = uuid.MustParse("0192f7a4-0000-7000-8000-0000000000aa")

// fakeWorkspaces is a workspace repository with a single workspace.
type fakeWorkspaces struct{ workspace.Repository }

func (fakeWorkspaces) RoleOf(_ context.Context, ws, u uuid.UUID) (workspace.Role, error) {
	if ws == testWorkspaceID && u == testUserID {
		return workspace.RoleMember, nil
	}
	return "", apperror.NotFound("workspace", ws)
}

func (fakeWorkspaces) GetByID(ctx context.Context, id uuid.UUID) (workspace.Workspace, error) {
	if scoped, ok := scope.WorkspaceID(ctx); !ok || scoped != id {
		return workspace.Workspace{}, apperror.NotFound("workspace", id)
	}
	w := workspace.Workspace{Name: "Product team"}
	w.ID = id
	return w, nil
}

type noDirectory struct{}

func (noDirectory) Profiles(context.Context, []uuid.UUID) (map[uuid.UUID]workspace.UserProfile, error) {
	return nil, nil
}

func newRouter(t *testing.T, pingErr error) (*gin.Engine, *bytes.Buffer) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	users := user.NewService(fakeUsers{})
	workspaces := workspace.NewService(fakeWorkspaces{}, noDirectory{})
	r, err := httpserver.NewRouter(
		httpserver.Deps{
			Logger:        logger.New(&logs, "test", slog.LevelDebug),
			Authenticator: fakeAuthn{},
			Users:         users,
			Workspaces:    workspaces,
		},
		httpserver.Handlers{
			System:    system.NewHandler(fakePinger{err: pingErr}, "test"),
			User:      user.NewHandler(users),
			Workspace: workspace.NewHandler(workspaces),
		},
	)
	require.NoError(t, err)
	return r, &logs
}

var withToken = map[string]string{"Authorization": "Bearer " + validToken}

func do(r http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeProblem(t *testing.T, body io.Reader) api.Problem {
	t.Helper()
	var p api.Problem
	require.NoError(t, json.NewDecoder(body).Decode(&p))
	return p
}

func TestHealthOK(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/health", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var h api.Health
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&h))
	assert.Equal(t, api.HealthStatusOk, h.Status)
	assert.Equal(t, api.HealthChecksOk, h.Checks["database"])
	assert.Len(t, rec.Header().Get("X-Trace-Id"), 32)
}

func TestHealthDatabaseDownReturnsProblem(t *testing.T) {
	r, logs := newRouter(t, errors.New("connection refused"))

	rec := do(r, http.MethodGet, "/api/v1/health", nil)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	p := decodeProblem(t, rec.Body)
	assert.Equal(t, "SERVICE_UNAVAILABLE", p.Code)
	assert.Equal(t, rec.Header().Get("X-Trace-Id"), p.TraceId)
	assert.NotContains(t, *p.Detail, "connection refused", "internal causes must not reach clients")

	assert.Contains(t, logs.String(), `"severity":"ERROR"`)
	assert.Contains(t, logs.String(), "connection refused", "the cause must be in the logs")
}

func TestUnknownRouteReturnsProblem(t *testing.T) {
	r, logs := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/nope", nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "ROUTE_NOT_FOUND", decodeProblem(t, rec.Body).Code)
	assert.Contains(t, logs.String(), `"severity":"WARNING"`)
}

func TestPanicIsRecoveredAsProblem(t *testing.T) {
	r, logs := newRouter(t, nil)
	r.GET("/boom", func(*gin.Context) { panic("kaboom") })

	rec := do(r, http.MethodGet, "/boom", withToken)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "INTERNAL_ERROR", decodeProblem(t, rec.Body).Code)
	assert.Contains(t, logs.String(), "kaboom")
}

func TestIncomingTraceparentIsReused(t *testing.T) {
	r, logs := newRouter(t, nil)
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"

	rec := do(r, http.MethodGet, "/api/v1/health", map[string]string{
		"traceparent": "00-" + traceID + "-00f067aa0ba902b7-01",
	})

	assert.Equal(t, traceID, rec.Header().Get("X-Trace-Id"))
	assert.Contains(t, logs.String(), traceID)
}

func TestMeReturnsTheAuthenticatedUser(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/me", withToken)

	require.Equal(t, http.StatusOK, rec.Code)
	var u api.User
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&u))
	assert.Equal(t, testUserID, u.Id)
	assert.Equal(t, "ana@example.com", string(u.Email))
}

func TestProtectedRoutesRequireAToken(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		wantCode string
	}{
		{"no header", nil, "MISSING_TOKEN"},
		{"wrong scheme", map[string]string{"Authorization": "Basic abc"}, "MISSING_TOKEN"},
		{"empty bearer", map[string]string{"Authorization": "Bearer "}, "MISSING_TOKEN"},
		{"invalid token", map[string]string{"Authorization": "Bearer nope"}, "INVALID_TOKEN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := newRouter(t, nil)

			rec := do(r, http.MethodGet, "/api/v1/me", tt.headers)

			require.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, tt.wantCode, decodeProblem(t, rec.Body).Code)
		})
	}
}

func TestRoutesOutsideTheContractAreProtectedByDefault(t *testing.T) {
	r, _ := newRouter(t, nil)
	r.GET("/internal/thing", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	assert.Equal(t, http.StatusUnauthorized, do(r, http.MethodGet, "/internal/thing", nil).Code)
	assert.Equal(t, http.StatusNoContent, do(r, http.MethodGet, "/internal/thing", withToken).Code)
}

func TestPublicRoutesNeedNoToken(t *testing.T) {
	r, _ := newRouter(t, nil)

	assert.Equal(t, http.StatusOK, do(r, http.MethodGet, "/api/v1/health", nil).Code)
}

func TestWorkspaceRoutesPutTheWorkspaceAndRoleInScope(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/workspaces/"+testWorkspaceID.String(), withToken)

	require.Equal(t, http.StatusOK, rec.Code)
	var w api.Workspace
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&w))
	assert.Equal(t, testWorkspaceID, w.Id)
	assert.Equal(t, api.RoleMember, w.MyRole)
}

func TestNonMembersGetNotFound(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/workspaces/"+uuid.NewString(), withToken)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "WORKSPACE_NOT_FOUND", decodeProblem(t, rec.Body).Code)
}

func TestMalformedWorkspaceIDIsRejected(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/workspaces/not-a-uuid", withToken)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "INVALID_WORKSPACE_ID", decodeProblem(t, rec.Body).Code)
}

func TestWorkspaceRoutesCheckAuthenticationFirst(t *testing.T) {
	r, _ := newRouter(t, nil)

	rec := do(r, http.MethodGet, "/api/v1/workspaces/"+testWorkspaceID.String(), nil)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMembersCannotRenameTheWorkspace(t *testing.T) {
	r, _ := newRouter(t, nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPatch,
		"/api/v1/workspaces/"+testWorkspaceID.String(), strings.NewReader(`{"name":"Nuevo"}`))
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "WORKSPACE_ADMIN_REQUIRED", decodeProblem(t, rec.Body).Code)
}
