# ADR-0004: Server state in TanStack Query, UI state in Zustand

**English** · [Español](0004-state-management.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Most state in a project manager is a cache of server data that other people change at the same time. Mixing it with UI state in one global store leads to stale data and manual synchronization.

## Decision

- **Server data** (tasks, projects…): **TanStack Query**, which caches, retries and handles optimistic updates. Real-time signals invalidate queries. 4xx errors are not retried.
- **UI state** (open panels, selections): **Zustand**, only when it must be shared between components; otherwise local component state.
- Server data is never copied into Zustand.

## Consequences

**Positive**

- One source of truth per kind of state; no manual synchronization.
- Generated hooks (orval) plug straight into TanStack Query.

**Negative / trade-offs**

- Two libraries to learn.

## Alternatives considered

- **Redux Toolkit / RTK Query:** capable, but more boilerplate.
- **Everything in one global store:** stale data and manual cache handling.
