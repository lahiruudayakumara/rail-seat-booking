# Submission Notes

| Item | Value |
|---|---|
| GitHub Repository URL | `<add URL>` |
| Live Demo URL | `<not deployed / add URL>` |
| API Documentation URL | `<local: http://localhost:8080/docs / add hosted URL>` |
| Commit Reference | `<add SHA>` |
| Submission Date | `<YYYY-MM-DD>` |

## Delivered

The core vertical slice is implemented: Go/Chi API, PostgreSQL migration and demonstration seed, segment availability, fare quotes, idempotent booking/cancellation, React booking UI and seat map, Compose orchestration, tests, CI and an OpenAPI 3.1 contract.

Core decisions are a Go/Chi modular monolith, React/Vite web client, PostgreSQL, ordered half-open route intervals, a partial GiST exclusion constraint as final inventory authority, explicit parameterized SQL through pgx, idempotent booking creation and integer-minor-unit fare snapshots. Direct no-payment confirmation is initial scope; only `HELD` and `CONFIRMED` block.

The concurrency guarantee is: for one train run and physical seat, two overlapping blocking bookings cannot both commit—even across API replicas. The loser receives 409. Adjacent segments remain sellable. The integration test launches 12 simultaneous transactions and proves exactly one database row succeeds.

Local startup is `cp .env.example .env && docker compose up --build`. Backend unit tests and frontend component tests are included; CI provisions PostgreSQL and runs the tagged 12-contender concurrency integration test three times. Broader E2E, load and security automation remain future hardening.

## Challenges, limitations and alternatives

The principal challenge is modeling absence-based inventory safely under concurrency; application checks and process locks were rejected in favor of PostgreSQL range exclusion. Other trade-offs—Go vs NestJS, Chi vs Gin/Fiber, PostgreSQL vs document/NoSQL stores, sqlc vs ORM, REST vs GraphQL, and monolith vs microservices—are recorded in ADRs.

Known limitations: no payment/auth implementation, group bookings, notifications, waitlist, live railway feeds, official fares/distances, production cloud deployment or admin UI. Demo geography/prices are labelled illustrative. Implemented extra credit is a responsive seat map and conflict-refresh UX; candidates are holds, realtime push, waitlists, analytics, notifications and localization.

## Submission checklist

- Deadline: Tuesday, August 4, 2026 at 11:59 PM
- [Submission form](https://docs.google.com/forms/d/e/1FAIpQLSc4GG1tTq9NcYNJQZVAz-1I0lsQvzD88VNYMnJoK6_4YCczxA/viewform?usp=dialog)
- Verify the GitHub repository is public and the default branch contains the tested implementation
- Replace every placeholder URL, date and commit SHA above
- Run the exact clean-clone Compose flow and capture final test output

## AI use and live ownership

AI tools assisted with documentation structure, consistency review and draft examples. The developer must verify every design, cite any externally sourced official data, disclose assistance per assignment policy, and remain accountable for implementation/security. In a live discussion the developer should be able to derive `[)` overlap, explain SQLSTATE `23P01`/409, trace the booking transaction, justify train vs train run, demonstrate fare snapshot arithmetic, discuss alternatives/failure recovery, and identify which claims are targets rather than implemented facts.
