# ADR-004: Structured JSON Logging with zerolog

## Status
Accepted

## Date
2026-04-22

## Context
Payment systems require auditable, searchable logs. Plain text logs are
difficult to query in log aggregation systems (Datadog, CloudWatch, Loki).

## Decision
We use `zerolog` for structured JSON logging with the following conventions:

- Production: JSON output to stdout (compatible with any log aggregator)
- Development: human-readable console output
- Every HTTP request logs: method, path, status, latency, request_id
- Every error includes the full error chain via `%w` wrapping
- No sensitive data (card numbers, passwords) ever logged

Log levels:
- `DEBUG`: internal flow, event dispatching
- `INFO`: service lifecycle, HTTP requests
- `ERROR`: recoverable errors with context
- `FATAL`: unrecoverable startup failures

## Consequences
**Positive:**
- Logs are machine-parseable by any modern log platform
- zerolog is the fastest Go logger (zero allocation in hot path)
- Consistent format across all services

**Negative:**
- JSON is less readable than plain text for local development
  (mitigated by ConsoleWriter in development mode)
