#!/usr/bin/env sh
set -eu

repository_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
api_base_url=${API_BASE_URL:-http://localhost:8080}
web_base_url=${WEB_BASE_URL:-http://localhost:3000}

"$repository_root/scripts/wait-for-url.sh" "$api_base_url/health" 60
"$repository_root/scripts/wait-for-url.sh" "$api_base_url/ready" 60
"$repository_root/scripts/wait-for-url.sh" "$web_base_url/health" 60

routes=$(curl --fail --silent --show-error "$api_base_url/api/v1/routes")
case "$routes" in
  *'"items"'*) ;;
  *) echo "Route response does not contain items." >&2; exit 1 ;;
esac

echo "Full-stack smoke test passed."
