#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
API_URL="${ORCHESTRA_API_URL:-http://127.0.0.1:8080}"
BACKEND_LOG="${ORCHESTRA_E2E_BACKEND_LOG:-$ROOT_DIR/.tmp/orchestra-focused-e2e-backend.log}"
BACKEND_DB="${ORCHESTRA_E2E_BACKEND_DB:-$ROOT_DIR/.tmp/orchestra-focused-e2e.db}"
BACKEND_CONFIG="${ORCHESTRA_E2E_BACKEND_CONFIG:-$ROOT_DIR/.tmp/orchestra-focused-e2e.yaml}"

backend_pid=""

cleanup() {
  if [[ -n "$backend_pid" ]]; then
    if command -v pkill >/dev/null 2>&1; then
      pkill -TERM -P "$backend_pid" >/dev/null 2>&1 || true
    fi
    kill "$backend_pid" >/dev/null 2>&1 || true
    wait "$backend_pid" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

backend_available() {
  curl -fsS "$API_URL/health" >/dev/null 2>&1
}

wait_for_backend() {
  local i
  for i in {1..45}; do
    if backend_available; then
      return 0
    fi
    sleep 1
  done
  return 1
}

mkdir -p "$(dirname "$BACKEND_LOG")" "$(dirname "$BACKEND_DB")" "$(dirname "$BACKEND_CONFIG")"

if backend_available; then
  echo "==> Reusing backend at $API_URL"
else
  if [[ "$API_URL" != "http://127.0.0.1:8080" && "$API_URL" != "http://localhost:8080" ]]; then
    echo "Backend is not available at custom ORCHESTRA_API_URL=$API_URL; start it first or use the default local URL." >&2
    exit 1
  fi

  echo "==> Starting backend at $API_URL"
  rm -f "$BACKEND_DB" "$BACKEND_DB-wal" "$BACKEND_DB-shm" "$BACKEND_DB-journal"
  cat >"$BACKEND_CONFIG" <<EOF
server:
  http_addr: ":8080"

terminal:
  max_sessions: 32
  idle_timeout: 30m

security:
  encryption_key: ""
  allowed_commands:
    - /bin/bash
    - /bin/zsh
    - /bin/cat
    - /usr/bin/env
    - claude
    - gemini
    - codex
    - qwen
    - aider
    - cursor-agent
  allowed_paths:
    - "$ROOT_DIR"
    - /Volumes/code
    - "~"
  allowed_origins:
    - "http://localhost:5173"
    - "http://127.0.0.1:5173"

auth:
  enabled: false
  jwt_secret: ""
  jwt_expiration: 24h
  allow_registration: false

storage:
  database: "$BACKEND_DB"
  workspaces: "$ROOT_DIR/.tmp/workspaces"
EOF
  (
    cd "$BACKEND_DIR"
    ORCHESTRA_CONFIG="$BACKEND_CONFIG" make run
  ) >"$BACKEND_LOG" 2>&1 &
  backend_pid="$!"

  if ! wait_for_backend; then
    echo "Backend did not become healthy. Last backend log lines:" >&2
    tail -n 120 "$BACKEND_LOG" >&2 || true
    exit 1
  fi
fi

ORCHESTRA_API_URL="$API_URL" \
ORCHESTRA_RUN_ALL_FOCUSED_E2E=1 \
"$ROOT_DIR/scripts/verify-mvp.sh"
