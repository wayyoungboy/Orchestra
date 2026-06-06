#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Backend tests, including tmux runtime smoke"
(
  cd "$ROOT_DIR/backend"
  go test ./...
)

echo "==> Frontend production build"
(
  cd "$ROOT_DIR/frontend"
  pnpm build
)

echo "==> Agent terminal Playwright spec typecheck"
(
  cd "$ROOT_DIR/frontend"
  pnpm exec tsc \
    --noEmit \
    --skipLibCheck \
    --moduleResolution bundler \
    --module ESNext \
    --target ES2020 \
    --lib ES2020,DOM \
    e2e/agent-terminal-runtime.spec.ts
)

echo "==> MVP verification passed"
