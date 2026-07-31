# Concurrency and Data Integrity

## Threat model and strategy

Two clients can read the same available seat before either writes. An application mutex protects only one process; `SELECT` then `INSERT` at `READ COMMITTED` is a time-of-check/time-of-use race. Lost updates can also occur during cancellation/status changes, and cached availability is inherently stale.

PostgreSQL is the single inventory authority. A partial GiST exclusion constraint rejects same-run, same-seat overlapping `[)` ranges in `HELD`/`CONFIRMED`. The API uses short transactions, conditional/locked status transitions, unique idempotency records and structured mapping of SQLSTATEs.

## Booking transaction

1. Parse and validate IDs, passenger data and idempotency key.
2. Load the future `SCHEDULED` train run and its route/train.
3. Resolve both route stations and ordered positions from the primary.
4. Require `origin_position < destination_position`; verify active reserved seat belongs to the run's train.
5. Calculate/validate the server fare and build immutable snapshot.
6. Begin a `READ COMMITTED` transaction with bounded statement/lock timeout.
7. Reserve `(scope,key_hash)`; on completed matching request, return stored result; mismatched hash returns 409.
8. Insert booking (initially `CONFIRMED`) and audit event.
9. Let the exclusion constraint validate overlap; update idempotency result.
10. Commit, then return `201`.

Validation may occur just before the transaction to keep locks short, but mutable facts needed for correctness are rechecked or constrained inside it. No confirmation is returned before commit.

## Isolation and locks

`READ COMMITTED` plus declarative constraints is recommended: it gives the invariant without serializing unrelated seats. `SERIALIZABLE` is valid but introduces broader serialization retries; `SELECT FOR UPDATE` has no row to lock when inventory is absence and would require synthetic seat/run locks; advisory/Redis locks add keying and failure risks. Pessimistic row locking is appropriate when changing an existing booking status. `version` enables optimistic conditional updates for admin edits, but cannot alone prevent range overlap.

PostgreSQL may block a competing insert until the first transaction resolves. If it conflicts, pgx exposes SQLSTATE `23P01`; roll back and return:

```json
{
  "code": "SEAT_NO_LONGER_AVAILABLE",
  "message": "This seat is no longer available for the selected journey segment.",
  "details": {"seatId": "UUID", "trainRunId": "UUID"},
  "requestId": "UUID"
}
```

HTTP `409 Conflict` is correct: the request syntax is valid, but current resource state prevents it. Do not retry another seat without user consent. The UI refreshes availability and preserves form input.

## Holds, expiry and cancellation

Holds are optional initially. A future hold inserts `HELD` with a configured UTC expiry and therefore blocks immediately. A worker atomically updates overdue rows with `WHERE status='HELD' AND hold_expires_at <= now()`, emits an audit/outbox event, and is safe to rerun. Queries must not merely ignore an overdue `HELD` before its status commits, or the constraint and query would disagree.

Cancellation locks the booking row, validates allowed status, conditionally sets `CANCELLED`, and audits in one transaction. A repeated cancel returns the already-cancelled resource (recommended idempotent behavior) or the catalogued domain response consistently. `COMPLETED` cannot be cancelled.

## Retries, timeouts and deadlocks

- Retry SQLSTATE `40001` (serialization) and `40P01` (deadlock) at most 2–3 times with capped jitter, only by replaying the entire transaction.
- Never auto-retry `23P01`; it is a seat conflict. Map unique idempotency races according to the stored request hash.
- Apply request, statement and lock timeouts; rollback on context cancellation. Return 503 for transient DB unavailability and avoid claiming failure/success after an unknown commit—clients resolve that with the idempotency key.
- Write related rows in consistent order: idempotency → booking → audit/outbox. Keep network calls outside transactions.

## Concurrency test

Against real PostgreSQL via testcontainers-go, create one run/seat and start at least 10 goroutines behind a closed start channel/barrier. Give each a distinct idempotency key and identical segment, release simultaneously, collect committed results, and assert exactly one `201`/row and nine `409`/`23P01` mappings. Repeat under `-race`, randomize pool sizes, assert no partial audit/idempotency data, then run adjacent, different-seat, different-run, cancellation and hold-expiry variants. The test must query final database state rather than trust only HTTP responses.
