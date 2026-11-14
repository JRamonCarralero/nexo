# ADR-0026: Example data for local development, created through the services

**English** · [Español](0026-dev-seed-data.es.md)

- **Status:** Accepted
- **Date:** 2026-11-14

## Context

Trying the app, demoing it and running end-to-end tests all need a database with believable data: people who can sign in, workspaces with members in different roles and a pending invitation. Writing it as SQL would duplicate business rules (ids, memberships, token hashes, Row-Level Security) and drift as the schema evolves. The seed must also never touch production ([ADR-0014](0014-superadmin-bootstrap.md)).

## Decision

- `make db-seed` runs `cmd/seed`, which builds the same application as the API and calls `devseed.Run` (package `internal/devseed`). `make db-reset` drops everything, migrates and seeds.
- The seed goes through the **real services**: people become users the way they do on a first login (`Resolve`), workspaces are created by their admin, members join by **accepting link invitations** (which are revoked afterwards) and then get their role, and one email invitation is left pending. The data therefore obeys the same rules and RLS as anything created from the app.
- Accounts are created in the **Firebase Auth emulator** (`firebase.EmulatorAccounts`) with verified emails and a known password, so they can sign in. That adapter refuses to work without an emulator host.
- `cmd/seed` **refuses to run** unless `APP_ENV=development` and `FIREBASE_AUTH_EMULATOR_HOST` is set. Invitation emails are discarded (`email.Discard`).
- It is **idempotent**: a workspace that already exists (by name, for its admin) is skipped.
- `devseed` is wiring-level code: modules and platform cannot import it (enforced by `arch_test.go`); `app.Seed` hands it the services.

## Consequences

**Positive**
- One source of truth for the rules; the seed breaks loudly if a rule changes.
- The same data serves manual testing, demos and the e2e suite.
- Safe by construction: development only, emulator only.

**Negative / trade-offs**
- It is slower than raw SQL (a handful of transactions).
- The seed must be updated when a feature changes how people join or how workspaces are created.

## Alternatives considered

- **SQL fixtures loaded with psql:** fast, but they duplicate invariants and bypass RLS and the token flow.
- **Calling the HTTP API from a script:** needs a running server and real ID tokens; the services give the same guarantees with fewer moving parts.
