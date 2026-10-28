# ADR-0009: Application errors and RFC 9457 Problem Details

**English** · [Español](0009-errors-problem-details.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Errors must be meaningful to three audiences: the frontend (which translates messages into Spanish and English), developers reading logs, and clients of the future public API. The author's CMS already used an `AppError` type.

## Decision

- `platform/apperror` defines the error type: **kind** (maps to one HTTP status), stable machine-readable **code** (`TASK_NOT_FOUND`), message for developers, per-field **details**, code **location** and the wrapped **cause** (`Unwrap`, `errors.Is/As`).
- Named constructors (`NotFound`, `Validation`, `Conflict`, `Forbidden`, `Unauthenticated`, `Unavailable`, `Internal`).
- Handlers only `return err`. One middleware converts it to **RFC 9457 Problem Details** with the code and a `traceId`.
- **5xx** responses never expose internal messages; the cause stays in the logs.
- `platform/database/postgres` translates PostgreSQL errors (no rows, unique, foreign key, RLS denial) into application errors.

## Consequences

**Positive**

- The frontend translates by code, so messages are localized and independent of backend wording.
- A standard error format for the public API.
- Modules never deal with HTTP statuses or PostgreSQL codes.

**Negative / trade-offs**

- Codes are part of the API contract: renaming one is a breaking change.
- Every new code needs a translation in the frontend.

## Alternatives considered

- **Plain Go errors + ad hoc JSON per handler:** inconsistent and leaks internals.
- **gRPC-style status codes only:** too coarse for UI messages.
