# ADR 0005: Use a versioned REST API

**Status:** Accepted

## Context

The web client needs a stable, inspectable contract for resource search and booking commands.

## Decision

Expose JSON REST under `/api/v1`, documented by OpenAPI 3.1, with explicit command endpoints for quote/cancel.

## Consequences

HTTP semantics, caches, tooling and generated clients are familiar. Some workflows need multiple requests and over/under-fetching must be managed with purpose-built representations.

## Alternatives Considered

GraphQL gives client-shaped reads but adds schema/resolver complexity and does not simplify transactional booking. gRPC is strong service-to-service but less direct for browsers. Event-driven commands improve decoupling but introduce eventual consistency and operational components inappropriate for initial synchronous confirmation.
