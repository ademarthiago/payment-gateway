# ADR-002: Outbox Pattern for Event Durability

## Status
Accepted

## Date
2026-04-22

## Context
In a payment system, losing an event after a successful database write is
unacceptable. A payment could be persisted but downstream systems (notifications,
fraud detection, analytics) would never know it happened.

The naive approach of "save to DB then publish to queue" fails when:
- The process crashes between the two operations
- The message broker is temporarily unavailable

## Decision
We implemented the Transactional Outbox Pattern:

1. Payment is saved to `payment.payments`
2. Domain event is saved to `payment.outbox` in the same logical flow
3. An `OutboxWorker` polls `payment.outbox` periodically
4. Worker publishes pending events via `EventPublisher`
5. On success, marks the outbox record as `processed`

Additionally, Go channels are used for in-process immediate delivery.
The outbox guarantees at-least-once delivery if the channel publish fails.

## Consequences
**Positive:**
- Event delivery is guaranteed even if the process crashes
- No distributed transactions required
- Simple to implement with existing PostgreSQL

**Negative:**
- At-least-once delivery — consumers must be idempotent
- Polling introduces latency (configurable via OUTBOX_WORKER_INTERVAL_SECONDS)
- Outbox table grows over time — requires periodic cleanup (future work)
