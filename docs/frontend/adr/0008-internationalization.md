# ADR-0008: Internationalization with i18next, errors translated by code

**English** · [Español](0008-internationalization.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Nexo is bilingual (Spanish and English) from the MVP. Backend error messages are written for developers, not end users.

## Decision

- **i18next** + **react-i18next**, resources in `shared/platform/i18n/locales/{es,en}.json`; language from the browser, English as fallback.
- Errors are translated **from the backend `code`** (`errors.<CODE>`) by `useErrorMessage`, with a generic fallback, and the `traceId` is shown so users can report it.

## Consequences

**Positive**

- Messages are localized and independent of backend wording.

**Negative / trade-offs**

- Every new backend code needs translations in both languages.

## Alternatives considered

- **Lingui / FormatJS:** good options; i18next has the largest ecosystem.
