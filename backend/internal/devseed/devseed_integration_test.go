//go:build integration

package devseed_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/devseed"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres/pgtest"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
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

// accounts is an identity provider that remembers who it has seen.
type accounts struct{ uids map[string]string }

func (a accounts) Ensure(_ context.Context, address, _, name string) (auth.Identity, error) {
	if a.uids[address] == "" {
		a.uids[address] = "uid-" + address
	}
	return auth.Identity{Subject: a.uids[address], Email: address, EmailVerified: true, Name: name, Provider: "password"}, nil
}

type directory struct{ users *user.Service }

func (d directory) Profiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]workspace.UserProfile, error) {
	list, err := d.users.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]workspace.UserProfile, len(list))
	for _, u := range list {
		out[u.ID] = workspace.UserProfile{ID: u.ID, Email: u.Email, Name: u.Name}
	}
	return out, nil
}

func services() devseed.Services {
	tm := postgres.NewTxManager(testDB.Pool)
	users := user.NewService(user.NewPostgresRepository(tm))
	repo := workspace.NewPostgresRepository(tm)
	return devseed.Services{
		Users:       users,
		Workspaces:  workspace.NewService(repo, directory{users}),
		Invitations: workspace.NewInvitationService(repo, repo, directory{users}, email.Discard{}, "http://localhost:5173"),
	}
}

func memberships(t *testing.T) []string {
	t.Helper()
	rows, err := testDB.OwnerPool.Query(context.Background(),
		`SELECT w.name || ' / ' || u.email || ' / ' || m.role
		   FROM memberships m JOIN workspaces w ON w.id = m.workspace_id JOIN users u ON u.id = m.user_id
		  ORDER BY 1`)
	require.NoError(t, err)
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		out = append(out, s)
	}
	require.NoError(t, rows.Err())
	return out
}

func TestRunCreatesTheExampleData(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.OwnerPool.Exec(ctx, `TRUNCATE invitations, memberships, workspaces, users CASCADE`)
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	acc := accounts{uids: map[string]string{}}

	require.NoError(t, devseed.Run(ctx, log, services(), acc))

	want := []string{
		"Lanzamiento / ana@nexo.local / admin",
		"Lanzamiento / luis@nexo.local / member",
		"Lanzamiento / marta@nexo.local / guest",
		"Marketing / ana@nexo.local / admin",
		"Marketing / luis@nexo.local / admin",
	}
	assert.Equal(t, want, memberships(t))

	// No open link is left behind; the only open invitation is the pending email.
	var open []string
	rows, err := testDB.OwnerPool.Query(ctx, `SELECT kind || ' ' || COALESCE(email, '') FROM invitations WHERE revoked_at IS NULL`)
	require.NoError(t, err)
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		open = append(open, s)
	}
	rows.Close()
	assert.Equal(t, []string{"email nuevo@nexo.local"}, open)

	t.Run("it can be run again without duplicating anything", func(t *testing.T) {
		require.NoError(t, devseed.Run(ctx, log, services(), acc))
		assert.Equal(t, want, memberships(t))
	})
}
