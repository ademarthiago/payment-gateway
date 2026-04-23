# ADR-003: Redis-based Idempotency Keys

## Status
Accepted

## Date
2026-04-22

## Context
Payment APIs are called over unreliable networks. Clients retry on timeout,
resulting in duplicate requests. Without idempotency, the same payment could
be charged twice — a critical business and compliance failure.

## Decision
Each payment creation request includes an `external_id` (the idempotency key).
Before processing:

1. Check Redis for `payment:create:{external_id}`
2. If exists, return the existing payment (no duplicate processing)
3. If not, process normally, then set the key with 24h TTL

Redis was chosen over PostgreSQL for idempotency checks because:
- O(1) lookup vs index scan
- TTL support built-in
- Reduces load on the primary database

## Consequences
**Positive:**
- Prevents double charges on network retries
- Fast O(1) Redis lookup before any DB operation
- 24h TTL balances safety and storage

**Negative:**
- Redis unavailability degrades (not blocks) the service
- TTL expiry means keys older than 24h won't deduplicate (acceptable for payments)
