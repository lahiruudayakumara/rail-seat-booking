# ADR 0011: Separate train from train run

**Status:** Accepted

## Context

Physical rolling-stock configuration recurs across dated services, while booking conflicts must be isolated per occurrence.

## Decision

Model Train (coaches/seats) separately from TrainRun (route, date/time and operational status).

## Consequences

The same seat identity can be sold independently on different runs and schedules can change without cloning all rolling stock. Equipment substitution needs an explicit future policy/version snapshot.

## Alternatives Considered

Combining train and run duplicates configuration and confuses identity. A schedule template alone cannot represent cancellations/delays/bookings. Copying full coach/seat rows per run simplifies historical snapshots but creates substantial data and synchronization overhead.
