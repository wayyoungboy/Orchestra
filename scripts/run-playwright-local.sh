#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"

append_no_proxy() {
  local current="${NO_PROXY:-}"
  local required=("127.0.0.1" "localhost" "::1")

  for entry in "${required[@]}"; do
    if [[ ",$current," != *",$entry,"* ]]; then
      if [[ -n "$current" ]]; then
        current="$current,$entry"
      else
        current="$entry"
      fi
    fi
  done

  export NO_PROXY="$current"
}

unset ALL_PROXY HTTPS_PROXY HTTP_PROXY all_proxy https_proxy http_proxy
append_no_proxy

if [[ "${ORCHESTRA_E2E_DRY_RUN:-}" == "1" ]]; then
  env | grep -E '^(ALL_PROXY|HTTPS_PROXY|HTTP_PROXY|NO_PROXY)=' | sort || true
  exit 0
fi

if [[ "${ORCHESTRA_SKIP_FRONTEND_BUILD:-}" != "1" ]]; then
  (
    cd "$FRONTEND_DIR"
    pnpm build
  )
fi

(
  cd "$FRONTEND_DIR"
  pnpm exec playwright test "$@" --project=chromium --reporter="${PLAYWRIGHT_REPORTER:-line}"
)
