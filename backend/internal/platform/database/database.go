// Package database defines the technology-agnostic data access contracts that
// services and repositories depend on: the generic Store, entities, queries,
// pagination and transactions.
//
// The PostgreSQL implementation lives in the postgres subpackage. Changing the
// database engine means writing a new implementation of these contracts (plus
// the modules' custom queries); services and handlers do not change.
package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Entity is a type persisted by a Store. Fields map to columns through `db`
// struct tags; embed Model or TenantModel to get the standard columns.
type Entity interface {
	// TableName is the table that stores the entity (e.g. "tasks").
	TableName() string
	// ResourceName is the singular name used in error codes (e.g. "task" → TASK_NOT_FOUND).
	ResourceName() string
}

// Model holds the columns every entity has. The Store manages them: the ID is
// generated (UUIDv7) when empty and the timestamps are set by the database.
type Model struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TenantModel is a Model that belongs to a workspace. The Store fills
// WorkspaceID from the request scope and only ever sees rows of that workspace.
type TenantModel struct {
	Model
	WorkspaceID uuid.UUID `db:"workspace_id"`
}

// Optional managed columns, recognized by their tag:
//
//	DeletedAt *time.Time `db:"deleted_at"` // soft delete: Delete sets it, reads skip the row
//	CreatedBy *uuid.UUID `db:"created_by"` // filled from the acting user in the scope

// Changes are the columns to modify in an Update, keyed by column name.
type Changes map[string]any

// Store is the generic CRUD repository. Anything that does not fit it cleanly
// (joins, JSONB, full-text search, bulk updates…) belongs in a custom query.
type Store[T Entity] interface {
	// Create inserts e and refreshes it with the stored values (ID, timestamps…).
	Create(ctx context.Context, e *T) error
	// GetByID returns the entity or a <RESOURCE>_NOT_FOUND error.
	GetByID(ctx context.Context, id uuid.UUID) (T, error)
	// Update applies changes and returns the updated entity.
	Update(ctx context.Context, id uuid.UUID, changes Changes) (T, error)
	// Delete removes the entity (soft delete when it has a deleted_at column).
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns a page of entities matching the query.
	List(ctx context.Context, q Query) (Page[T], error)
	// Count returns how many entities match the filters.
	Count(ctx context.Context, filters ...Filter) (int64, error)
}

// Transactor runs fn in a transaction carried by ctx. Nested calls join the
// outer transaction. It commits when fn returns nil and rolls back otherwise.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
