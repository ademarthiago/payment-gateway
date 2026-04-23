# ADR-001: Hexagonal Architecture (Ports & Adapters)

## Status
Accepted

## Date
2026-04-22

## Context
We need an architecture that allows the payment domain to evolve independently
from infrastructure concerns (database, cache, HTTP framework). The system must
be testable without spinning up external dependencies.

## Decision
We adopted Hexagonal Architecture (Ports & Adapters) as described by Alistair
Cockburn. The domain defines interfaces (ports) and infrastructure provides
implementations (adapters).

Project structure:

    cmd/api/          - entrypoint (wiring only)
    internal/domain/  - entities, value objects, ports (zero dependencies)
    internal/usecase/ - application logic (depends only on ports)
    internal/adapter/ - postgres, redis, http, event (implements ports)
    pkg/              - shared utilities

## Consequences

**Positive:**
- Domain has zero infrastructure dependencies
- Use cases are testable with mocks, no Docker required
- Adapters are swappable (e.g. replace PostgreSQL with MySQL surgically)
- Clear boundaries enforced by Go's package system

**Negative:**
- More files and boilerplate compared to a layered MVC approach
- Developers unfamiliar with the pattern have a learning curve
