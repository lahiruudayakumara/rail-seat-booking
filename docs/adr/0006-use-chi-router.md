# ADR 0006: Use Chi router

**Status:** Accepted

## Context

The Go HTTP layer needs routing and composable middleware without a large framework.

## Decision

Use Chi on `net/http`, keeping handlers thin and standard-library-compatible.

## Consequences

Low framework coupling and easy `httptest` usage come with explicit assembly of validation, error mapping and middleware.

## Alternatives Considered

Gin and Fiber offer broader convenience APIs but introduce their own contexts/conventions (Fiber is not `net/http`). The standard library alone is viable, especially with newer patterns, but Chi provides ergonomic groups/parameters. Echo is capable but more framework than needed.
