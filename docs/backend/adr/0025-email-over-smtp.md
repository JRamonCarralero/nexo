# ADR-0025: Email over SMTP behind a Mailer interface

**English** · [Español](0025-email-over-smtp.es.md)

- **Status:** Accepted
- **Date:** 2026-11-07

## Context

Invitations are the first feature that sends email; notifications and digests will follow. Development must not send real email, production needs a reliable provider (Resend's free tier was chosen in the hosting decision), and the provider may change.

## Decision

- Modules depend on the **`email.Mailer`** interface (`platform/email`). The only implementation, `platform/email/smtp`, speaks **SMTP** using the standard library.
- One adapter serves every environment:
  - **Locally and in tests: Mailpit** (Docker), which catches every message and shows it at <http://localhost:8025>. Integration tests send through a Mailpit container and read the message back through its API.
  - **In production: Resend's SMTP relay** (`smtp.resend.com`, port 465, implicit TLS, the API key as password).
- TLS modes `none`, `starttls` and `tls`. The API refuses to start in production with `SMTP_TLS=none`, or with an `APP_BASE_URL` that is not HTTPS (links carry tokens).
- Messages are MIME with a text part and an optional HTML part (`multipart/alternative`), UTF-8, quoted-printable, and Q-encoded subjects.

## Consequences

**Positive**

- No provider SDK in the code; switching provider is a configuration change, or one new adapter.
- What is tested locally is what runs in production (same adapter, same protocol).

**Negative / trade-offs**

- SMTP gives no delivery status (bounces, opens) like a provider's HTTP API would. If that is needed, an HTTP adapter can be added behind the same interface.

## Alternatives considered

- **Resend's HTTP SDK:** richer features, but ties the code to one provider and needs a fake for local development.
- **Logging emails instead of sending them locally:** simpler, but HTML rendering and encoding would go untested.
