# CI/CD Design

Pull requests run least-privilege, pinned GitHub Actions with concurrency cancellation:

1. Go format check, `go vet ./...`, `golangci-lint`, `go test -race ./...` (unit/integration with PostgreSQL service/testcontainers).
2. `pnpm install --frozen-lockfile`, `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`.
3. OpenAPI lint/bundle plus generated-client drift check.
4. Goose migrate empty DB, seed twice, upgrade a previous fixture and run schema/sqlc verification.
5. Build API/web Docker images, scan dependencies/secrets/images and run Compose smoke tests.

Use separate jobs for fast feedback, then a required aggregate gate. Cache Go build/module directories keyed by `go.sum` and pnpm store keyed by lockfile; never cache `.env`, credentials or mutable build output. Test reports and scan artifacts have short retention and no PII.

Main repeats validation from a clean checkout, builds immutable images once, tags by commit SHA, generates SBOM/provenance, scans and optionally publishes to a registry. Deployment remains a protected, optional environment job with approval, one migration job, rolling/canary rollout, smoke tests and automatic stop/rollback signals. Fork PRs receive no write-capable secrets. Dependabot/Renovate-style updates and scheduled vulnerability/E2E checks are recommended.

Branch protection requires reviewed PRs and passing format, lint, tests, contract, migration, build and scan gates. Production credentials use OIDC/workload identity, not long-lived repository secrets. A failed migration blocks rollout; a failed post-deploy smoke test stops promotion and invokes the recovery runbook.
