# ADR-0024: Invitations by email and by link

**English** · [Español](0024-invitations.es.md)

- **Status:** Accepted
- **Date:** 2026-11-07

## Context

People join a workspace only through an invitation ([ADR-0023](0023-workspaces-and-roles.md)). An invitation is a credential: whoever holds it can enter the workspace. It travels by email or chat, may be forwarded, and its token appears in a URL.

## Decision

- **Two kinds:**
  - **email** — for one person, **single use**, expires in **7 days**. Accepting requires an account with **the same email, verified**: a forwarded link is useless to someone else, and nobody can register the invitee's address without verifying it to take the seat. Google accounts are verified already.
  - **link** — shareable and **reusable** until it expires (**30 days**) or is revoked. It can **never grant the admin role** (enforced by a table check, the creation function and the service).
- **Tokens:** 256 random bits, base64url. Only their **SHA-256 hash** is stored, so a database leak exposes no usable invitation. Email tokens reach only the recipient's inbox; link URLs are returned once, when created.
- **Admins only** create, list and revoke invitations (service check plus RLS). Inviting an email again **revokes the previous** pending invitation; inviting a current member is rejected (`ALREADY_MEMBER`).
- **Accepting** goes through `nexo_accept_invitation` (`SECURITY DEFINER`): the person is not a member yet, so RLS would not let them create their membership. It is **idempotent** and never changes an existing member's role.
- A **preview** (`GET /invitations/{token}`) shows the workspace, role, inviter and status (pending, expired, revoked, accepted) before accepting. Both preview and accept require being signed in.
- Tokens **never reach the logs**: the logging middleware replaces the `token` path parameter with `[REDACTED]`. Malformed tokens are rejected without touching the database.
- If the email cannot be sent, the invitation is **revoked** and the API answers `EMAIL_DELIVERY_FAILED`, so no invitation exists that nobody received. The transactional outbox ([ADR-0015](0015-commands-and-outbox.md)) will turn this into retries.
- Invitations live in the **`workspace` module**: they are how memberships are created.
- The email is **bilingual** (Spanish, then English) because the invitee's language is unknown.

## Consequences

**Positive**

- A leaked database, a forwarded link or an unverified account cannot be used to take an email invitation.
- Rules hold at the database level even if the service is bypassed (tests prove it).

**Negative / trade-offs**

- People who sign up with email and password must verify their address before accepting an email invitation.
- Without the outbox, a temporary SMTP failure makes the admin retry by hand.

## Alternatives considered

- **Store tokens in clear:** simpler lookups, but a database leak would expose every pending invitation.
- **Accept email invitations without a verified email:** less friction, but someone holding a forwarded link could register the address and join.
- **Allow admin links:** convenient, but a link can end up anywhere.
