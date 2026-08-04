# Local Development

The application is containerized for a clean-machine workflow.

Prerequisites: Git and Docker Engine/Desktop with Compose v2. Go 1.25, Node 24 and pnpm 11.14 are optional for running services natively. Allocate at least 4 GB to Docker.

```bash
git clone https://github.com/lahiruudayakumara/rail-seat-booking.git
cd rail-seat-booking
cp .env.example .env
docker compose up --build
```

`.env.example` contains non-secret local placeholders for the database, ports, CORS origins and service settings. Set local-only credentials in `.env`; never commit it. Compose automatically migrates and loads idempotent demonstration data before starting the API.

Commands:

| Task | Command |
|---|---|
| Start/rebuild | `docker compose up --build` |
| Background start | `docker compose up --build -d` |
| Logs | `docker compose logs -f api web db migrate` |
| Migrate | `make migrate-up` |
| Demo seed | `make seed` |
| All tests | `make test` |
| Stop | `docker compose down` |
| Rebuild one service | `docker compose build --no-cache api` then `docker compose up -d api` |
| Reset local DB | `make reset` (destructive to the local Compose volume) |

Expected URLs are frontend `http://localhost:3000`, API `http://localhost:8080`, and docs `http://localhost:8080/docs`. PostgreSQL remains private inside the Compose network at `db:5432`. Verify `/health` then `/ready`.

Troubleshooting: inspect `docker compose ps` and logs first. For port conflicts, stop the owning process or change host-only port values and preserve container ports. For migration failures, identify the first failed Goose version; do not manually mark it successful—fix forward or reset only disposable local data. For unhealthy DB/API, inspect the health command, DB credentials/network, migration job exit, disk space and clock; increase start period on slow machines rather than disabling health checks. Seed must be safe to rerun and fails clearly on schema mismatch.
