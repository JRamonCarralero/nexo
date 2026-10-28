# ADR-0018: Testing strategy

**English** · [Español](0018-testing-strategy.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Security (tenant isolation), data access and adapters matter more than coverage numbers. Tests must run the same way locally and in CI.

## Decision

- **Unit** (`make test`, no Docker): services with **mockery** mocks (testify), table-driven tests, the router with fake adapters, and **`arch_test.go`**.
- **Integration** (`make test-integration`, build tag `integration`): real **PostgreSQL** and the **Firebase Auth emulator** with **testcontainers**; one container per package, roles and migrations applied, two pools (API role and owner).
- **Mandatory isolation tests** for every tenant entity, including queries with no `WHERE` to prove RLS works on its own.
- Always `-race`; golangci-lint with depguard. No global coverage target.

## Consequences

**Positive**

- The riskiest parts are tested against real infrastructure.
- The same commands run locally and in CI.

**Negative / trade-offs**

- Integration tests need Docker and take seconds per package.
- The emulator image is built outside testcontainers because local Docker credential helpers can break image builds.

## Alternatives considered

- **Mocks of the database:** fast, but they do not prove RLS, constraints or SQL.
