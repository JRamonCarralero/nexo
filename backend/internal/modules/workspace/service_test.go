package workspace_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace/mocks"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

var (
	me    = uuid.MustParse("0192f7a4-0000-7000-8000-000000000001")
	other = uuid.MustParse("0192f7a4-0000-7000-8000-000000000002")
	wsID  = uuid.MustParse("0192f7a4-0000-7000-8000-0000000000aa")
)

type env struct {
	repo  *mocks.Repository
	users *mocks.UserDirectory
	svc   *workspace.Service
}

func newEnv(t *testing.T) env {
	repo, users := mocks.NewRepository(t), mocks.NewUserDirectory(t)
	return env{repo: repo, users: users, svc: workspace.NewService(repo, users)}
}

func asUser() context.Context { return scope.WithUser(context.Background(), me) }

func inWorkspace(role workspace.Role) context.Context {
	return scope.WithWorkspaceRole(scope.WithWorkspace(asUser(), wsID), string(role))
}

func code(t *testing.T, err error) string {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func membership(user uuid.UUID, role workspace.Role) workspace.Membership {
	m := workspace.Membership{UserID: user, Role: role}
	m.ID = uuid.New()
	m.WorkspaceID = wsID
	return m
}

func TestCreateValidatesTheName(t *testing.T) {
	for name, input := range map[string]string{
		"empty":    "   ",
		"too long": strings.Repeat("a", workspace.MaxNameLength+1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newEnv(t).svc.Create(asUser(), input)
			assert.Equal(t, workspace.CodeInvalidName, code(t, err))
		})
	}
}

func TestCreateMakesTheCreatorAdmin(t *testing.T) {
	e := newEnv(t)
	var created uuid.UUID
	e.repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(w workspace.Workspace) bool {
		created = w.ID
		return w.Name == "Product team" && w.ID.Version() == 7
	}), mock.MatchedBy(func(id uuid.UUID) bool { return id.Version() == 7 })).Return(nil).Once()
	e.repo.EXPECT().GetByID(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, id uuid.UUID) (workspace.Workspace, error) {
			w := workspace.Workspace{Name: "Product team"}
			w.ID = id
			return w, nil
		}).Once()

	got, err := e.svc.Create(asUser(), "  Product team  ")

	require.NoError(t, err)
	assert.Equal(t, created, got.ID)
	assert.Equal(t, workspace.RoleAdmin, got.Role)
	assert.Equal(t, "Product team", got.Name, "the name is trimmed")
}

func TestRenameIsForAdminsOnly(t *testing.T) {
	_, err := newEnv(t).svc.Rename(inWorkspace(workspace.RoleMember), "Nuevo")
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err))
}

func TestRenameByAdmin(t *testing.T) {
	e := newEnv(t)
	renamed := workspace.Workspace{Name: "Nuevo"}
	renamed.ID = wsID
	e.repo.EXPECT().Rename(mock.Anything, wsID, "Nuevo").Return(renamed, nil).Once()

	got, err := e.svc.Rename(inWorkspace(workspace.RoleAdmin), "Nuevo")

	require.NoError(t, err)
	assert.Equal(t, "Nuevo", got.Name)
}

func TestMembersCombineMembershipsAndProfiles(t *testing.T) {
	e := newEnv(t)
	mine, theirs := membership(me, workspace.RoleAdmin), membership(other, workspace.RoleMember)
	e.repo.EXPECT().ListMemberships(mock.Anything).Return([]workspace.Membership{mine, theirs}, nil).Once()
	e.users.EXPECT().Profiles(mock.Anything, []uuid.UUID{me, other}).Return(map[uuid.UUID]workspace.UserProfile{
		me:    {ID: me, Name: "Ana"},
		other: {ID: other, Name: "Luis"},
	}, nil).Once()

	members, err := e.svc.Members(inWorkspace(workspace.RoleMember))

	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "Ana", members[0].Profile.Name)
	assert.Equal(t, workspace.RoleMember, members[1].Role)
}

func TestChangeRoleIsForAdminsOnly(t *testing.T) {
	_, err := newEnv(t).svc.ChangeRole(inWorkspace(workspace.RoleMember), other, workspace.RoleAdmin)
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err))
}

func TestChangeRoleRejectsUnknownRoles(t *testing.T) {
	_, err := newEnv(t).svc.ChangeRole(inWorkspace(workspace.RoleAdmin), other, "owner")
	assert.Equal(t, workspace.CodeInvalidRole, code(t, err))
}

func TestChangeRoleToTheSameRoleDoesNotWrite(t *testing.T) {
	e := newEnv(t)
	m := membership(other, workspace.RoleMember)
	e.repo.EXPECT().GetMembership(mock.Anything, other).Return(m, nil).Once()
	e.users.EXPECT().Profiles(mock.Anything, []uuid.UUID{other}).Return(nil, nil).Once()

	got, err := e.svc.ChangeRole(inWorkspace(workspace.RoleAdmin), other, workspace.RoleMember)

	require.NoError(t, err)
	assert.Equal(t, workspace.RoleMember, got.Role)
}

func TestChangeRolePropagatesLastAdmin(t *testing.T) {
	e := newEnv(t)
	m := membership(me, workspace.RoleAdmin)
	e.repo.EXPECT().GetMembership(mock.Anything, me).Return(m, nil).Once()
	e.repo.EXPECT().UpdateRole(mock.Anything, m.ID, workspace.RoleMember).
		Return(workspace.Membership{}, apperror.Conflict(workspace.CodeLastAdmin, "last admin")).Once()

	_, err := e.svc.ChangeRole(inWorkspace(workspace.RoleAdmin), me, workspace.RoleMember)

	assert.Equal(t, workspace.CodeLastAdmin, code(t, err))
}

func TestAnyoneCanLeave(t *testing.T) {
	e := newEnv(t)
	m := membership(me, workspace.RoleGuest)
	e.repo.EXPECT().GetMembership(mock.Anything, me).Return(m, nil).Once()
	e.repo.EXPECT().RemoveMembership(mock.Anything, m.ID).Return(nil).Once()

	require.NoError(t, e.svc.RemoveMember(inWorkspace(workspace.RoleGuest), me))
}

func TestOnlyAdminsRemoveOthers(t *testing.T) {
	err := newEnv(t).svc.RemoveMember(inWorkspace(workspace.RoleMember), other)
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err))
}

func TestWorkspaceOperationsNeedAWorkspaceInScope(t *testing.T) {
	_, err := newEnv(t).svc.Members(asUser())
	assert.Equal(t, apperror.CodeInternal, code(t, err))
}
