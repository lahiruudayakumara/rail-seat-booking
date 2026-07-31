# ADR 0003: Use half-open route segments

**Status:** Accepted

## Context

A seat is occupied on travelled legs, not while merely sharing a boundary station.

## Decision

Assign ordered station positions and represent travel as `[origin_position,destination_position)` with origin strictly before destination.

## Consequences

`[0,4)` and `[4,7)` are adjacent and may share a seat; `[0,4)` and `[3,6)` overlap. Boundary arithmetic is unambiguous, but route order must be validated/versioned and reverse services need their own ordering.

## Alternatives Considered

Closed intervals would falsely conflict at a handover station. Open intervals would omit the first travelled leg. Explicit per-leg rows are intuitive and portable but multiply rows and complicate atomic multi-leg acquisition. A bitset is compact but has fixed-size/version/query drawbacks.
