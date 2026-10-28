# ADR-0021: Real time through Firestore signal documents

**English** · [Español](0021-realtime-signals.es.md)

- **Status:** Accepted (not implemented yet)
- **Date:** 2026-10-22

## Context

Long-lived WebSocket or SSE connections keep Cloud Run instances busy and would exhaust the free tier with a few users online all day.

## Decision

On each change, the backend writes a tiny document `signals/{projectId}` in Firestore; the web app listens with `onSnapshot` and refetches from the API. PostgreSQL remains the source of truth. The mechanism sits behind a `Notifier` interface (Go) and a hook (React) so it can be replaced by WebSockets later.

## Consequences

**Positive**

- Fits Firestore's free tier and uses existing experience with it.

**Negative / trade-offs**

- One more service; clients refetch instead of receiving the data.

## Alternatives considered

- **WebSockets/SSE on Cloud Run:** consumes the free tier.
- **Polling:** simplest, less real time.
- **Ably / Pusher:** managed, but another vendor.
