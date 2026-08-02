# Database Recovery Runbook

## Trigger

Use this runbook when `/ready` fails because PostgreSQL is unavailable, corrupted, or restored from backup.

## Response

1. Stop booking traffic or mark the API unavailable; do not accept writes against an uncertain database state.
2. Capture timestamps, API/database logs, migration status, and the last known successful booking reference without copying passenger data.
3. Confirm storage, connectivity, credentials, PostgreSQL health, and available disk space.
4. Restore into an isolated database first and run `DATABASE_URL=... make migration-status`.
5. Verify the booking exclusion constraint, row counts, latest audit events, and several known non-sensitive references.
6. Point a staging API at the restored database and run `make smoke` and `make integration`.
7. Schedule the production cutover, preserve the failed database for investigation, and monitor readiness and booking conflicts.

## Safety

Never run `make reset` against a database containing valuable data. Never mark a failed migration successful manually. Escalate if booking and audit records disagree or if the exclusion constraint is missing.
