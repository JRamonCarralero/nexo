# ADR-0010: Structured logging: one line per request

**English** · [Español](0010-structured-logging.es.md)

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

Logs will be read in Google Cloud Logging. Logging the same error in every layer creates noise, and errors must be traceable from a user report to the log line.

## Decision

- `log/slog` (standard library). JSON with Cloud Logging field names (`severity`, `message`) outside development; readable text in development.
- A **trace id** per request: taken from `traceparent` or `X-Cloud-Trace-Context`, or generated; returned as `X-Trace-Id` and in every Problem; added to every log record written with the request context.
- **One log line per request**, written by the logging middleware: `INFO` normally, `WARNING` for 4xx without internals, `ERROR` for 5xx with code location and cause chain. Health checks log at `DEBUG`.

## Consequences

**Positive**

- No duplicated logs; each request is one searchable record.
- A user can quote the trace code shown in the UI to find the exact log line.

**Negative / trade-offs**

- Code that wants extra logs must use the request context to keep the trace id.
- Distributed tracing (OpenTelemetry) is not set up yet.

## Alternatives considered

- **zap / zerolog:** faster, but an extra dependency; slog is enough.
- **Log errors where they happen:** simple, but duplicates entries across layers.
