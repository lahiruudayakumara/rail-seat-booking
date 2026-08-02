# Monorepo Structure

```text
rail-seat-booking/
├── apps/
│   ├── api/                    # Go modular-monolith API and migration command
│   └── web/                    # React/Vite passenger application
├── database/
│   ├── migrations/             # Ordered Goose schema migrations
│   ├── queries/                # SQL query definitions
│   └── seed/                   # Idempotent demonstration data
├── docs/
│   ├── adr/                    # Architecture decision records
│   └── runbooks/               # Operational response procedures
├── examples/api/               # Executable HTTP request examples
├── scripts/                    # Portable verification and smoke helpers
├── tests/
│   ├── integration/            # Running-system API journeys
│   └── load/                   # Disposable-environment contention tests
├── .github/                    # Governance, issue forms, dependency automation, CI
├── compose.yaml                # One-command local orchestration
├── Makefile                    # Stable developer command interface
├── .env.example                # Non-secret environment contract
├── go.mod                      # Go dependency boundary
├── package.json                # JavaScript workspace commands
├── pnpm-workspace.yaml         # Frontend workspace membership
└── README.md
```

## Ownership rules

`apps/api` is the Go deployable. Its `cmd` packages contain process entry points, `internal/app` is the composition root, feature packages own business capabilities, and `internal/platform` owns shared transport conventions. Go unit and package tests remain beside source.

`apps/web` is the browser deployable. API adapters, application state, hooks, shared components, layouts, route pages, feature sections, localization, and types have separate directories. Component tests remain beside frontend source.

`database` owns durable schema and demonstration data. Existing shared migrations must not be edited after use outside disposable development databases; changes receive new ordered migrations.

Root `tests` are reserved for behavior crossing package, process, or application boundaries. Root `scripts` contain small portable commands used by both developers and CI. Complex logic belongs in tested application code rather than shell scripts.

`examples` must contain demonstration values only. `docs/runbooks` describe operational response, while architecture and ADR files explain design. No credentials, local `.env`, passenger data, build output, package caches, or test databases are committed.

Shared `packages/`, deployment manifests, and infrastructure directories should be introduced only when real code or an actual target requires them; empty architecture placeholders are intentionally avoided.
