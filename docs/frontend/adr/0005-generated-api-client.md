# ADR-0005: API client generated with orval

**English** · [Español](0005-generated-api-client.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The API contract is the source of truth (backend [ADR-0003](../../backend/adr/0003-contract-first-api.md)). Writing fetch calls and types by hand would drift from it.

## Decision

- **orval** generates types, request functions and **TanStack Query hooks** from `api/openapi.yaml` into `src/shared/api/generated` (never edited by hand; checked in CI).
- Every request goes through one **mutator** (`http-client.ts`): it returns the body on success and throws an **`ApiError`** with the backend `code`, status and `traceId` otherwise (`NETWORK_ERROR` and `UNEXPECTED_RESPONSE` when there is no Problem body).
- `ApiError` is the error type of every generated hook.

## Consequences

**Positive**

- Contract changes break the build where the frontend no longer fits.
- Errors are handled the same way everywhere.

**Negative / trade-offs**

- Generated code is verbose and follows orval's conventions.

## Alternatives considered

- **openapi-typescript + openapi-fetch:** lighter, but no generated query hooks.
- **Hand-written client:** drifts from the contract.
