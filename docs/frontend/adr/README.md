# Architecture Decision Records — Frontend

**English** · [Español](README.es.md)

An ADR records one significant decision: the context, what was decided, why, and what it costs. ADRs are not edited once accepted: a new ADR supersedes an old one and both stay in the history.

| ADR | Title | Status | Date |
|---|---|---|---|
| [0001](0001-react-vite-spa.md) | React single-page application built with Vite | Accepted | 2026-10-22 |
| [0002](0002-typescript-strict.md) | Strict TypeScript, pinned to version 6 | Accepted | 2026-10-22 |
| [0003](0003-tanstack-router.md) | TanStack Router with file-based routes | Accepted | 2026-10-22 |
| [0004](0004-state-management.md) | Server state in TanStack Query, UI state in Zustand | Accepted | 2026-10-22 |
| [0005](0005-generated-api-client.md) | API client generated with orval | Accepted | 2026-10-22 |
| [0006](0006-feature-structure-and-adapters.md) | Feature-based structure and platform adapters | Accepted | 2026-10-22 |
| [0007](0007-styling-and-components.md) | Tailwind CSS and shadcn/ui components | Accepted (partly implemented) | 2026-10-22 |
| [0008](0008-internationalization.md) | Internationalization with i18next, errors translated by code | Accepted | 2026-10-22 |
| [0009](0009-testing-strategy.md) | Testing strategy | Accepted (partly implemented) | 2026-10-22 |
| [0010](0010-node-lts-and-pnpm.md) | Node LTS and pinned pnpm | Accepted | 2026-10-22 |

## Writing a new ADR

1. Copy the template below into `NNNN-short-title.md` (next free number) and its Spanish twin `NNNN-short-title.es.md`.
2. Start as **Proposed**; change to **Accepted** when merged.
3. To change a decision, write a new ADR and mark the old one **Superseded by ADR-NNNN**.
4. Add it to the table above.

## Template

```markdown
# ADR-NNNN: Title

**English** · [Español](NNNN-short-title.es.md)

- **Status:** Proposed | Accepted | Accepted (not implemented yet) | Superseded by ADR-NNNN
- **Date:** YYYY-MM-DD

## Context
What problem or force makes a decision necessary.

## Decision
What we decided, stated plainly.

## Consequences
**Positive**
- …

**Negative / trade-offs**
- …

## Alternatives considered
- **Option:** why it was not chosen.
```
