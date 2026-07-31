# ADR 0004: Use PostgreSQL exclusion constraints

**Status:** Accepted

## Context

Concurrent replicas can both observe availability before either inserts.

## Decision

Use a partial GiST exclusion constraint on equality of train run/seat and overlap of `[)` position ranges for `HELD`/`CONFIRMED`.

## Consequences

No conflicting pair can commit; SQLSTATE `23P01` maps to 409. This adds PostgreSQL/GiST knowledge and potential hotspot waits, so transactions stay short and are load-tested.

## Alternatives Considered

Application-only checks/mutexes fail across processes. Redis locks add lease/partition/fencing complexity and still need durable verification. `SELECT FOR UPDATE` has no absent booking row to lock unless synthetic locks serialize more work. Serializable transactions/advisory locks are viable but require careful retries/key discipline. Exclusion most directly declares the invariant.
