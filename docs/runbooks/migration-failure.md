# Migration Failure Runbook

## Trigger

The migration container exits nonzero, the API remains unready, or schema state differs between environments.

## Response

1. Keep the API stopped and record the first failing migration and PostgreSQL error.
2. Run `DATABASE_URL=... make migration-status` using the same credentials and database.
3. Determine whether the migration transaction rolled back fully. Inspect schema state read-only before changing anything.
4. For an unapplied migration, correct it and test against a fresh database plus a production-like backup.
5. If a migration partially applied outside a transaction, write an explicit forward repair; do not edit migration history already shared with another environment.
6. Re-run migrations, seed only disposable environments, then execute backend tests and smoke checks.

## Rollback

Prefer a forward fix. Use a down migration only when it is proven data-safe and the application version being restored expects the prior schema. Take a verified backup before destructive schema changes.
