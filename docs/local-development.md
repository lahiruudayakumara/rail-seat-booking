# Local Development

This is the target workflow for the implementation phase; application containers/configuration are not present in the documentation-only phase.

Prerequisites: Git, Docker Engine/Desktop with Compose v2, and optionally Go/pnpm/Node versions pinned by the future toolchain files. Allocate at least 4 GB to Docker.

```bash
git clone <repository-url>
cd segment-train-booking
cp .env.example .env
docker compose up --build
```

`.env.example` will contain non-secret placeholders for `POSTGRES_*`/database URL, ports, log level, CORS origins, `SERVICE_TIMEZONE=Asia/Colombo`, hold duration and booking horizon. Set local-only credentials in `.env`; never commit it.

Planned commands:

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
| Reset local DB | `docker compose down -v` then start/migrate/seed (destructive) |

Expected URLs are frontend `http://localhost:3000`, API `http://localhost:8080`, docs `http://localhost:8080/docs`, PostgreSQL `localhost:5432`. Verify `/health` then `/ready`.

Troubleshooting: inspect `docker compose ps` and logs first. For port conflicts, stop the owning process or change host-only port values and preserve container ports. For migration failures, identify the first failed Goose version; do not manually mark it successful—fix forward or reset only disposable local data. For unhealthy DB/API, inspect the health command, DB credentials/network, migration job exit, disk space and clock; increase start period on slow machines rather than disabling health checks. Seed must be safe to rerun and fails clearly on schema mismatch.
