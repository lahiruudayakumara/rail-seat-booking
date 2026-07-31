# ADR 0002: Use PostgreSQL

**Status:** Accepted

## Context

Bookings require relational configuration, transactions and atomic exclusion of overlapping ranges.

## Decision

Use PostgreSQL as system of record, including `int4range`, GiST and `btree_gist` exclusion constraints.

## Consequences

Integrity and queries remain centralized and auditable; operations require PostgreSQL expertise, migrations, backup/restore and connection management. The design is deliberately database-specific.

## Alternatives Considered

MySQL lacks an equivalent declarative range exclusion and would need lock tables/serialization. MongoDB fits flexible documents but makes cross-document inventory invariants awkward. DynamoDB can use conditional writes with a precomputed segment/lock model but adds denormalization and operational complexity. PostgreSQL best expresses the invariant.
