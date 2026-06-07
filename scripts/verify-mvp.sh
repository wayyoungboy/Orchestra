#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Backend tests, including tmux runtime smoke"
(
  cd "$ROOT_DIR/backend"
  go test ./...
)

echo "==> Backend gofmt check"
"$ROOT_DIR/scripts/check-go-format.sh"

echo "==> Backend vet"
(
  cd "$ROOT_DIR/backend"
  go vet ./...
)

echo "==> Frontend production build"
(
  cd "$ROOT_DIR/frontend"
  pnpm build
)

echo "==> Frontend unit tests"
(
  cd "$ROOT_DIR/frontend"
  pnpm test
)

echo "==> Frontend lint check"
(
  cd "$ROOT_DIR/frontend"
  pnpm lint:check
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
    e2e/env.d.ts \
    e2e/*.spec.ts
)

echo "==> Focused Playwright E2E runner environment"
"$ROOT_DIR/scripts/test-terminal-e2e-runner-env.sh"

run_all_focused_e2e="${ORCHESTRA_RUN_ALL_FOCUSED_E2E:-}"

if [[ "$run_all_focused_e2e" == "1" || "${ORCHESTRA_RUN_MVP_CHAT_E2E:-}" == "1" ]]; then
  echo "==> MVP chat flow E2E"
  ORCHESTRA_SKIP_FRONTEND_BUILD=1 "$ROOT_DIR/scripts/run-mvp-chat-e2e.sh"
fi

if [[ "$run_all_focused_e2e" == "1" || "${ORCHESTRA_RUN_MVP_MEMBER_SESSION_E2E:-}" == "1" ]]; then
  echo "==> MVP member session flow E2E"
  ORCHESTRA_SKIP_FRONTEND_BUILD=1 "$ROOT_DIR/scripts/run-mvp-member-session-e2e.sh"
fi

if [[ "$run_all_focused_e2e" == "1" || "${ORCHESTRA_RUN_MVP_TASK_E2E:-}" == "1" ]]; then
  echo "==> MVP task flow E2E"
  ORCHESTRA_SKIP_FRONTEND_BUILD=1 "$ROOT_DIR/scripts/run-mvp-task-e2e.sh"
fi

if [[ "$run_all_focused_e2e" == "1" || "${ORCHESTRA_RUN_TERMINAL_E2E:-}" == "1" ]]; then
  echo "==> Agent terminal runtime E2E"
  ORCHESTRA_SKIP_FRONTEND_BUILD=1 "$ROOT_DIR/scripts/run-terminal-e2e.sh"
fi

echo "==> MVP verification passed"
