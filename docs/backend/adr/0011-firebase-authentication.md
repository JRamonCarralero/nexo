# ADR-0011: Firebase Authentication behind an Authenticator interface

**English** · [Español](0011-firebase-authentication.es.md)

- **Status:** Accepted
- **Date:** 2026-10-23

## Context

Building secure authentication (passwords, Google sign-in, email verification, resets) is expensive and risky. The author knows Firebase, whose free tier covers 50,000 monthly active users. The provider might change later.

## Decision

- The web app signs in with **Firebase Authentication** and sends the **ID token** as a bearer token.
- The backend verifies it through the **`Authenticator`** interface (`platform/auth`); the Firebase Admin SDK lives only in `platform/auth/firebase`.
- Locally and in tests, the **Firebase Auth emulator** (`demo-nexo` project, Docker) replaces the real service. The API refuses to start in production with the emulator configured.
- Nexo identifies people by **email**: tokens without email are rejected.

## Consequences

**Positive**

- No passwords stored by Nexo; sign-in methods come for free.
- Changing provider means one new adapter package.
- Integration tests run against the real emulator, no mocks of the SDK.

**Negative / trade-offs**

- Dependency on Google for identity; user accounts live outside our database.
- Token revocation is not checked on every request (it would need an extra call).

## Alternatives considered

- **Own authentication (e.g. Better Auth):** full control, but more security surface to own.
- **Auth0 / Clerk:** good products, but paid beyond small free tiers.
