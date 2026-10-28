# ADR-0013: User provisioning through SECURITY DEFINER functions

**English** · [Español](0013-user-provisioning.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

A user must be created on their first authenticated request, before the API knows their Nexo id, so RLS (which works per user) cannot allow it. RLS also filters rows, not columns: a policy letting users update their own row would let them set `is_superadmin`.

## Decision

- Users are created **automatically by the auth middleware** on the first authenticated request (just-in-time).
- The API role **cannot insert users**: the only entry points are two **`SECURITY DEFINER`** functions owned by the table owner: `nexo_resolve_user` and `nexo_bootstrap_superadmin`.
- **Column privileges**: the API role can update only `name`, `avatar_url` and `updated_at`.
- A pre-provisioned account is linked to an identity only when the **email is verified**; otherwise the login gets `EMAIL_IN_USE`.

## Consequences

**Positive**

- No intermediate state of "signed in to Firebase but unknown to Nexo".
- Privileged operations are narrow, named and testable.
- Nobody can make themselves superadmin, even with a bug in the API.

**Negative / trade-offs**

- Some logic lives in PL/pgSQL, which is less familiar and needs integration tests.
- One extra query per authenticated request (cacheable later).

## Alternatives considered

- **Explicit `POST /me` after login:** one more step that can fail.
- **A second database role with BYPASSRLS for system tasks:** broader power held by the same process.
