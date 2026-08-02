# Service Unavailable Runbook

## Detection

- `/health` failure indicates the API process or network path is unavailable.
- `/ready` failure indicates PostgreSQL is unavailable or too slow.
- Frontend `/health` failure indicates the Nginx container is unavailable.

## Response

1. Run `docker compose ps` and inspect sanitized `api`, `db`, `migrate`, `seed`, and `web` logs.
2. Check whether migration and seed jobs completed before restarting dependent services.
3. Confirm database credentials, container DNS, disk capacity, memory pressure, and host port ownership.
4. Restart only the failed service when its dependency state is known. Avoid restart loops that hide the first error.
5. Run `make smoke`, then watch readiness, error rate, latency, and booking conflicts.

## Escalation

Escalate immediately for suspected data loss, overlapping committed bookings, credential exposure, or repeated migration failures. Preserve logs and the failed state before destructive recovery.
