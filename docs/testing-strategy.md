# Testing Strategy

## Layers and gates

- **Go unit:** table-driven domain validation, range examples, status transitions and fare arithmetic; Testify only where clearer.
- **Repository/integration:** real PostgreSQL with testcontainers-go; migrations, constraints, queries, rollback and SQLSTATE mapping. Do not substitute SQLite.
- **API/contract:** `httptest` plus real DB for status, headers, validation, auth placeholder, idempotency and OpenAPI response validation.
- **Frontend:** Vitest/React Testing Library for user-observable form, loading/empty/error, accessibility and 409 behavior; mock at HTTP boundary.
- **E2E:** Playwright against Compose for search → book → reference → cancel and conflict flows across desktop/mobile.
- **Migration:** empty upgrade, previous-version upgrade, seed twice, rollback in development, schema/sqlc drift.
- **Smoke:** liveness/readiness, stations, search and one synthetic booking/cancel in an isolated environment.
- **Load/security:** k6/Vegeta-style scenario design for availability/booking hotspots; SAST, dependency/container scans, secret scan, fuzz parsers, DAST against a test deployment.

## Core matrix

| # | Scenario | Expected | Primary layer |
|---:|---|---|---|
| 1 | Same seat `[0,4)`, `[4,7)` | Both allowed | DB integration |
| 2 | Same seat `[0,4)`, `[3,6)` | Second rejected | DB/API |
| 3 | Same seat exact segment | Second rejected | DB/API |
| 4 | Different seats, same segment | Both allowed | DB |
| 5 | Same seat, different runs | Both allowed | DB |
| 6 | Cancelled booking then overlap | New booking allowed | Integration |
| 7 | Expired hold then overlap | Allowed after committed expiry | Integration |
| 8 | Reverse journey | 422 | Unit/API |
| 9 | Equal endpoints | 422 | Unit/API |
| 10 | 10+ concurrent identical requests | Exactly one commit | Concurrency |
| 11 | Cumulative-distance fare | Exact snapshot | Unit/integration |
| 12 | Disabled/deleted seat | Cannot book | API/DB |
| 13 | Unreserved coach seat | Cannot select/book | API |
| 14 | Past run | Cannot book | API |
| 15 | Station outside route | 422 | API |

Also cover idempotent replay (one booking, same response), mismatched key (409), quote expiry, ambiguous fare rules, boundary time zones/DST assumptions, cancellation races, request cancellation, pool exhaustion and migration failure.

## Goroutine concurrency design

Start PostgreSQL container and seed one run/seat. Create `n >= 10` goroutines, each with a separate client/transaction and idempotency key. Each waits on `<-start`; closing `start` is the barrier. Send identical booking requests, collect results through a buffered channel, wait with `sync.WaitGroup`, then count exactly one success and `n-1` seat conflicts. Query the DB for exactly one blocking booking and one matching audit. Run repeatedly and under `go test -race`; do not weaken synchronization with sleeps.

## Quality and environments

Critical booking/fare/status policy seeks complete branch coverage; coverage trends inform review but do not replace scenario assertions. Tests use UTC internally and explicit `Asia/Colombo` fixtures. Seed data is deterministic, transactions isolate tests where possible, and parallel tests never share inventory keys. PRs run unit/integration/race/frontend/build/migration checks; nightly or pre-release runs E2E, load and deeper security scans. Performance baselines report dataset, hardware, concurrency and percentile—not unsupported guarantees.
