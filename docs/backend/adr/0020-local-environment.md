# ADR-0020: Local environment with Docker Compose and emulators

**English** · [Español](0020-local-environment.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

Phase 1 is fully local and free. The development machine already has a native PostgreSQL on port 5432.

## Decision

`make up` starts **PostgreSQL 17 on port 5433** (to coexist with a native one) and the **Firebase Auth emulator** (ports 9099 and 4000 for its UI). Mailpit and MinIO will join when email and file storage arrive. Cloud services (Neon, Cloud Run, R2, Resend) are for deployment only.

## Consequences

**Positive**

- Zero cost and no accounts needed to develop.
- The same images are used by integration tests.

**Negative / trade-offs**

- Docker Desktop is required.

## Alternatives considered

- **Shared cloud dev environment:** realistic, but costs money and needs credentials.
