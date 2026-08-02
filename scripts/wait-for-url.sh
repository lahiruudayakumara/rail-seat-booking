#!/usr/bin/env sh
set -eu

url=${1:?usage: wait-for-url.sh URL [TIMEOUT_SECONDS]}
timeout_seconds=${2:-60}
elapsed=0

while [ "$elapsed" -lt "$timeout_seconds" ]; do
  if curl --fail --silent --show-error --output /dev/null "$url"; then
    echo "Ready: $url"
    exit 0
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done

echo "Timed out after ${timeout_seconds}s waiting for $url" >&2
exit 1
