# ADR-0004: PostgreSQL as the main database

**English** · [Español](0004-postgresql.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

A project manager needs combined filters (assignee OR assignee, priority, date ranges, labels), full-text search, relational data (dependencies, memberships, reports) and predictable costs. The author knows Firestore well from work.

## Decision

**PostgreSQL** is the source of truth for all business data (17 locally and in tests; Neon or Cloud SQL when deployed). Extensions `citext`, `unaccent` and `pg_trgm` are enabled from the first migration. Firestore may still be used as a **notification channel** for real time ([ADR-0021](0021-realtime-signals.md)), never as a data store.

## Consequences

**Positive**

- Rich queries, joins, transactions and full-text search (accent-insensitive).
- JSONB gives schema flexibility where it is needed (custom fields, [ADR-0016](0016-custom-fields-jsonb.md)).
- Row-Level Security makes tenant isolation a database guarantee ([ADR-0006](0006-multi-tenancy-rls.md)).
- Predictable cost: Firestore charges per document read (opening a 200-task board = 200 reads).

**Negative / trade-offs**

- No built-in real time or offline mode: they have to be added on top.
- A managed instance is not free on GCP (Cloud SQL); the free option is an external provider (Neon).

## Alternatives considered

- **Firestore:** real time and offline included and well known, but limited OR/range filters, no full-text search and per-read pricing.
- **MySQL:** viable, but weaker JSON support and no Row-Level Security.
