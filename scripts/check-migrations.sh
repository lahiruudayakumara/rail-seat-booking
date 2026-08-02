#!/usr/bin/env sh
set -eu

: "${DATABASE_URL:?DATABASE_URL must point to the database being inspected}"
repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repository_root"

go run ./apps/api/cmd/migrate status
