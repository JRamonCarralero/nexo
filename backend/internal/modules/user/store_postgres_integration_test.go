//go:build integration

package user_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
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

type env struct {
	tm   *postgres.TxManager
	repo *user.PostgresRepository
	svc  *user.Service
}

func newEnv(t *testing.T) env {
	t.Helper()
	_, err := testDB.OwnerPool.Exec(context.Background(), `TRUNCATE users CASCADE`)
	require.NoError(t, err)
	tm := postgres.NewTxManager(testDB.Pool)
	repo := user.NewPostgresRepository(tm)
	return env{tm: tm, repo: repo, svc: user.NewService(repo)}
}

func asUser(id uuid.UUID) context.Context { return scope.WithUser(context.Background(), id) }

func (e env) me(t *testing.T, id uuid.UUID) user.User {
	t.Helper()
	u, err := e.svc.Me(asUser(id))
	require.NoError(t, err)
	return u
}

func newIdentity() auth.Identity {
	n := uuid.NewString()[:8]
	return auth.Identity{Subject: "uid-" + n, Email: n + "@example.com", Name: "User " + n}
}

func appCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func TestFirstLoginCreatesTheUser(t *testing.T) {
	e := newEnv(t)
	id := newIdentity()

	userID, err := e.svc.Resolve(context.Background(), id)
	require.NoError(t, err)

	u := e.me(t, userID)
	assert.Equal(t, uuid.Version(7), u.ID.Version())
	assert.Equal(t, id.Email, u.Email)
	assert.Equal(t, id.Name, u.Name)
	require.NotNil(t, u.AuthSubject)
	assert.Equal(t, id.Subject, *u.AuthSubject)
	assert.False(t, u.IsSuperadmin)
}

func TestNextLoginsReturnTheSameUserAndRefreshEmailData(t *testing.T) {
	e := newEnv(t)
	id := newIdentity()
	first, err := e.svc.Resolve(context.Background(), id)
	require.NoError(t, err)

	id.EmailVerified = true
	id.Name = "Changed in the provider"
	second, err := e.svc.Resolve(context.Background(), id)
	require.NoError(t, err)

	assert.Equal(t, first, second)
	u := e.me(t, second)
	assert.True(t, u.EmailVerified)
	assert.NotEqual(t, "Changed in the provider", u.Name, "a name already set in Nexo is kept")
}

func TestEmailIsCaseInsensitive(t *testing.T) {
	e := newEnv(t)
	id := newIdentity()
	_, err := e.svc.Resolve(context.Background(), id)
	require.NoError(t, err)

	other := newIdentity()
	other.Email = strings.ToUpper(id.Email)
	_, err = e.svc.Resolve(context.Background(), other)

	assert.Equal(t, "EMAIL_IN_USE", appCode(t, err))
}

func TestPreProvisionedUserIsLinkedOnlyWithVerifiedEmail(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	result, err := e.svc.EnsureSuperadmin(ctx, "boss@example.com")
	require.NoError(t, err)
	require.Equal(t, user.BootstrapCreated, result)

	unverified := auth.Identity{Subject: "uid-boss", Email: "boss@example.com"}
	_, err = e.svc.Resolve(ctx, unverified)
	assert.Equal(t, "EMAIL_IN_USE", appCode(t, err), "an unverified email must not take over the account")

	unverified.EmailVerified = true
	userID, err := e.svc.Resolve(ctx, unverified)
	require.NoError(t, err)

	u := e.me(t, userID)
	assert.True(t, u.IsSuperadmin)
	require.NotNil(t, u.AuthSubject)
	assert.Equal(t, "uid-boss", *u.AuthSubject)
}

func TestEnsureSuperadminOutcomes(t *testing.T) {
	ctx := context.Background()

	t.Run("promotes an existing verified user", func(t *testing.T) {
		e := newEnv(t)
		id := newIdentity()
		id.EmailVerified = true
		userID, err := e.svc.Resolve(ctx, id)
		require.NoError(t, err)

		result, err := e.svc.EnsureSuperadmin(ctx, id.Email)
		require.NoError(t, err)
		assert.Equal(t, user.BootstrapPromoted, result)
		assert.True(t, e.me(t, userID).IsSuperadmin)
	})

	t.Run("does not promote an unverified email", func(t *testing.T) {
		e := newEnv(t)
		id := newIdentity()
		userID, err := e.svc.Resolve(ctx, id)
		require.NoError(t, err)

		result, err := e.svc.EnsureSuperadmin(ctx, id.Email)
		require.NoError(t, err)
		assert.Equal(t, user.BootstrapUnverified, result)
		assert.False(t, e.me(t, userID).IsSuperadmin)
	})

	t.Run("does nothing when a superadmin exists", func(t *testing.T) {
		e := newEnv(t)
		_, err := e.svc.EnsureSuperadmin(ctx, "first@example.com")
		require.NoError(t, err)

		result, err := e.svc.EnsureSuperadmin(ctx, "second@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.BootstrapExists, result)

		var n int
		require.NoError(t, testDB.OwnerPool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n))
		assert.Equal(t, 1, n)
	})
}

func TestUsersCannotSeeEachOther(t *testing.T) {
	e := newEnv(t)
	a, err := e.svc.Resolve(context.Background(), newIdentity())
	require.NoError(t, err)
	b, err := e.svc.Resolve(context.Background(), newIdentity())
	require.NoError(t, err)

	_, err = e.repo.GetByID(asUser(a), b)
	assert.Equal(t, "USER_NOT_FOUND", appCode(t, err))

	_, err = e.repo.GetByID(context.Background(), a)
	assert.Equal(t, "USER_NOT_FOUND", appCode(t, err), "without a user in scope nothing is visible")
}

func TestUsersCannotMakeThemselvesSuperadmin(t *testing.T) {
	e := newEnv(t)
	id, err := e.svc.Resolve(context.Background(), newIdentity())
	require.NoError(t, err)

	err = e.tm.WithinTx(asUser(id), func(ctx context.Context) error {
		_, err := e.tm.Conn(ctx).Exec(ctx, `UPDATE users SET is_superadmin = true WHERE id = $1`, id)
		return postgres.TranslateError(err)
	})

	assert.Equal(t, apperror.CodeForbidden, appCode(t, err))
	assert.False(t, e.me(t, id).IsSuperadmin)
}

func TestTheAPIRoleCannotInsertUsersDirectly(t *testing.T) {
	e := newEnv(t)

	err := e.tm.WithinTx(context.Background(), func(ctx context.Context) error {
		_, err := e.tm.Conn(ctx).Exec(ctx,
			`INSERT INTO users (id, email, created_at, updated_at) VALUES ($1, 'x@example.com', now(), now())`, uuid.New())
		return postgres.TranslateError(err)
	})

	assert.Equal(t, apperror.CodeForbidden, appCode(t, err))
}
