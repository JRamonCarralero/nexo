# ADR-0006: Feature-based structure and platform adapters

**English** · [Español](0006-feature-structure-and-adapters.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The backend is organized by domain with adapters behind interfaces. Organizing the frontend by file type (components/, hooks/, services/) does not scale and lets third-party SDKs (Firebase) spread everywhere.

## Decision

```
src/app/       bootstrap, router, providers, routes
src/features/  one folder per feature, public API through its index.ts
src/shared/    api/ (generated client), platform/ (adapters: i18n, auth, realtime…), ui/
```
ESLint enforces it: importing **Firebase outside `shared/platform`** is an error, and so is importing a **feature's internals** instead of its `index.ts`.

## Consequences

**Positive**

- Mirrors the backend modules; features stay cohesive.
- Changing a provider touches one adapter.

**Negative / trade-offs**

- Some judgement about what is shared and what belongs to a feature.

## Alternatives considered

- **By file type:** familiar, but mixes unrelated code as the app grows.
