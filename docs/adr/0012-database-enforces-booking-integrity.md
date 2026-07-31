# ADR 0012: Database enforces booking integrity

**Status:** Accepted

## Context

Multiple clients, API replicas, workers and future admin tools can mutate bookings; every path must preserve the same invariant.

## Decision

Treat PostgreSQL constraints, foreign keys, checks and atomic transactions as the final integrity boundary. Application validation remains necessary for useful messages and policy.

## Consequences

Bypassing one service cannot create invalid committed state, and failures roll back atomically. Schema changes demand rigorous migrations and PostgreSQL-aware tests; database errors require stable domain mapping.

## Alternatives Considered

Application-only validation is friendlier but race-prone and duplicated across writers. A dedicated inventory microservice still needs durable atomicity and adds network/operations complexity. Event-sourced inventory can serialize commands but introduces projections, replay and eventual-consistency complexity. A modular monolith plus database constraints is the simplest correct initial architecture.
