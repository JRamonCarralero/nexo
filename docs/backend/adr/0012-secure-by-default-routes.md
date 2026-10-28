# ADR-0012: Secure-by-default routes derived from the contract

**English** · [Español](0012-secure-by-default-routes.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

Forgetting to protect a new route is a classic security bug. The code generator does not mark which operations require authentication in the Gin wrapper.

## Decision

- The contract applies `bearerAuth` **globally**; an operation is public only if it declares `security: []` (today only `/health`).
- At startup the router reads the embedded contract and builds the set of public routes.
- The auth middleware requires a token on **every route not in that set**, including routes that are not in the contract at all.

## Consequences

**Positive**

- Forgetting to declare a route can never make it public.
- Public routes are visible and reviewable in the contract.

**Negative / trade-offs**

- Making a route public requires a contract change (intended).

## Alternatives considered

- **Per-route middleware in code:** easy to forget.
- **Request validation middleware from the spec (kin-openapi):** also validates bodies, but heavier; it may be added later.
