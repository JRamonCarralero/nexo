# ADR-0003: Contract-first API with OpenAPI

**English** · [Español](0003-contract-first-api.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Backend (Go) and frontend (TypeScript) cannot share types directly. The API will also be public later (integrations, MCP server, CLI), so it must be stable and documented.

## Decision

`api/openapi.yaml` is the **source of truth**. From it:

- **oapi-codegen** generates the Go server: models, Gin routes and a **strict server** interface (handlers receive typed requests and return typed responses or an error).
- **orval** generates the TypeScript client and TanStack Query hooks.
- The spec is embedded in the binary; the router reads it to know which routes are public ([ADR-0012](0012-secure-by-default-routes.md)).

CI fails if generated code is stale, lints the contract (Redocly) and detects breaking changes against `main` (oasdiff).

## Consequences

**Positive**

- One definition, two typed sides: a contract change that breaks the frontend fails to compile.
- Handlers only `return err`; transport details stay in one place.
- The public API is documented from day one and breaking changes are caught in CI.

**Negative / trade-offs**

- Every endpoint starts in YAML, which is slower than writing a handler directly.
- Generated code has its own quirks (e.g. handler types with the same name need aliases to be embedded together).

## Alternatives considered

- **Code-first** (generate OpenAPI from Go annotations): faster to write, but the contract becomes a by-product and drifts.
- **gRPC / Connect:** strong typing, but a worse fit for a public REST API and browser clients.
