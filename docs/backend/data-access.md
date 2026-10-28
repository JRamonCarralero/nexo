# Data access

**English** · [Español](data-access.es.md)

How the backend reads and writes data: the generic `Store`, the request scope, Row-Level Security and how to add a new table.

## Packages

| Package | Contents | Who uses it |
|---|---|---|
| [`platform/database`](../../backend/internal/platform/database/) | Technology-agnostic contracts: `Store[T]`, `Entity`, `Model`, `TenantModel`, `Query`, `Filter`, `Page`, `Transactor` | Services and repositories |
| [`platform/database/postgres`](../../backend/internal/platform/database/postgres/) | PostgreSQL implementation: connection, `TxManager`, `Store[T]`, error translation | `internal/app` (wiring) and the modules' `store_postgres.go` files |
| [`platform/database/postgres/pgtest`](../../backend/internal/platform/database/postgres/pgtest/) | Throwaway PostgreSQL for integration tests | Tests |
| [`platform/scope`](../../backend/internal/platform/scope/) | The acting user and workspace, carried in the `context` | Middlewares (set it), database layer (reads it) |

Changing the database engine means writing a new implementation of the contracts and rewriting the custom queries. Services and handlers do not change.

## Entities

An entity is a struct with `db` tags that embeds `database.Model` (global) or `database.TenantModel` (belongs to a workspace):

```go
type Task struct {
	database.TenantModel            // id, workspace_id, created_at, updated_at
	Title     string     `db:"title"`
	CreatedBy *uuid.UUID `db:"created_by"` // optional: filled from the acting user
	DeletedAt *time.Time `db:"deleted_at"` // optional: enables soft delete
}

func (Task) TableName() string    { return "tasks" }
func (Task) ResourceName() string { return "task" } // → TASK_NOT_FOUND
```

The Store manages these columns; callers never write them:

| Column | Behaviour |
|---|---|
| `id` | UUIDv7 generated on create if empty (time-ordered: good for indexes and cursors) |
| `workspace_id` | Taken from the request scope. Every read and write filters by it |
| `created_at`, `updated_at` | Set by the database (`now()`); `updated_at` changes on every update |
| `created_by` | The acting user from the scope; `NULL` for system actions (bootstrap, seed) |
| `deleted_at` | `Delete` marks the row; reads skip marked rows |

## The Store

```go
store := postgres.NewStore[Task](txManager) // in internal/app

task := Task{Title: "Write docs"}
err := store.Create(ctx, &task)                       // task now has id, timestamps…
task, err = store.GetByID(ctx, id)                    // TASK_NOT_FOUND if absent
task, err = store.Update(ctx, id, database.Changes{"title": "New title"})
err = store.Delete(ctx, id)
page, err := store.List(ctx, database.Query{
	Filters: []database.Filter{database.Eq("status_id", statusID)},
	OrderBy: "rank", Limit: 50, After: cursor,
})
n, err := store.Count(ctx, database.In("id", ids))
```

Services depend on `database.Store[T]` (the interface), never on `postgres.Store[T]`.

**Anything that does not fit cleanly** (joins, JSONB, full-text search, bulk updates, reordering…) is a custom query with sqlc in the module's `store_postgres.go`. The Store must not grow into a homemade ORM.

### Pagination

- **Cursor (default):** pass `Page.NextCursor` as `Query.After` to get the next page. Stable while rows are created or moved, and constant cost. `OrderBy` must be a non-nullable column; ties are broken by `id`. A cursor only works with the ordering that produced it.
- **Offset:** `Query.Offset` for small administrative lists. It cannot be combined with `After`.
- **Totals:** `Count` is a separate call: only ask for it when the UI shows a total.

## Request scope and transactions

The middlewares put the acting user and the workspace in the `context` (`scope.WithUser`, `scope.WithWorkspace`). The `TxManager` publishes them to PostgreSQL at the start of every transaction:

```sql
SELECT set_config('app.user_id', '<uuid>', true), set_config('app.workspace_id', '<uuid>', true)
```

- **Every Store operation runs in a transaction**, so the scope always reaches the database.
- `TxManager.WithinTx(ctx, fn)` groups several operations (even from different modules) into one transaction. Nested calls join the outer one.
- Custom queries must run inside `WithinTx` and use `TxManager.Conn(ctx)`.

## Row-Level Security

Isolation between workspaces does not depend only on the code remembering to filter: PostgreSQL enforces it.

- The API connects as **`nexo_api`**, member of **`nexo_app`**. It is **not the owner** of the tables (owners bypass RLS). Migrations run as the owner.
- `nexo_user_id()` and `nexo_workspace_id()` return the scope of the current transaction (`NULL` when not set).
- Every tenant table has a policy comparing `workspace_id` with `nexo_workspace_id()`.
- RLS is **enabled, not forced**: the owner (migrations and `SECURITY DEFINER` functions) must bypass it ([ADR-0022](adr/0022-rls-without-force.md)). A test fails if any table forces RLS.

The database **fails closed**: without a scope, or outside a transaction, tenant tables return no rows, and writing a row into another workspace is rejected with `FORBIDDEN`.

### Adding a tenant table

```sql
-- +goose Up
CREATE TABLE tasks (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces (id),
    title        text NOT NULL,
    created_by   uuid,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    deleted_at   timestamptz
);
CREATE INDEX tasks_workspace_idx ON tasks (workspace_id) WHERE deleted_at IS NULL;

ALTER TABLE tasks ENABLE ROW LEVEL SECURITY; -- Not FORCE: the owner must keep bypassing RLS (ADR-0022).
CREATE POLICY tenant_isolation ON tasks
    USING (workspace_id = nexo_workspace_id())
    WITH CHECK (workspace_id = nexo_workspace_id());

-- +goose Down
DROP TABLE tasks;
```

`nexo_app` gets `SELECT, INSERT, UPDATE, DELETE` on new tables automatically (default privileges from migration `00002`).

Then add the isolation tests (mandatory for every tenant entity): another workspace cannot read, update or delete the rows.

## Roles per environment

| Environment | How the roles are created |
|---|---|
| Local | [`db/local/roles.sql`](../../backend/db/local/roles.sql): on a fresh Docker volume and on every `make up` |
| Tests | `pgtest` runs the same file in the test container |
| Production | Provisioned with the infrastructure (never with the local password) |

## Testing

```go
// TestMain, once per package (build tag: integration)
testDB, err = pgtest.Start(ctx)

// testDB.Pool      → connects as nexo_api: RLS applies, like production
// testDB.OwnerPool → table owner (nexo_owner, NOT a superuser, like a managed
//                    database): create fixtures or inspect rows bypassing RLS
```

See [`store_integration_test.go`](../../backend/internal/platform/database/postgres/store_integration_test.go) for examples, including RLS tests that run queries with no `WHERE` at all.
