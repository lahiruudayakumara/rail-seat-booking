# Observability

## Notification delivery

Booking confirmation and cancellation transactions write durable outbox events. Each API process runs a bounded dispatcher using PostgreSQL `FOR UPDATE SKIP LOCKED`, so multiple replicas may safely compete for work. Failed deliveries use capped exponential backoff, and abandoned `PROCESSING` leases become eligible again after five minutes. The admin train-run dashboard exposes pending or failed deliveries; structured logs include message ID, topic, and attempt without passenger data.

Local development uses the log notification provider. Production deployments must replace it with the selected email/SMS adapter and configure that provider's credentials outside version control.

Use OpenTelemetry-compatible boundaries but begin with structured JSON logs and Prometheus metrics. Every request gets an accepted/generated UUID `request_id`; propagate a W3C trace/correlation ID to future dependencies. Return `X-Request-ID`. Log route templates, never high-cardinality raw paths, passenger data, tokens, references or request bodies.

| Metric | Type/labels | Purpose |
|---|---|---|
| `http_requests_total` | counter: method, route, status_class | Traffic/error rate |
| `http_request_duration_seconds` | histogram: method, route, status_class | API latency |
| `booking_attempts_total` | counter: outcome, coach_class | Demand (no IDs) |
| `booking_conflicts_total` | counter: reason | Seat contention |
| `booking_confirmed_total` | counter: coach_class | Successful sales |
| `fare_quotes_total` | counter: outcome | Fare-quote volume/errors |
| `availability_query_duration_seconds` | histogram: outcome | Hot-query latency |
| `database_pool_connections` | gauge: state (`idle`,`acquired`,`max`) | Pool saturation |

Also capture DB acquire duration, transaction duration, deadlocks, hold-expiry lag and queue/outbox backlog when enabled. Avoid station/run/seat UUID labels; use logs/traces for exemplars.

`GET /health` proves the process loop is alive and does not query dependencies. `GET /ready` verifies primary DB connectivity, pool acquisition and expected migration version; it fails during startup/drain. Neither exposes credentials/topology publicly.

Dashboards use RED (rate/errors/duration), booking funnel/outcomes, availability p50/p95/p99 and DB pool/lock views. Suggested alerts: sustained 5xx >2% for 5 min, p95 above SLO, readiness failures across replicas, pool >85% acquired, unexpected conflict-rate step change, zero confirmations with normal attempts, migration mismatch, backup/restore failure. Conflict alerts are baseline-relative because legitimate contention is not a server error.

Tracing is a production placeholder: instrument inbound HTTP, service use cases and sanitized SQL operation names; sample errors/conflicts more highly, never attach PII. Define SLOs only after a measured baseline; initial goals are 99.9% availability and documented p95 targets from requirements. Runbooks link each alert to ownership, diagnostics, mitigation, rollback and communication.
