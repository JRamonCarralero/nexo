# ADR-0011: Frontend authentication behind an adapter

**English** · [Español](0011-frontend-authentication.es.md)

- **Status:** Accepted
- **Date:** 2026-11-09

## Context

The API trusts Firebase ID tokens ([backend auth](../../backend/auth.md)). The web app must sign people in (email/password and Google), keep the session, send the token on every request and handle email verification, which the API requires to accept email invitations. We also want to be able to change the identity provider without touching screens, as we do in the backend.

## Decision

- The Firebase JS SDK (modular) is used **only** in `shared/platform/auth/firebase.ts`, behind an `AuthService` interface (state, `getIdToken`, sign in/up, Google, sign out, password reset, verification, `refreshUser`). ESLint forbids importing `firebase` elsewhere.
- React reads the state with `useSyncExternalStore` through `useAuth()`; the service is created in `main.tsx` and passed to `App`, so tests inject a fake one.
- The HTTP client receives a **token provider** and sends `Authorization: Bearer`. On `INVALID_TOKEN` it retries **once** with a force-refreshed token.
- Route **guards** (`requireAuth`, `redirectIfSignedIn`) run in `beforeLoad` using the router context. The `redirect` search param goes through `safeRedirect` (same-app paths only) to avoid open redirects.
- The query cache is cleared whenever the signed-in uid changes.
- Configuration comes from `VITE_FIREBASE_*`. In development it defaults to the local emulator; a **production build refuses** the emulator URL and requires the real project settings.
- After sign-up we force a token refresh so the token carries the display name, and the user profile is created by the API on the first request.

## Consequences

**Positive**
- Screens and tests do not know about Firebase; replacing the provider touches one folder.
- Everything works locally with the emulator and no cloud account.
- A crafted login link cannot redirect people elsewhere; an emulator build cannot reach production by mistake.

**Negative / trade-offs**
- The Firebase SDK adds weight to the bundle (loaded with the app, since every page needs auth).
- The interface covers what we use today; new providers or MFA will extend it.

## Alternatives considered

- **Call Firebase directly from components:** quicker, but couples screens and tests to the SDK.
- **Session cookies issued by our API:** more backend work and no benefit now; the API is already stateless with ID tokens.
