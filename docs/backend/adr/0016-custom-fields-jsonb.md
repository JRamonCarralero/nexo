# ADR-0016: Custom fields in JSONB

**English** · [Español](0016-custom-fields-jsonb.es.md)

- **Status:** Accepted (not implemented yet)
- **Date:** 2026-10-22

## Context

Nexo must be highly configurable (custom fields, task types) without changing the schema for every workspace.

## Decision

Custom values are stored in a **`custom_fields JSONB`** column, described by a `field_definitions` table and indexed with GIN. The data model is prepared from the MVP; the feature arrives in phase 2.

## Consequences

**Positive**

- Fast, filterable with indexes, no schema changes per workspace.

**Negative / trade-offs**

- Type checking of values happens in the application, not in column types.

## Alternatives considered

- **Entity-attribute-value table:** flexible but slow and complex to query.
- **Dynamic columns (`ALTER TABLE`):** unworkable with thousands of workspaces.
