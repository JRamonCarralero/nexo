# ADR-0023: Workspaces, memberships and roles

**English** · [Español](0023-workspaces-and-roles.es.md)

- **Status:** Accepted
- **Date:** 2026-10-28

## Context

Everything in Nexo happens inside a workspace. People belong to several workspaces with different roles, and isolation between them must not depend on application code only.

## Decision

- One module, **`workspace`**, owns both `workspaces` and `memberships` (a membership has no life outside its workspace).
- Three roles: **admin** (manages the workspace, members and invitations), **member** (works in it) and **guest** (external collaborator; restrictions arrive with projects).
- **Creating a workspace** goes through `nexo_create_workspace`, which inserts the workspace and its first admin atomically. Members join through invitations (next PR); the API role cannot insert memberships.
- **Visibility (RLS):** a user sees the workspaces they belong to, their own memberships everywhere and all memberships of the current workspace; people who share a workspace see each other's profile.
- **Permissions (RLS + service):** only admins rename the workspace and change roles; admins remove anyone; anyone can leave. The service checks first for clear errors (`WORKSPACE_ADMIN_REQUIRED`); RLS enforces the same rules if the service is bypassed. Column privileges allow changing only `name`, `role` and timestamps.
- **At least one admin:** a trigger rejects demoting or removing the last admin (`LAST_ADMIN`), locking the workspace row to avoid races.
- The workspace middleware answers **404** to non-members, so the existence of a workspace is not revealed.
- Helper functions used by policies (`nexo_workspace_role`, `nexo_is_member`, `nexo_shares_workspace`) are `SECURITY DEFINER` so policies on `memberships` can read `memberships` without recursion.

## Consequences

**Positive**

- Isolation and admin rules hold even with a bug in the service.
- A workspace can never be left without an admin.

**Negative / trade-offs**

- Permission rules exist twice (service and policies) and must be kept in sync; integration tests cover both.
- Guest restrictions are deferred until projects exist.

## Alternatives considered

- **Separate `membership` module:** both tables change together, so two modules would talk through ports for no benefit.
- **Roles as a permissions table:** more flexible, but premature with three roles.
