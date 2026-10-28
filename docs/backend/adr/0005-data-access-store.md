# ADR-0005: Technology-agnostic data access: generic Store plus custom queries

**English** · [Español](0005-data-access-store.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The author's practice at work is a generic CRUD layer so that changing the database touches as little as possible. Many Nexo entities are plain CRUD, but many screens need queries that do not fit CRUD (reordering, combined filters, JSONB, search). Every repository must also apply the same cross-cutting rules: tenant filter, soft delete, timestamps, transactions and error translation.

## Decision

- `platform/database` holds **technology-agnostic contracts**: `Store[T]`, `Entity`, `Model`/`TenantModel`, `Query`, `Filter`, `Page`, `Transactor`. Services depend only on these.
- `platform/database/postgres` holds everything PostgreSQL/pgx: connection, `TxManager`, error translation and the generic **`Store[T]`** (Go generics, not `table string`).
- The Store **centralizes the cross-cutting rules**: UUIDv7 ids, timestamps, `created_by`, `workspace_id` from the request scope, soft delete, a transaction per operation and translation to `apperror`.
- Anything that does not fit cleanly is a **custom query** in the module's `store_postgres.go` (sqlc is planned for them). The Store must not grow into a homemade ORM.

Detailed guide: [data-access.md](../data-access.md).

## Consequences

**Positive**

- Little boilerplate for CRUD entities, and repositories cannot forget the tenant filter.
- Changing the engine means a new Store implementation plus the custom queries; services and handlers do not change.
- Full PostgreSQL power where it matters (JSONB, RLS, search).

**Negative / trade-offs**

- Custom queries are engine-specific and must be rewritten if the engine changes.
- Reflection-based mapping is checked at runtime, not at compile time (mitigated by tests and by `NewStore` panicking on invalid entities).

## Alternatives considered

- **ORM (GORM, ent):** less SQL to write, but lowest-common-denominator features and hidden queries.
- **sqlc only:** fully typed, but every CRUD operation written by hand and cross-cutting rules repeated in every query.
- **Generic CRUD with `table string` and `map[string]any`:** what is used at work; no type safety.
