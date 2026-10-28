# ADR-0019: SQL migrations with goose, embedded in the binary

**English** · [Español](0019-migrations-goose.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The schema includes things only SQL expresses well: RLS policies, roles, grants, functions. Migrations must run the same way locally, in tests and in deployment.

## Decision

Plain SQL migrations in `backend/db/migrations`, run with **goose** (`make db-migrate`, `db-rollback`, `db-new`, `db-reset`). They are embedded with `go:embed`, so tests apply them without depending on the working directory. Migrations run as the table owner; `db/local/roles.sql` creates the local login roles.

## Consequences

**Positive**

- Full SQL, readable diffs, up and down in the same file.

**Negative / trade-offs**

- No automatic diff from models: migrations are written by hand.

## Alternatives considered

- **Atlas / ORM auto-migrations:** convenient, but awkward for policies, roles and functions.
