# ADR-0015: Commands and events with a transactional outbox

**English** · [Español](0015-commands-and-outbox.es.md)

- **Status:** Accepted (not implemented yet)
- **Date:** 2026-10-22

## Context

Activity history, notifications, real-time signals and, later, automations, webhooks and AI agents all react to the same changes. Humans and agents must follow the same rules.

## Decision

- Every write is a **command** (`CreateTask`, `MoveTask`…) that checks permissions and rules, changes the data and stores an **event** in `activity_events` **in the same transaction** (outbox pattern).
- A dispatcher (in-process first, Pub/Sub later) delivers events to consumers: activity history, notifications, real-time signals; later automations, webhooks and agents.
- It is **not** event sourcing: state lives in normal tables; events are a log and a trigger.
- This command layer is also what makes a future local-first migration possible.

## Consequences

**Positive**

- New features plug in as event consumers without touching existing code.
- Never a change without its event, nor an event without its change.

**Negative / trade-offs**

- More moving parts: dispatcher, idempotent consumers, retries.

## Alternatives considered

- **Calling side effects directly from services:** simpler, but they get lost on failures and couple modules.
- **Full event sourcing:** powerful, but heavy for this product.
