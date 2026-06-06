#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

output="$(
  ALL_PROXY=socks5://127.0.0.1:7897 \
  HTTPS_PROXY=socks5://127.0.0.1:7897 \
  HTTP_PROXY=http://127.0.0.1:8888 \
  NO_PROXY=example.com \
  ORCHESTRA_E2E_DRY_RUN=1 \
  "$ROOT_DIR/scripts/run-terminal-e2e.sh"
)"

if grep -q "ALL_PROXY=" <<<"$output"; then
  echo "ALL_PROXY leaked into terminal E2E environment" >&2
  exit 1
fi

if grep -q "HTTPS_PROXY=" <<<"$output"; then
  echo "HTTPS_PROXY leaked into terminal E2E environment" >&2
  exit 1
fi

if grep -q "HTTP_PROXY=" <<<"$output"; then
  echo "HTTP_PROXY leaked into terminal E2E environment" >&2
  exit 1
fi

if ! grep -q "NO_PROXY=.*example.com" <<<"$output"; then
  echo "existing NO_PROXY entries were not preserved" >&2
  exit 1
fi

if ! grep -q "NO_PROXY=.*127.0.0.1" <<<"$output" ||
  ! grep -q "NO_PROXY=.*localhost" <<<"$output" ||
  ! grep -q "NO_PROXY=.*::1" <<<"$output"; then
  echo "localhost NO_PROXY entries were not added" >&2
  exit 1
fi

echo "terminal E2E runner environment is sanitized"
