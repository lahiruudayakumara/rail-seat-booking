# Planned Monorepo Structure

```text
segment-train-booking/
├── apps/
│   ├── api/
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   └── tests/
│   └── web/
│       ├── src/{app,components,features,lib,pages}/
│       └── tests/
├── packages/
│   ├── api-client/
│   ├── shared-types/
│   └── config/
├── database/
│   ├── migrations/
│   ├── queries/
│   └── seed/
├── docs/adr/
├── scripts/
├── .github/workflows/
├── compose.yaml
├── Makefile
├── .env.example
├── .gitignore
├── go.mod
├── pnpm-workspace.yaml
└── README.md
```

`apps/api` is the Go deployable; `cmd/server` is composition/bootstrap, `internal` contains feature and platform code, and `tests` cross-feature/API fixtures. `apps/web` is the Vite SPA: `app` providers/router, `pages` route composition, `features` use cases, `components` shared presentation, and `lib` infrastructure. Frontend filenames use kebab-case (`seat-map.tsx`); Go packages/directories use idiomatic compact lowercase (`trainrun`, not `train_run`).

`api-client` is generated from the normative OpenAPI file and must not be hand-edited. `shared-types` contains frontend-only stable types when generation cannot express a UI concept—not duplicate Go domain models. `config` holds shared lint/TypeScript/Tailwind configurations.

`database/migrations` contains ordered Goose SQL with up/down development paths and safe production guidance; `queries` is sqlc input; `seed` is clearly marked idempotent demo configuration. `docs` records requirements/design/operations/ADRs. `scripts` holds small portable developer/CI helpers. `.github/workflows` validates PR/main. Root Compose/Make/env files offer a consistent interface; `go.mod` can anchor a Go workspace initially, and pnpm manages web packages.

Generated files carry headers and CI checks regeneration. No secrets, local `.env`, build artifacts, node modules or test databases are committed. Implementation scaffolding will be added in Phase 1; this document is not application implementation.
