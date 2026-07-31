# Submission Notes

| Item | Value |
|---|---|
| GitHub Repository URL | `<add URL>` |
| Live Demo URL | `<not deployed / add URL>` |
| API Documentation URL | `<local: http://localhost:8080/docs / add hosted URL>` |
| Commit Reference | `<add SHA>` |
| Submission Date | `<YYYY-MM-DD>` |

## Delivered

At this stage, the complete product/architecture documentation set and OpenAPI 3.1 design are delivered. Application source, runtime Compose files, migrations and tests are intentionally not yet implemented. Update this section during implementation rather than overstating completion.

Core decisions are a Go/Chi modular monolith, React/Vite web client, PostgreSQL, ordered half-open route intervals, a partial GiST exclusion constraint as final inventory authority, pgx/sqlc persistence, idempotent booking creation and integer-minor-unit fare snapshots. Direct no-payment confirmation is initial scope; only `HELD` and `CONFIRMED` block.

The intended concurrency guarantee is: for one train run and physical seat, two overlapping blocking bookings cannot both commit—even across API replicas. The loser receives 409. Adjacent segments remain sellable. The planned test launches at least 10 simultaneous identical requests and proves exactly one database row succeeds.

Local target is `cp .env.example .env && docker compose up --build`. Planned coverage includes domain, repository, integration, API, concurrency, frontend, E2E, migration, smoke, load and security layers.

## Challenges, limitations and alternatives

The principal challenge is modeling absence-based inventory safely under concurrency; application checks and process locks were rejected in favor of PostgreSQL range exclusion. Other trade-offs—Go vs NestJS, Chi vs Gin/Fiber, PostgreSQL vs document/NoSQL stores, sqlc vs ORM, REST vs GraphQL, and monolith vs microservices—are recorded in ADRs.

Known limitations: no payment/auth implementation, group bookings, notifications, waitlist, live railway feeds, official fares/distances, production cloud deployment or completed admin UI. Demo geography/prices must be labeled illustrative. Extra-credit candidates are holds, realtime refresh, waitlists, analytics, notifications, verified localization and polished admin workflows.

## AI use and live ownership

AI tools assisted with documentation structure, consistency review and draft examples. The developer must verify every design, cite any externally sourced official data, disclose assistance per assignment policy, and remain accountable for implementation/security. In a live discussion the developer should be able to derive `[)` overlap, explain SQLSTATE `23P01`/409, trace the booking transaction, justify train vs train run, demonstrate fare snapshot arithmetic, discuss alternatives/failure recovery, and identify which claims are targets rather than implemented facts.
