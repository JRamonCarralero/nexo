# ADR-0003: TanStack Router with file-based routes

**English** · [Español](0003-tanstack-router.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

One product principle is "everything is linkable": every view and filter has a URL. Filters live in search params, so they need to be typed and validated.

## Decision

**TanStack Router** with **file-based routes** in `src/app/routes`. The Vite plugin (and `tsr generate` in CI) builds `routeTree.gen.ts`; routes are code-split automatically. The router context carries the `QueryClient`.

## Consequences

**Positive**

- Fully typed params and search params: a broken link is a compile error.
- Built-in data loading that works with TanStack Query.

**Negative / trade-offs**

- Less known than React Router, with fewer examples.
- One more generated file to keep up to date (checked in CI).

## Alternatives considered

- **React Router:** the standard and used in the author's CMS, but search params are not typed.
