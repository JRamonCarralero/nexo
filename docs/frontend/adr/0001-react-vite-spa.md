# ADR-0001: React single-page application built with Vite

**English** · [Español](0001-react-vite-spa.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

The app lives behind a login, so SEO and server rendering do not matter; speed of interaction does. It needs mature libraries for drag and drop, a command palette and accessible components. The author's strength is the backend; the frontend will be built with heavy help from tooling and AI, so the size of the ecosystem matters.

## Decision

A **React** single-page application built with **Vite**, served as static files. In development, Vite proxies `/api` to the Go API, so there is no CORS.

## Consequences

**Positive**

- The richest ecosystem for what Nexo needs (dnd-kit, cmdk, Radix/shadcn).
- Static hosting (Firebase Hosting or Cloudflare Pages) is free.
- The author's other projects already use React and Vite.

**Negative / trade-offs**

- The first load downloads the whole app (mitigated by route-level code splitting).

## Alternatives considered

- **Angular:** already touched by the author and structured like backend code, but fewer libraries for boards and command palettes.
- **Next.js:** server rendering brings no benefit behind a login and adds a server to run.
