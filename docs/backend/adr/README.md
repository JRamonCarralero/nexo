# Architecture Decision Records — Backend

**English** · [Español](README.es.md)

An ADR records one significant decision: the context, what was decided, why, and what it costs. ADRs are not edited once accepted: a new ADR supersedes an old one and both stay in the history.

| ADR | Title | Status | Date |
|---|---|---|---|
| [0001](0001-go-for-the-backend.md) | Go for the backend | Accepted | 2026-10-22 |
| [0002](0002-modular-monolith.md) | Modular monolith | Accepted | 2026-10-22 |
| [0003](0003-contract-first-api.md) | Contract-first API with OpenAPI | Accepted | 2026-10-22 |
| [0004](0004-postgresql.md) | PostgreSQL as the main database | Accepted | 2026-10-22 |
| [0005](0005-data-access-store.md) | Technology-agnostic data access: generic Store plus custom queries | Accepted | 2026-10-22 |
| [0006](0006-multi-tenancy-rls.md) | Multi-tenancy with a shared schema and Row-Level Security | Accepted | 2026-10-23 |
| [0007](0007-uuidv7-identifiers.md) | UUIDv7 identifiers generated in Go | Accepted | 2026-10-22 |
| [0008](0008-cursor-pagination.md) | Cursor pagination by default | Accepted | 2026-10-22 |
| [0009](0009-errors-problem-details.md) | Application errors and RFC 9457 Problem Details | Accepted | 2026-10-22 |
| [0010](0010-structured-logging.md) | Structured logging: one line per request | Accepted | 2026-10-22 |
| [0011](0011-firebase-authentication.md) | Firebase Authentication behind an Authenticator interface | Accepted | 2026-10-23 |
| [0012](0012-secure-by-default-routes.md) | Secure-by-default routes derived from the contract | Accepted | 2026-10-23 |
| [0013](0013-user-provisioning.md) | User provisioning through SECURITY DEFINER functions | Accepted | 2026-10-23 |
| [0014](0014-superadmin-bootstrap.md) | Platform superadmin bootstrap | Accepted | 2026-10-23 |
| [0015](0015-commands-and-outbox.md) | Commands and events with a transactional outbox | Accepted (not implemented yet) | 2026-10-22 |
| [0016](0016-custom-fields-jsonb.md) | Custom fields in JSONB | Accepted (not implemented yet) | 2026-10-22 |
| [0017](0017-workspace-in-url-path.md) | Workspace in the URL path | Accepted (not implemented yet) | 2026-10-23 |
| [0018](0018-testing-strategy.md) | Testing strategy | Accepted | 2026-10-22 |
| [0019](0019-migrations-goose.md) | SQL migrations with goose, embedded in the binary | Accepted | 2026-10-22 |
| [0020](0020-local-environment.md) | Local environment with Docker Compose and emulators | Accepted | 2026-10-23 |
| [0021](0021-realtime-signals.md) | Real time through Firestore signal documents | Accepted (not implemented yet) | 2026-10-22 |

## Writing a new ADR

1. Copy the template below into `NNNN-short-title.md` (next free number) and its Spanish twin `NNNN-short-title.es.md`.
2. Start as **Proposed**; change to **Accepted** when merged.
3. To change a decision, write a new ADR and mark the old one **Superseded by ADR-NNNN**.
4. Add it to the table above.

## Template

```markdown
# ADR-NNNN: Title

**English** · [Español](NNNN-short-title.es.md)

- **Status:** Proposed | Accepted | Accepted (not implemented yet) | Superseded by ADR-NNNN
- **Date:** YYYY-MM-DD

## Context
What problem or force makes a decision necessary.

## Decision
What we decided, stated plainly.

## Consequences
**Positive**
- …

**Negative / trade-offs**
- …

## Alternatives considered
- **Option:** why it was not chosen.
```
