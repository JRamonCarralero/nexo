//go:build integration

package workspace_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres/pgtest"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

var testDB *pgtest.Database

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	testDB, err = pgtest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	testDB.Stop(ctx)
	os.Exit(code)
}

// world has three people: ana creates a workspace, luis joins it as a member
// and eva is an outsider with a workspace of her own.
type world struct {
	tm    *postgres.TxManager
	repo  *workspace.PostgresRepository
	users *user.Service
	svc   *workspace.Service

	ana, luis, eva uuid.UUID
	ws, evasWS     uuid.UUID
}

func newWorld(t *testing.T) world {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.OwnerPool.Exec(ctx, `TRUNCATE memberships, workspaces, users CASCADE`)
	require.NoError(t, err)

	tm := postgres.NewTxManager(testDB.Pool)
	w := world{tm: tm, repo: workspace.NewPostgresRepository(tm), users: user.NewService(user.NewPostgresRepository(tm))}
	w.svc = workspace.NewService(w.repo, directory{w.users})

	w.ana = w.signUp(t, "ana")
	w.luis = w.signUp(t, "luis")
	w.eva = w.signUp(t, "eva")

	created, err := w.svc.Create(asUserID(w.ana), "Product team")
	require.NoError(t, err)
	w.ws = created.ID
	evas, err := w.svc.Create(asUserID(w.eva), "Eva's")
	require.NoError(t, err)
	w.evasWS = evas.ID

	// Invitations arrive in the next PR: until then, memberships are added as the owner.
	_, err = testDB.OwnerPool.Exec(ctx,
		`INSERT INTO memberships (id, workspace_id, user_id, role, created_at, updated_at)
		 VALUES ($1, $2, $3, 'member', now(), now())`, uuid.Must(uuid.NewV7()), w.ws, w.luis)
	require.NoError(t, err)
	return w
}

func (w world) signUp(t *testing.T, name string) uuid.UUID {
	t.Helper()
	id, err := w.users.Resolve(context.Background(), auth.Identity{Subject: "uid-" + name, Email: name + "@example.com", Name: name})
	require.NoError(t, err)
	return id
}

// in returns the context the workspace middleware would build for user in ws.
func (w world) in(t *testing.T, userID, ws uuid.UUID) context.Context {
	t.Helper()
	ctx := asUserID(userID)
	role, err := w.svc.RoleOf(ctx, ws)
	if err != nil {
		role = "" // not a member: the middleware would answer 404, tests go further on purpose
	}
	return scope.WithWorkspaceRole(scope.WithWorkspace(ctx, ws), role)
}

func asUserID(id uuid.UUID) context.Context { return scope.WithUser(context.Background(), id) }

// directory mirrors the adapter in internal/app.
type directory struct{ users *user.Service }

func (d directory) Profiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]workspace.UserProfile, error) {
	list, err := d.users.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]workspace.UserProfile{}
	for _, u := range list {
		out[u.ID] = workspace.UserProfile{ID: u.ID, Email: u.Email, Name: u.Name}
	}
	return out, nil
}

func TestCreatorIsTheFirstAdmin(t *testing.T) {
	w := newWorld(t)

	mine, err := w.svc.ListMine(asUserID(w.ana))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, w.ws, mine[0].ID)
	assert.Equal(t, workspace.RoleAdmin, mine[0].Role)
	require.NotNil(t, mine[0].CreatedBy)
	assert.Equal(t, w.ana, *mine[0].CreatedBy)
}

func TestMembersSeeTheWorkspaceWithTheirRole(t *testing.T) {
	w := newWorld(t)

	mine, err := w.svc.ListMine(asUserID(w.luis))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, workspace.RoleMember, mine[0].Role)

	members, err := w.svc.Members(w.in(t, w.luis, w.ws))
	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "ana@example.com", members[0].Profile.Email, "co-members see each other's profile")
}

func TestOutsidersSeeNothing(t *testing.T) {
	w := newWorld(t)
	outsider := w.in(t, w.eva, w.ws)

	_, err := w.svc.RoleOf(asUserID(w.eva), w.ws)
	assert.Equal(t, "WORKSPACE_NOT_FOUND", code(t, err))

	_, err = w.repo.GetByID(outsider, w.ws)
	assert.Equal(t, "WORKSPACE_NOT_FOUND", code(t, err))

	memberships, err := w.repo.ListMemberships(outsider)
	require.NoError(t, err)
	assert.Empty(t, memberships)

	_, err = w.repo.Rename(outsider, w.ws, "Hacked")
	assert.Equal(t, "WORKSPACE_NOT_FOUND", code(t, err))

	profiles, err := w.users.GetByIDs(asUserID(w.eva), []uuid.UUID{w.ana, w.luis})
	require.NoError(t, err)
	assert.Empty(t, profiles, "people who share no workspace cannot see each other")

	mine, err := w.svc.ListMine(asUserID(w.eva))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, w.evasWS, mine[0].ID)
}

// The database enforces the admin rules on its own, even if the service
// checks were bypassed.
func TestMembersCannotAdministerAtTheDatabaseLevel(t *testing.T) {
	w := newWorld(t)
	asLuis := w.in(t, w.luis, w.ws)

	_, err := w.repo.Rename(asLuis, w.ws, "Hacked")
	assert.Equal(t, "WORKSPACE_NOT_FOUND", code(t, err), "the update policy hides the row from non-admins")

	anaMembership, err := w.repo.GetMembership(asLuis, w.ana)
	require.NoError(t, err)
	_, err = w.repo.UpdateRole(asLuis, anaMembership.ID, workspace.RoleGuest)
	assert.Equal(t, "MEMBER_NOT_FOUND", code(t, err))

	assert.Equal(t, "MEMBER_NOT_FOUND", code(t, w.repo.RemoveMembership(asLuis, anaMembership.ID)))
}

func TestAdminsChangeRolesButOneAdminAlwaysRemains(t *testing.T) {
	w := newWorld(t)
	asAna := w.in(t, w.ana, w.ws)

	_, err := w.svc.ChangeRole(asAna, w.ana, workspace.RoleMember)
	assert.Equal(t, workspace.CodeLastAdmin, code(t, err), "the only admin cannot step down")

	promoted, err := w.svc.ChangeRole(asAna, w.luis, workspace.RoleAdmin)
	require.NoError(t, err)
	assert.Equal(t, workspace.RoleAdmin, promoted.Role)

	_, err = w.svc.ChangeRole(asAna, w.ana, workspace.RoleMember)
	require.NoError(t, err, "with another admin, ana can step down")

	err = w.svc.RemoveMember(w.in(t, w.luis, w.ws), w.luis)
	assert.Equal(t, workspace.CodeLastAdmin, code(t, err), "luis is now the only admin and cannot leave")
}

func TestMembersCanLeave(t *testing.T) {
	w := newWorld(t)

	require.NoError(t, w.svc.RemoveMember(w.in(t, w.luis, w.ws), w.luis))

	mine, err := w.svc.ListMine(asUserID(w.luis))
	require.NoError(t, err)
	assert.Empty(t, mine)
}

func TestTheAPIRoleCannotAddMembersOrChangeProtectedColumns(t *testing.T) {
	w := newWorld(t)
	asAna := w.in(t, w.ana, w.ws)

	tests := map[string]string{
		"insert membership":  `INSERT INTO memberships (id, workspace_id, user_id, role, created_at, updated_at) VALUES (gen_random_uuid(), '` + w.ws.String() + `', '` + w.eva.String() + `', 'admin', now(), now())`,
		"insert workspace":   `INSERT INTO workspaces (id, name, created_at, updated_at) VALUES (gen_random_uuid(), 'x', now(), now())`,
		"move a membership":  `UPDATE memberships SET user_id = '` + w.eva.String() + `' WHERE user_id = '` + w.luis.String() + `'`,
		"rewrite created_by": `UPDATE workspaces SET created_by = '` + w.eva.String() + `'`,
	}
	for name, sql := range tests {
		t.Run(name, func(t *testing.T) {
			err := w.tm.WithinTx(asAna, func(ctx context.Context) error {
				_, err := w.tm.Conn(ctx).Exec(ctx, sql)
				return postgres.TranslateError(err)
			})
			assert.Equal(t, apperror.CodeForbidden, code(t, err))
		})
	}
}

func TestDeletedWorkspacesDisappear(t *testing.T) {
	w := newWorld(t)
	_, err := testDB.OwnerPool.Exec(context.Background(), `UPDATE workspaces SET deleted_at = now() WHERE id = $1`, w.ws)
	require.NoError(t, err)

	mine, err := w.svc.ListMine(asUserID(w.ana))
	require.NoError(t, err)
	assert.Empty(t, mine)

	_, err = w.svc.RoleOf(asUserID(w.ana), w.ws)
	assert.Equal(t, "WORKSPACE_NOT_FOUND", code(t, err))

	profiles, err := w.users.GetByIDs(asUserID(w.ana), []uuid.UUID{w.luis})
	require.NoError(t, err)
	assert.Empty(t, profiles, "a deleted workspace no longer makes people visible to each other")
}
