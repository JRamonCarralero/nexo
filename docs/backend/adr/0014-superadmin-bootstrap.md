# ADR-0014: Platform superadmin bootstrap

**English** · [Español](0014-superadmin-bootstrap.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

The platform needs an operator account in every environment, created without manual SQL and without storing passwords.

## Decision

- At startup, `internal/bootstrap` runs idempotent tasks. If `BOOTSTRAP_SUPERADMIN_EMAIL` is set and no superadmin exists, it promotes the user with that email (only if verified) or creates a pre-provisioned row linked on the first login with a verified email.
- It never demotes or deletes anyone. An existing unverified account is **not** promoted (a warning is logged).
- The superadmin is a **platform** role, distinct from a workspace admin, and does **not** bypass workspace isolation (an audited support mode will be designed separately).
- Development data (seed) is a separate mechanism that never runs in production.

## Consequences

**Positive**

- Reproducible in every environment with one variable.
- Safe against email squatting.

**Negative / trade-offs**

- The superadmin has no powers yet; they will be added with explicit, audited features.

## Alternatives considered

- **Manual SQL per environment:** error-prone and undocumented.
- **Superadmin bypassing RLS:** convenient, but the largest possible security hole.
