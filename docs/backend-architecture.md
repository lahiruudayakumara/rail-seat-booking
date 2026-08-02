# Backend Architecture

The backend is a feature-oriented Go modular monolith. It keeps the booking transaction inside one process while separating HTTP transport, business policy, persistence, and shared infrastructure.

## Package structure

```text
apps/api/
├── cmd/
│   ├── migrate/                 # Goose migration command
│   └── server/                  # Process startup and graceful shutdown
└── internal/
    ├── app/                     # Composition root and route registration
    ├── config/                  # Environment configuration
    ├── database/                # Narrow pool/transaction interface
    ├── httpmiddleware/          # CORS, security headers, access logging
    ├── platform/
    │   ├── apperror/            # Stable application errors
    │   └── httpx/               # JSON, pagination, IDs, error responses
    ├── apidocs/                 # OpenAPI and documentation endpoints
    ├── health/                  # Liveness and database readiness
    ├── route/                   # Routes
    ├── station/                 # Stations on a route
    ├── trainrun/                # Dated train services
    ├── journey/                 # Shared segment resolution policy
    ├── seat/                    # Shared seat representation
    ├── availability/            # Segment-specific seat search
    ├── fare/                    # Fare calculation and quote persistence
    ├── passengerauth/           # Optional accounts and revocable sessions
    └── booking/                 # Booking lifecycle and transaction boundary
```

Feature packages use only the layers they need:

```text
handler -> service -> repository -> database.DBTX -> PostgreSQL
```

- A **handler** owns HTTP parsing, status codes, headers, and response serialization.
- A **service** owns validation, business policy, and use-case orchestration.
- A **repository** owns parameterized SQL and database result mapping.
- A **model** contains the feature's request, response, and domain data structures.

The `app` package is the only composition root. Constructors make dependencies explicit; there is no global mutable state, dependency-injection framework, or service locator. Packages below `internal` cannot be imported by programs outside this application.

## Dependency and ownership rules

- Transport details do not enter repositories.
- SQL and pgx row scanning do not enter handlers.
- Feature packages do not construct each other's repositories.
- Shared journey resolution is injected into availability and fare services.
- The small `database.DBTX` interface lets repositories work with either a pgx pool or a pgx transaction.
- Cross-cutting HTTP behavior belongs in `httpmiddleware` or `platform/httpx`, not in feature handlers.
- Domain errors are converted to the stable API error envelope in one place.

## Correctness boundary

Availability is a read-side suggestion; it can become stale immediately. Booking creation therefore does not rely on an application-level “check then insert.” PostgreSQL atomically enforces non-overlap using a GiST exclusion constraint over:

```text
(train_run_id, seat_id, int4range(origin_position, destination_position, '[)'))
```

The half-open interval permits adjacent bookings to share the physical seat. For example, `[Colombo, Kandy)` and `[Kandy, Badulla)` do not overlap.

`booking.Service.Create` owns a single transaction that:

1. reserves or replays the idempotency key;
2. locks and validates the unexpired fare quote;
3. inserts the passenger and confirmed booking;
4. records an audit event;
5. connects the idempotency record to the booking; and
6. commits all state together.

PostgreSQL error `23P01` is translated to `409 SEAT_NO_LONGER_AVAILABLE`. This remains correct across goroutines, processes, and multiple API replicas. The integration test starts concurrent transactions and asserts that exactly one overlapping booking succeeds.

Cancellation locks the booking row, applies an idempotent state transition, and writes its audit event in the same transaction.

Guest and registered checkout share the same booking invariant. A guest passenger row has no account owner and is managed through contact verification plus a short-lived booking-management token. A signed-in checkout copies the trusted account profile into the passenger snapshot and records `account_id`; history and cancellation then enforce that ownership. Account sessions are random opaque values, hashed before persistence, revocable, expiring, and transported in HTTP-only SameSite cookies.

## Runtime design

- `config.Load` centralizes database, CORS, pool, server timeout, shutdown, and OpenAPI settings.
- The server sets header, read, write, idle, and graceful-shutdown timeouts.
- `slog` emits structured request and server logs with request IDs.
- Request middleware adds recovery, request IDs, client IP handling, security headers, configured-origin CORS, and access logging.
- JSON bodies are size-limited and reject unknown fields.
- `/health` reports process liveness; `/ready` checks PostgreSQL with a bounded context.
- `/openapi.yaml` serves the contract and `/docs` provides its entry point.

## Why a modular monolith

Booking, idempotency, audit, passenger, and fare-snapshot writes need one strong consistency boundary. Splitting them into services now would introduce network failure and distributed transaction concerns without improving the core booking invariant. Feature packages provide clear ownership and test seams while retaining one deployable API. A package can be extracted later only when its scaling or ownership needs justify the operational cost.

## Testing strategy

- Handler/router tests verify HTTP health and CORS behavior.
- Service unit tests verify fare arithmetic without a database.
- The tagged booking integration test verifies the real PostgreSQL exclusion constraint under contention.
- CI runs normal tests, race detection, vet, compilation, and the database-backed concurrency test.

Run backend checks with:

```bash
go test -race ./...
go vet ./...
go build ./apps/api/cmd/server
```

With migrated and seeded PostgreSQL available:

```bash
TEST_DATABASE_URL='postgres://...' go test -race -tags=integration ./apps/api/internal/booking -run Concurrent -count=3
```
