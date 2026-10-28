# ADR-0002: Strict TypeScript, pinned to version 6

**English** · [Español](0002-typescript-strict.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The frontend consumes a generated, fully typed API client. TypeScript 7 (the native Go port) was just released, but linting tools did not support it yet.

## Decision

TypeScript in **strict mode** with extra checks (`noUncheckedIndexedAccess`, `noUnusedLocals`, `verbatimModuleSyntax`), the `@/` alias for `src/`, and **TypeScript 6** pinned until typescript-eslint supports version 7.

## Consequences

**Positive**

- Contract changes surface as compile errors.
- A stable toolchain.

**Negative / trade-offs**

- Missing out on TypeScript 7's faster compiler for now.

## Alternatives considered

- **TypeScript 7 right away:** faster, but lint would break.
