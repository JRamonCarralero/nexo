# ADR-0007: Tailwind CSS and shadcn/ui components

**English** · [Español](0007-styling-and-components.es.md)

- **Status:** Accepted (partly implemented)
- **Date:** 2026-10-22

## Context

Nexo is for everyone, including non-technical users: it must be accessible (keyboard, contrast) and consistent, without building a component library from scratch.

## Decision

- **Tailwind CSS 4** (Vite plugin) for styling; class order normalized by the Prettier plugin. *Implemented.*
- **shadcn/ui** (Radix primitives, copied into `shared/ui`), **dnd-kit** for drag and drop and **cmdk** for the `Ctrl+K` command palette. *Added with the first screens.*

## Consequences

**Positive**

- Accessible primitives (Radix) without a heavy dependency: the components are our code.
- Fast iteration with utility classes.

**Negative / trade-offs**

- Long class lists in markup.
- Copied components must be updated by us.

## Alternatives considered

- **MUI / Ant Design:** complete, but heavy and hard to give our own look.
