# Booking Conflict Runbook

## Expected conflicts

`409 SEAT_NO_LONGER_AVAILABLE` is normal when passengers race for the same overlapping segment. The client must clear the lost selection, preserve passenger input, refresh availability, and ask the passenger to choose again.

## Investigation trigger

Investigate when conflict rates rise unexpectedly, overlapping bookings appear committed, adjacent segments are rejected, or database error `23P01` is returned as a generic 500.

## Response

1. Correlate safe request IDs, train-run IDs, seat IDs, segment positions, status codes, and database constraint errors. Do not log passenger details or idempotency keys.
2. Confirm only `HELD` and `CONFIRMED` rows participate in the partial exclusion constraint.
3. Verify intervals use `[origin_position,destination_position)` and station positions are ordered.
4. Run the tagged Go concurrency test against an isolated migrated database.
5. Run the k6 contention scenario only against disposable data.
6. If the constraint is absent or invalid, stop booking writes until integrity is restored.

Application availability checks are advisory. Never replace the database constraint with an in-process mutex or a check-then-insert sequence.
