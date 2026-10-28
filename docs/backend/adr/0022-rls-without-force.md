# ADR-0022: Row-Level Security without FORCE; the owner runs privileged functions

**English** · [Español](0022-rls-without-force.es.md)

- **Status:** Accepted
- **Date:** 2026-10-28

## Context

[ADR-0006](0006-multi-tenancy-rls.md) enabled RLS with `ENABLE` **and** `FORCE`. `FORCE` makes even the **table owner** subject to the policies. The privileged `SECURITY DEFINER` functions ([ADR-0013](0013-user-provisioning.md)) run as the owner and must see and write rows the acting user cannot, so under `FORCE` they broke: creating a user failed with *"new row violates row-level security policy"*.

Nobody noticed because locally and in the first tests the owner was a **superuser**, and superusers always bypass RLS. In managed databases (Neon, Cloud SQL) the owner is not a superuser, so it would have failed in production.

## Decision

- Tables use **`ENABLE ROW LEVEL SECURITY` without `FORCE`**. Migration `00004` removes it from `users`.
- The API still never connects as the owner (it uses `nexo_api`), so RLS applies to every API query exactly as before. Only owner sessions bypass it: migrations and the `SECURITY DEFINER` functions, which is what they need.
- **Integration tests run migrations as a non-superuser owner** (`nexo_owner`), like a managed database.
- Two regression tests: the test owner is not a superuser, and no table forces RLS.

## Consequences

**Positive**

- Privileged functions work the same in tests and in production.
- No loss of protection: the API role is not the owner.

**Negative / trade-offs**

- Any session as the owner sees everything: the owner's credentials must stay out of the API (they already do).

## Alternatives considered

- **Keep FORCE and give the owner BYPASSRLS:** needs a superuser to grant, often not available in managed databases.
- **Keep FORCE and add policies for the owner:** more SQL and easy to get wrong.
