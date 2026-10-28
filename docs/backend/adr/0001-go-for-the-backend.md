# ADR-0001: Go for the backend

**English** · [Español](0001-go-for-the-backend.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Nexo needs an HTTP API that is fast, cheap to run on a free tier (Cloud Run) and maintainable for years by a small team. The author's professional strength is Go on GCP; frontend experience is more limited. TypeScript on both sides was the other serious option because it allows sharing types and validation schemas.

## Decision

The backend is written in **Go** (Gin as HTTP router, pgx for PostgreSQL). Type sharing with the frontend is solved by a contract-first API ([ADR-0003](0003-contract-first-api.md)) instead of a shared language.

## Consequences

**Positive**

- Uses the author's main expertise: fewer mistakes in the most critical part.
- Single static binary, low memory and fast cold starts: a good fit for Cloud Run's free tier.
- The official MCP SDK exists for Go, and local-first sync engines read from PostgreSQL, so neither future feature depends on the backend language.

**Negative / trade-offs**

- Types are not shared natively with the frontend: they have to be generated from the contract.
- Validation rules may need to be expressed twice (contract and Go code).

## Alternatives considered

- **TypeScript with NestJS:** shared types and Zod schemas, and NestJS was already known; rejected because the backend is where the author's experience matters most.
- **Firebase-only backend (Cloud Functions + Firestore):** fastest start, but weak for relational queries and costly per read (see [ADR-0004](0004-postgresql.md)).
