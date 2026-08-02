# Release Rollback Runbook

## Before release

Record the application commit, image identifiers, migration status, configuration changes, verification output, and rollback owner. Confirm database changes are backward compatible before deploying application code.

## Rollback decision

Rollback for sustained unavailability, incorrect fares, booking-integrity risk, severe passenger-flow regression, or security exposure. Prefer disabling booking writes over continuing with uncertain integrity.

## Procedure

1. Stop or drain new booking traffic.
2. Restore the previous API and web application versions.
3. Do not roll back the database automatically. Confirm the previous application is compatible with the current schema.
4. If a schema rollback is necessary, back up data and follow the migration-failure runbook.
5. Run health, readiness, smoke, and a controlled booking/cancellation flow.
6. Monitor errors, latency, database health, and conflict rates before reopening traffic.
7. Document the incident, affected interval, cause, corrective action, and follow-up tests without including passenger data.
