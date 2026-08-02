#!/usr/bin/env sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

unformatted=$(gofmt -l apps/api)
if [ -n "$unformatted" ]; then
  echo "Go files require formatting:" >&2
  echo "$unformatted" >&2
  exit 1
fi

go vet ./...
go test -race ./...
pnpm lint
pnpm typecheck
pnpm test
pnpm build
docker compose config --quiet

echo "Repository verification passed."
