# ADR-0002: Modular monolith

**English** · [Español](0002-modular-monolith.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The product will grow to many domains (workspaces, projects, tasks, comments, notifications, automations, integrations, AI agents). With one developer, microservices add operational cost without benefit, but a classic layered monolith tends to become a tangle where everything depends on everything.

## Decision

A **modular monolith**: one binary split by domain under `internal/modules/<name>`. Each module has its own layers (handler → service → repository) and exposes a small surface to the rest. Rules:

- A module **never imports another module**. When it needs something from another, it declares a minimal interface (port) and `internal/app` wires the implementation.
- Each table has a single owner module, the only one allowed to write to it.
- Infrastructure lives in `internal/platform` and knows nothing about the business.
- The rules are enforced by `backend/arch_test.go` (imports between our packages) and `depguard` in golangci-lint (third-party libraries).

## Consequences

**Positive**

- Simple to deploy and debug: one process, one database, real transactions across modules.
- Boundaries are explicit and tested, so they survive growth.
- A module could be extracted into a service later without redesign.

**Negative / trade-offs**

- Requires discipline: cross-module needs go through ports, which is more code than a direct call.
- Screens that combine several modules need a read-only query layer (`internal/query`) instead of free joins in services.

## Alternatives considered

- **Layered monolith** (handlers/, services/, repositories/ for everything): fastest to start, ages badly.
- **Pure hexagonal / clean architecture:** clean domain but a lot of ceremony in Go for a mostly CRUD MVP.
- **Microservices:** cost without benefit at this size.
