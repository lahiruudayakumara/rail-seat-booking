# Contributing

## Development flow

1. Create a focused branch from `dev`.
2. Keep commits small, buildable, and written in Conventional Commit style.
3. Add or update tests with behavior changes.
4. Update the OpenAPI contract and documentation when public behavior changes.
5. Open a pull request into `dev`; promote tested releases from `dev` to `main`.

Do not combine unrelated backend, frontend, database, and documentation work in one commit. Never commit `.env`, credentials, passenger details, production booking references, or copied production logs.

## Local verification

```bash
cp .env.example .env
docker compose up --build
go test -race ./...
go vet ./...
pnpm lint
pnpm lint:openapi
pnpm typecheck
pnpm test
pnpm build
```

Run the commands relevant to the change and record the exact results in the pull request.

## Database changes

- Add a new ordered Goose migration; never edit a migration already applied outside local disposable environments.
- Verify the latest migration can run `up`, roll back with `down`, and run `up` again against a disposable database.
- Prefer forward-compatible expand/migrate/contract changes.
- Keep seed data deterministic and idempotent.
- Preserve the PostgreSQL exclusion constraint as the final seat-overlap authority.
- Add a contention test for changes affecting booking acquisition or segment occupancy.

## Review expectations

Reviewers check correctness first, particularly transaction boundaries, idempotency, half-open journey ranges, integer money arithmetic, input validation, accessibility, and safe logging. A stale availability response must never weaken the database booking invariant.
