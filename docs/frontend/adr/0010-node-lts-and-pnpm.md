# ADR-0010: Node LTS and pinned pnpm

**English** · [Español](0010-node-lts-and-pnpm.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Some dependencies (jsdom 30) require recent Node versions, and local and CI environments must match.

## Decision

- **Node 24 LTS**, pinned in `.nvmrc`; CI reads it (`node-version-file`). Odd, non-LTS majors are not used; major bumps are done by hand together with `@types/node` and the emulator image.
- **pnpm**, pinned through `packageManager` in `package.json`; installs use `--frozen-lockfile` in CI.

## Consequences

**Positive**

- Reproducible builds; the same versions everywhere.

**Negative / trade-offs**

- Node majors need a deliberate upgrade.

## Alternatives considered

- **npm / Yarn:** pnpm is faster and stricter with dependencies.
