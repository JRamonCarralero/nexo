# ADR-0017: Workspace in the URL path

**English** · [Español](0017-workspace-in-url-path.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

Most operations happen inside a workspace. The API needs to know which one, and validate that the user belongs to it, before reaching the handler.

## Decision

Workspace-scoped endpoints live under **`/api/v1/workspaces/{workspaceId}/…`**. A middleware validates the membership and puts the workspace in the request scope, which then reaches RLS.

## Consequences

**Positive**

- Explicit and linkable: every resource has a full URL.
- Membership is checked in one place before any handler runs.

**Negative / trade-offs**

- Longer paths.

## Alternatives considered

- **`X-Workspace-Id` header:** shorter paths, but the context is hidden and easier to get wrong.
