//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

func setupTable(t *testing.T) *postgres.TxManager {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.OwnerPool.Exec(ctx, `CREATE TABLE IF NOT EXISTS tx_probe (id int PRIMARY KEY); TRUNCATE tx_probe`)
	require.NoError(t, err)
	return postgres.NewTxManager(testDB.Pool)
}

func count(t *testing.T) int {
	t.Helper()
	var n int
	require.NoError(t, testDB.OwnerPool.QueryRow(context.Background(), `SELECT count(*) FROM tx_probe`).Scan(&n))
	return n
}

func TestWithinTxCommits(t *testing.T) {
	tm := setupTable(t)

	err := tm.WithinTx(context.Background(), func(ctx context.Context) error {
		_, err := tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`)
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, 1, count(t))
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	tm := setupTable(t)
	boom := errors.New("boom")

	err := tm.WithinTx(context.Background(), func(ctx context.Context) error {
		if _, err := tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`); err != nil {
			return err
		}
		return boom
	})

	assert.ErrorIs(t, err, boom)
	assert.Equal(t, 0, count(t))
}

func TestWithinTxRollsBackOnPanic(t *testing.T) {
	tm := setupTable(t)

	assert.Panics(t, func() {
		_ = tm.WithinTx(context.Background(), func(ctx context.Context) error {
			_, _ = tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`)
			panic("unexpected")
		})
	})
	assert.Equal(t, 0, count(t))
}

func TestNestedWithinTxJoinsOuterTransaction(t *testing.T) {
	tm := setupTable(t)

	err := tm.WithinTx(context.Background(), func(ctx context.Context) error {
		if err := tm.WithinTx(ctx, func(ctx context.Context) error {
			_, err := tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`)
			return err
		}); err != nil {
			return err
		}
		return errors.New("outer fails")
	})

	require.Error(t, err)
	assert.Equal(t, 0, count(t), "inner work must roll back with the outer transaction")
}

func TestTranslatesRealUniqueViolation(t *testing.T) {
	tm := setupTable(t)
	ctx := context.Background()
	_, err := tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`)
	require.NoError(t, err)

	_, err = tm.Conn(ctx).Exec(ctx, `INSERT INTO tx_probe VALUES (1)`)
	assert.ErrorContains(t, postgres.TranslateError(err), "ALREADY_EXISTS")
}

func TestMigrationsEnableExtensions(t *testing.T) {
	var n int
	err := testDB.Pool.QueryRow(context.Background(),
		`SELECT count(*) FROM pg_extension WHERE extname IN ('citext','unaccent','pg_trgm')`).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}

func TestWithinTxPublishesScope(t *testing.T) {
	tm := setupTable(t)
	user, ws := uuid.New(), uuid.New()
	ctx := scope.WithWorkspace(scope.WithUser(context.Background(), user), ws)

	var gotUser, gotWS *uuid.UUID
	err := tm.WithinTx(ctx, func(ctx context.Context) error {
		return tm.Conn(ctx).QueryRow(ctx, `SELECT nexo_user_id(), nexo_workspace_id()`).Scan(&gotUser, &gotWS)
	})

	require.NoError(t, err)
	require.NotNil(t, gotUser)
	require.NotNil(t, gotWS)
	assert.Equal(t, user, *gotUser)
	assert.Equal(t, ws, *gotWS)
}

func TestWithinTxWithoutScopeLeavesItEmpty(t *testing.T) {
	tm := setupTable(t)

	var gotUser, gotWS *uuid.UUID
	err := tm.WithinTx(context.Background(), func(ctx context.Context) error {
		return tm.Conn(ctx).QueryRow(ctx, `SELECT nexo_user_id(), nexo_workspace_id()`).Scan(&gotUser, &gotWS)
	})

	require.NoError(t, err)
	assert.Nil(t, gotUser)
	assert.Nil(t, gotWS)
}

func TestAPIRoleIsNotTableOwner(t *testing.T) {
	var owner, current string
	err := testDB.Pool.QueryRow(context.Background(),
		`SELECT current_user, (SELECT pg_get_userbyid(relowner) FROM pg_class WHERE relname = 'goose_db_version')`).
		Scan(&current, &owner)
	require.NoError(t, err)
	assert.Equal(t, "nexo_api", current)
	assert.NotEqual(t, owner, current, "the API role must not own tables, or RLS would not apply")
}
