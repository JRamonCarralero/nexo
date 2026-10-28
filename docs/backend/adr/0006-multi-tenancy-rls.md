# ADR-0006: Multi-tenancy with a shared schema and Row-Level Security

**English** · [Español](0006-multi-tenancy-rls.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

Leaking data between workspaces is the worst possible bug in a multi-tenant product. Filtering by `workspace_id` in code is necessary but fragile: one forgotten `WHERE` is enough.

## Decision

- One database and schema; every tenant table has `workspace_id`.
- **Row-Level Security** on every tenant table (`ENABLE` + `FORCE`), with policies comparing `workspace_id` to `nexo_workspace_id()`.
- The API connects as **`nexo_api`** (member of **`nexo_app`**), which **does not own the tables** (owners bypass RLS). Migrations run as the owner.
- `TxManager` publishes the request scope at the start of every transaction (`set_config('app.user_id' | 'app.workspace_id', …, true)`), and every Store operation runs in a transaction.
- Column privileges and `SECURITY DEFINER` functions cover what RLS cannot (see [ADR-0013](0013-user-provisioning.md)).

## Consequences

**Positive**

- Defense in depth: isolation holds even if application code forgets a filter.
- The database fails closed: no scope, or no transaction, means no tenant rows.
- Cheap to operate: one schema, normal migrations.

**Negative / trade-offs**

- Every query must run inside a transaction with the scope applied (one extra round trip per transaction).
- Two database roles per environment, and tests need an owner pool and an API-role pool.
- RLS policies are SQL that must be reviewed and tested like code (isolation tests are mandatory per entity).

## Alternatives considered

- **Database or schema per tenant:** strong isolation, but migrations and connections multiply with every workspace.
- **Application filtering only:** simplest, but a single bug leaks data.
