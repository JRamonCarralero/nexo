# ADR-0007: UUIDv7 identifiers generated in Go

**English** · [Español](0007-uuidv7-identifiers.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Ids must be safe to expose in URLs, unique across environments and good for B-tree indexes and cursor pagination. PostgreSQL 17 has no native UUIDv7 function.

## Decision

Every entity id is a **UUIDv7**, generated in Go (`uuid.NewV7`) by the Store or by the service before calling a database function.

## Consequences

**Positive**

- Time-ordered: inserts append to indexes and ordering by id means creation order, which simplifies cursors.
- Not guessable like serial integers and globally unique.

**Negative / trade-offs**

- 16 bytes instead of 8, and less readable than numbers (short human keys like `NEX-42` will be added for tasks).
- They reveal the approximate creation time.

## Alternatives considered

- **Serial integers:** compact, but guessable and problematic across environments.
- **UUIDv4:** random, which fragments indexes and gives no natural order.
