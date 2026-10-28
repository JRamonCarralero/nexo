# ADR-0008: Cursor pagination by default

**English** · [Español](0008-cursor-pagination.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Task lists change in real time while people scroll. Offset pagination skips or repeats items when rows are inserted or moved, and gets slower on deep pages. `COUNT(*)` is not free in PostgreSQL.

## Decision

- `Store.List` uses **keyset (cursor) pagination** by default: `(order column, id)` compared with the last row of the previous page; one extra row is fetched to know if there are more.
- The cursor is opaque (base64url JSON) and only valid for the ordering that produced it.
- **Offset** is available for small administrative lists, never combined with a cursor.
- **`Count`** is a separate call, used only when the UI shows a total; per-group counters use one `GROUP BY`.

## Consequences

**Positive**

- Stable pages with live data and constant cost per page.
- Fits infinite scroll and "load more" interfaces.

**Negative / trade-offs**

- No "jump to page 7"; ordering columns must be non-nullable.
- Cursors break if the ordering changes (they are rejected with `INVALID_CURSOR`).

## Alternatives considered

- **Offset everywhere:** familiar page numbers, but unstable and slow on live data.
