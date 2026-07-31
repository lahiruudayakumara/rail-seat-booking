# ADR 0001: Use Go for the backend

**Status:** Accepted

## Context

The API needs predictable concurrency, fast startup, simple containers, strong tooling and maintainable explicit transactions.

## Decision

Use Go for the modular-monolith REST API, standard context/error conventions and standard testing package.

## Consequences

Small binaries and straightforward goroutine-based concurrency suit API/worker tasks. Engineers must handle errors explicitly and avoid leaking transport/persistence types into policy; some UI-shared types require OpenAPI generation.

## Alternatives Considered

NestJS offers a productive batteries-included TypeScript ecosystem but heavier runtime/framework indirection. Java/.NET provide mature enterprise ecosystems with more baseline ceremony/resources. Rust provides stronger low-level safety but higher team learning cost. Go was chosen for this scope, not as a universal performance claim.
