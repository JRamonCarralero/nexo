//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// widget is a tenant entity with soft delete and created_by, protected by RLS
// exactly like the real tables will be.
type widget struct {
	database.TenantModel
	Name      string     `db:"name"`
	Rank      int        `db:"rank"`
	CreatedBy *uuid.UUID `db:"created_by"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (widget) TableName() string    { return "widgets" }
func (widget) ResourceName() string { return "widget" }

// gadget is a global entity (no workspace) with hard delete.
type gadget struct {
	database.Model
	Label string `db:"label"`
}

func (gadget) TableName() string    { return "gadgets" }
func (gadget) ResourceName() string { return "gadget" }

const storeSchema = `
DROP TABLE IF EXISTS widgets, gadgets;

CREATE TABLE widgets (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    name         text NOT NULL,
    rank         int  NOT NULL DEFAULT 0,
    created_by   uuid,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    deleted_at   timestamptz
);
ALTER TABLE widgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE widgets FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON widgets
    USING (workspace_id = nexo_workspace_id())
    WITH CHECK (workspace_id = nexo_workspace_id());

CREATE TABLE gadgets (
    id         uuid PRIMARY KEY,
    label      text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
`

type fixture struct {
	tm      *postgres.TxManager
	widgets *postgres.Store[widget]
	gadgets *postgres.Store[gadget]
	user    uuid.UUID
	wsA     uuid.UUID
	wsB     uuid.UUID
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	_, err := testDB.OwnerPool.Exec(context.Background(), storeSchema)
	require.NoError(t, err)

	tm := postgres.NewTxManager(testDB.Pool)
	return fixture{
		tm:      tm,
		widgets: postgres.NewStore[widget](tm),
		gadgets: postgres.NewStore[gadget](tm),
		user:    uuid.New(),
		wsA:     uuid.New(),
		wsB:     uuid.New(),
	}
}

func (f fixture) in(ws uuid.UUID) context.Context {
	return scope.WithWorkspace(scope.WithUser(context.Background(), f.user), ws)
}

func (f fixture) create(t *testing.T, ws uuid.UUID, name string, rank int) widget {
	t.Helper()
	w := widget{Name: name, Rank: rank}
	require.NoError(t, f.widgets.Create(f.in(ws), &w))
	return w
}

func requireAppError(t *testing.T, err error, code string) {
	t.Helper()
	var appErr *apperror.Error
	require.True(t, errors.As(err, &appErr), "expected an application error, got %v", err)
	assert.Equal(t, code, appErr.Code)
}

func TestCreateFillsManagedColumns(t *testing.T) {
	f := newFixture(t)

	w := f.create(t, f.wsA, "first", 1)

	assert.Equal(t, uuid.Version(7), w.ID.Version(), "ids are UUIDv7")
	assert.Equal(t, f.wsA, w.WorkspaceID, "workspace comes from the scope")
	require.NotNil(t, w.CreatedBy)
	assert.Equal(t, f.user, *w.CreatedBy, "created_by comes from the scope")
	assert.WithinDuration(t, time.Now(), w.CreatedAt, time.Minute)
	assert.Equal(t, w.CreatedAt, w.UpdatedAt)
	assert.Nil(t, w.DeletedAt)
}

func TestCreateKeepsExplicitID(t *testing.T) {
	f := newFixture(t)
	id := uuid.Must(uuid.NewV7())

	w := widget{Name: "explicit"}
	w.ID = id
	require.NoError(t, f.widgets.Create(f.in(f.wsA), &w))

	assert.Equal(t, id, w.ID)
}

func TestCreateIgnoresWorkspaceSetByCaller(t *testing.T) {
	f := newFixture(t)

	w := widget{Name: "sneaky"}
	w.WorkspaceID = f.wsB
	require.NoError(t, f.widgets.Create(f.in(f.wsA), &w))

	assert.Equal(t, f.wsA, w.WorkspaceID, "the scope always wins")
}

func TestGetByID(t *testing.T) {
	f := newFixture(t)
	w := f.create(t, f.wsA, "first", 1)

	got, err := f.widgets.GetByID(f.in(f.wsA), w.ID)
	require.NoError(t, err)
	assert.Equal(t, w, got)

	_, err = f.widgets.GetByID(f.in(f.wsA), uuid.New())
	requireAppError(t, err, "WIDGET_NOT_FOUND")
}

func TestOtherWorkspaceCannotSeeOrTouchRows(t *testing.T) {
	f := newFixture(t)
	w := f.create(t, f.wsA, "secret", 1)
	other := f.in(f.wsB)

	_, err := f.widgets.GetByID(other, w.ID)
	requireAppError(t, err, "WIDGET_NOT_FOUND")

	_, err = f.widgets.Update(other, w.ID, database.Changes{"name": "hacked"})
	requireAppError(t, err, "WIDGET_NOT_FOUND")

	requireAppError(t, f.widgets.Delete(other, w.ID), "WIDGET_NOT_FOUND")

	page, err := f.widgets.List(other, database.Query{})
	require.NoError(t, err)
	assert.Empty(t, page.Items)

	n, err := f.widgets.Count(other)
	require.NoError(t, err)
	assert.Zero(t, n)

	got, err := f.widgets.GetByID(f.in(f.wsA), w.ID)
	require.NoError(t, err)
	assert.Equal(t, "secret", got.Name, "the row is untouched")
}

// The Store filters by workspace, but RLS must protect the data on its own:
// these queries have no WHERE clause at all.
func TestRowLevelSecurityWithoutStoreFilters(t *testing.T) {
	f := newFixture(t)
	f.create(t, f.wsA, "a1", 1)
	f.create(t, f.wsA, "a2", 2)
	f.create(t, f.wsB, "b1", 1)

	countIn := func(ctx context.Context) int {
		var n int
		require.NoError(t, f.tm.WithinTx(ctx, func(ctx context.Context) error {
			return f.tm.Conn(ctx).QueryRow(ctx, `SELECT count(*) FROM widgets`).Scan(&n)
		}))
		return n
	}
	assert.Equal(t, 2, countIn(f.in(f.wsA)))
	assert.Equal(t, 1, countIn(f.in(f.wsB)))
	assert.Equal(t, 0, countIn(context.Background()), "no scope, no rows")

	var outside int
	require.NoError(t, testDB.Pool.QueryRow(context.Background(), `SELECT count(*) FROM widgets`).Scan(&outside))
	assert.Zero(t, outside, "outside a transaction RLS fails closed")
}

func TestRowLevelSecurityRejectsWritesIntoAnotherWorkspace(t *testing.T) {
	f := newFixture(t)

	err := f.tm.WithinTx(f.in(f.wsA), func(ctx context.Context) error {
		_, err := f.tm.Conn(ctx).Exec(ctx,
			`INSERT INTO widgets (id, workspace_id, name, created_at, updated_at) VALUES ($1, $2, 'x', now(), now())`,
			uuid.New(), f.wsB)
		return postgres.TranslateError(err)
	})

	requireAppError(t, err, apperror.CodeForbidden)
}

func TestTenantEntityRequiresWorkspaceScope(t *testing.T) {
	f := newFixture(t)
	ctx := scope.WithUser(context.Background(), f.user)

	err := f.widgets.Create(ctx, &widget{Name: "orphan"})
	requireAppError(t, err, apperror.CodeInternal)

	_, err = f.widgets.List(ctx, database.Query{})
	requireAppError(t, err, apperror.CodeInternal)
}

func TestUpdate(t *testing.T) {
	f := newFixture(t)
	w := f.create(t, f.wsA, "before", 1)
	ctx := f.in(f.wsA)

	got, err := f.widgets.Update(ctx, w.ID, database.Changes{"name": "after", "rank": 5})
	require.NoError(t, err)
	assert.Equal(t, "after", got.Name)
	assert.Equal(t, 5, got.Rank)
	assert.Equal(t, w.CreatedAt, got.CreatedAt)

	unchanged, err := f.widgets.Update(ctx, w.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, got, unchanged, "no changes means a plain read")
}

func TestUpdateRejectsManagedAndUnknownColumns(t *testing.T) {
	f := newFixture(t)
	w := f.create(t, f.wsA, "x", 1)
	ctx := f.in(f.wsA)

	for _, col := range []string{"id", "workspace_id", "created_at", "updated_at", "created_by", "deleted_at", "nope"} {
		t.Run(col, func(t *testing.T) {
			_, err := f.widgets.Update(ctx, w.ID, database.Changes{col: nil})
			requireAppError(t, err, apperror.CodeInternal)
		})
	}
}

func TestSoftDelete(t *testing.T) {
	f := newFixture(t)
	w := f.create(t, f.wsA, "doomed", 1)
	ctx := f.in(f.wsA)

	require.NoError(t, f.widgets.Delete(ctx, w.ID))

	_, err := f.widgets.GetByID(ctx, w.ID)
	requireAppError(t, err, "WIDGET_NOT_FOUND")
	requireAppError(t, f.widgets.Delete(ctx, w.ID), "WIDGET_NOT_FOUND")

	n, err := f.widgets.Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, n)

	var deletedAt *time.Time
	require.NoError(t, testDB.OwnerPool.QueryRow(context.Background(),
		`SELECT deleted_at FROM widgets WHERE id = $1`, w.ID).Scan(&deletedAt))
	assert.NotNil(t, deletedAt, "the row is kept and marked")
}

func TestListCursorPagination(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)
	var all []uuid.UUID
	for i := range 5 {
		all = append(all, f.create(t, f.wsA, "w", i).ID)
	}

	var seen []uuid.UUID
	q := database.Query{Limit: 2}
	pages := 0
	for {
		page, err := f.widgets.List(ctx, q)
		require.NoError(t, err)
		pages++
		for _, w := range page.Items {
			seen = append(seen, w.ID)
		}
		if !page.HasMore {
			assert.Empty(t, page.NextCursor)
			break
		}
		require.NotEmpty(t, page.NextCursor)
		q.After = page.NextCursor
	}

	assert.Equal(t, 3, pages)
	assert.Equal(t, all, seen, "default order is id, i.e. creation order")
}

func TestListOrderByColumnDescWithTies(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)
	low := f.create(t, f.wsA, "low", 1)
	tieA := f.create(t, f.wsA, "tieA", 5)
	tieB := f.create(t, f.wsA, "tieB", 5)
	high := f.create(t, f.wsA, "high", 9)

	var names []string
	q := database.Query{OrderBy: "rank", Desc: true, Limit: 1}
	for {
		page, err := f.widgets.List(ctx, q)
		require.NoError(t, err)
		for _, w := range page.Items {
			names = append(names, w.Name)
		}
		if !page.HasMore {
			break
		}
		q.After = page.NextCursor
	}

	// Ties on rank are broken by id in the same direction.
	assert.Equal(t, []string{high.Name, tieB.Name, tieA.Name, low.Name}, names)
}

func TestListFiltersAndCount(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)
	a := f.create(t, f.wsA, "alpha", 1)
	b := f.create(t, f.wsA, "beta", 2)
	f.create(t, f.wsA, "gamma", 2)

	page, err := f.widgets.List(ctx, database.Query{Filters: []database.Filter{database.Eq("rank", 2)}})
	require.NoError(t, err)
	assert.Len(t, page.Items, 2)

	page, err = f.widgets.List(ctx, database.Query{Filters: []database.Filter{database.In("id", []uuid.UUID{a.ID, b.ID})}})
	require.NoError(t, err)
	assert.Len(t, page.Items, 2)

	n, err := f.widgets.Count(ctx, database.Eq("rank", 2), database.NotNull("created_by"))
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
}

func TestListOffset(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)
	for i := range 3 {
		f.create(t, f.wsA, "w", i)
	}

	page, err := f.widgets.List(ctx, database.Query{OrderBy: "rank", Limit: 2, Offset: 2})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, 2, page.Items[0].Rank)
	assert.False(t, page.HasMore)
	assert.Empty(t, page.NextCursor)
}

func TestListRejectsInvalidQueries(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)

	_, err := f.widgets.List(ctx, database.Query{After: "not-a-cursor"})
	requireAppError(t, err, "INVALID_CURSOR")

	_, err = f.widgets.List(ctx, database.Query{After: "x", Offset: 1})
	requireAppError(t, err, "INVALID_QUERY")

	_, err = f.widgets.List(ctx, database.Query{OrderBy: "deleted_at"})
	requireAppError(t, err, apperror.CodeInternal) // nullable columns cannot drive a cursor

	_, err = f.widgets.List(ctx, database.Query{Filters: []database.Filter{database.Eq("nope", 1)}})
	requireAppError(t, err, apperror.CodeInternal)
}

func TestCursorFromAnotherOrderingIsRejected(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)
	for i := range 3 {
		f.create(t, f.wsA, "w", i)
	}
	page, err := f.widgets.List(ctx, database.Query{Limit: 1})
	require.NoError(t, err)

	_, err = f.widgets.List(ctx, database.Query{OrderBy: "rank", After: page.NextCursor})
	requireAppError(t, err, "INVALID_CURSOR")
}

func TestGlobalEntityWithHardDelete(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background() // global entities need no workspace

	g := gadget{Label: "g"}
	require.NoError(t, f.gadgets.Create(ctx, &g))
	require.NoError(t, f.gadgets.Delete(ctx, g.ID))

	var n int
	require.NoError(t, testDB.OwnerPool.QueryRow(ctx, `SELECT count(*) FROM gadgets`).Scan(&n))
	assert.Zero(t, n, "rows without deleted_at are really deleted")
}

func TestTransactionSpansSeveralStores(t *testing.T) {
	f := newFixture(t)
	ctx := f.in(f.wsA)

	err := f.tm.WithinTx(ctx, func(ctx context.Context) error {
		if err := f.widgets.Create(ctx, &widget{Name: "w"}); err != nil {
			return err
		}
		if err := f.gadgets.Create(ctx, &gadget{Label: "g"}); err != nil {
			return err
		}
		return errors.New("abort")
	})
	require.Error(t, err)

	nw, err := f.widgets.Count(ctx)
	require.NoError(t, err)
	ng, err := f.gadgets.Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, nw+ng, "both inserts roll back together")
}
