# ADR 0007: Use pgx and sqlc

**Status:** Accepted

## Context

The team needs PostgreSQL-specific features, visible SQL and compile-time checked query bindings without a heavy ORM.

## Decision

Use pgx/pgxpool and sqlc-generated Go code from reviewed SQL; wrap generated queries behind feature repositories.

## Consequences

SQL and plans remain explicit with low runtime magic. Schema/query changes require regeneration, and complex mappings/transactions need deliberate adapters.

## Alternatives Considered

GORM speeds basic CRUD but can obscure SQL/locking and add runtime behavior. Ent offers typed schema/query generation but another abstraction/migration model. Bun is lighter yet still ORM-shaped. Raw pgx maximizes control but repeats scanning/types. sqlc balances explicit SQL and type safety.
