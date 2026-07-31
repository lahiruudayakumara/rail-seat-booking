# ADR 0010: Store a fare snapshot on each booking

**Status:** Accepted

## Context

Fare rules, distances, taxes and rounding can change after purchase; historical value must remain explainable.

## Decision

Persist total in integer minor units plus currency/scale, fare-rule ID and a versioned itemized calculation snapshot at confirmation.

## Consequences

Receipts, audits and future refunds remain stable. Snapshot data duplicates configuration and requires versioned schemas, but is intentionally immutable.

## Alternatives Considered

Recalculate on read risks changing history. Store only total is compact but cannot explain it. Store only a rule reference fails when rules are corrected/retired. `NUMERIC(12,2)` is exact but assumes currency scale; bigint minor units is selected.
