# Payment Gateway

A production-ready POC of a payment gateway built with Go, demonstrating
hexagonal architecture, event-driven design, and clean code practices.

> Built as part of my portfolio to show how I think about architecture,
> not just how I write code.

---

## The Problem

Every payment system eventually faces three hard problems.

**Double charges** — the customer clicks "pay" twice, or the network fails and
the client retries. Without idempotency, the charge runs twice. That is a
chargeback waiting to happen.

**Payment processed, order not updated** — the bank authorized the transaction,
but the system crashed before saving it. Money left the customer account.
Order is still pending. Support team goes crazy.

**Provider lock-in** — today it is Stripe, tomorrow it is PIX, next year it is
something else. Without proper abstraction, every new provider means rewriting
the core.

## How This Solves It

| Problem | Solution |
|---|---|
| Double charges | Redis idempotency key per external_id |
| Lost events | Outbox Pattern — event only leaves after DB persistence |
| Provider lock-in | Ports and Adapters — provider is a swappable adapter |

---

## Architecture

This project uses Hexagonal Architecture (Ports and Adapters). The domain
knows nothing about PostgreSQL, Redis, or HTTP. It only knows interfaces.

The project structure:

- cmd/api — entrypoint, wiring only
- internal/domain — entities, value objects, ports (zero deps)
- internal/usecase — application logic (depends only on ports)
- internal/adapter/http — chi router, handlers, middleware
- internal/adapter/postgres — pgx repository implementations
- internal/adapter/redis — idempotency store
- internal/adapter/event — channel publisher, outbox worker, dispatcher
- pkg/logger — zerolog structured logging
- migrations — versioned SQL up and down
- docs/adr — architecture decision records

The core rule: dependencies point inward. Adapters depend on the domain.
The domain depends on nothing.

---

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.24 |
| HTTP | chi v5 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Logging | zerolog JSON |
| Container | Docker and Docker Compose |
| CI/CD | GitHub Actions |
| Docs | Swagger UI |

---

## Getting Started

The only requirement is Docker.

Clone and run:

    git clone git@github.com:ademarthiago/payment-gateway.git
    cd payment-gateway
    cp .env.example .env
    docker compose up --build

Services available:

| Service | URL |
|---|---|
| API | http://localhost:8088 |
| Health check | http://localhost:8088/health |
| Swagger UI | http://localhost:8089 |
| PostgreSQL | localhost:5433 |
| Redis | localhost:6380 |

---

## API

Create a payment:

    curl -X POST http://localhost:8088/api/v1/payments \
      -H "Content-Type: application/json" \
      -d '{"external_id":"order-001","amount":9900,"currency":"BRL","provider":"mock","description":"Pro plan subscription"}'

Get a payment:

    curl http://localhost:8088/api/v1/payments/\{id\}

Process a refund:

    curl -X POST http://localhost:8088/api/v1/payments/\{id\}/refund \
      -H "Content-Type: application/json" \
      -d '{"amount":9900,"reason":"Customer requested cancellation"}'

Amounts are in cents. 9900 means R$ 99,00.

---

## Running Tests

    make test
    make test-unit
    make test-integration

Or directly with Docker:

    docker compose run --rm app go test ./... -v -race -cover

---

## Architecture Decision Records

Key decisions documented in docs/adr:

- ADR-001 — Why Hexagonal Architecture
- ADR-002 — Why Outbox Pattern for event durability
- ADR-003 — Why Redis for idempotency keys
- ADR-004 — Why zerolog for structured logging

---

## Production Considerations

This is a POC. A production-grade version would also need:

- Real provider integration (Stripe, PagSeguro, PIX)
- OAuth2 or API key authentication
- Rate limiting and throttling
- Circuit breaker pattern
- Retry with exponential backoff
- OpenTelemetry tracing
- Prometheus metrics and Grafana dashboards
- Secrets management (Vault or AWS Secrets Manager)
- PCI-DSS compliance controls
- Outbox cleanup job

These are intentionally out of scope here. The goal was to show the
architectural foundation that makes all of the above possible to add
without rewriting the core.

---

## License

MIT

## Known Issues

### pgx vulnerability (CVE pending)

`github.com/jackc/pgx/v5` versions up to `v5.8.0` have a known critical vulnerability.
The fix is available in `v5.9.2+`, which requires Go 1.25.

This project currently uses Go 1.24. Upgrading to Go 1.25 and pgx v5.9.2 is the
recommended fix for production use.

Tracked in: https://github.com/jackc/pgx/security/advisories
