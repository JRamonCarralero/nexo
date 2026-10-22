//go:build integration

package database_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/dbtest"
)

var testDB *dbtest.Database

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	testDB, err = dbtest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	testDB.Stop(ctx)
	os.Exit(code)
}

func setupTable(t *testing.T) *database.TxManager {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS tx_probe (id int PRIMARY KEY); TRUNCATE tx_probe`)
	require.NoError(t, err)
	return database.NewTxManager(testDB.Pool)
}

func count(t *testing.T) int {
	t.Helper()
	var n int
	require.NoError(t, testDB.Pool.QueryRow(context.Background(), `SELECT count(*) FROM tx_probe`).Scan(&n))
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
	assert.ErrorContains(t, database.TranslateError(err), "ALREADY_EXISTS")
}

func TestMigrationsEnableExtensions(t *testing.T) {
	var n int
	err := testDB.Pool.QueryRow(context.Background(),
		`SELECT count(*) FROM pg_extension WHERE extname IN ('citext','unaccent','pg_trgm')`).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}
