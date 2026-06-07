#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

files="$(find "$ROOT_DIR/backend" -name '*.go' -not -path '*/vendor/*' -print)"
if [[ -z "$files" ]]; then
  exit 0
fi

unformatted="$(printf '%s\n' "$files" | xargs gofmt -l)"
if [[ -n "$unformatted" ]]; then
  echo "Go files need gofmt:" >&2
  echo "$unformatted" >&2
  echo "Run: gofmt -w <files>" >&2
  exit 1
fi

unimported="$(
  cd "$ROOT_DIR/backend"
  printf '%s\n' "$files" | xargs go run golang.org/x/tools/cmd/goimports -l
)"
if [[ -n "$unimported" ]]; then
  echo "Go files need goimports:" >&2
  echo "$unimported" >&2
  echo "Run: cd backend && go run golang.org/x/tools/cmd/goimports -w <files>" >&2
  exit 1
fi
