# Orchestra - Multi-Agent Collaboration Web System

## Project Overview

Orchestra is a web-based multi-agent collaboration system, supporting multiple AI agents (Claude Code, Gemini CLI, etc.) running in parallel with orchestration capabilities. Product behavior is aligned with a **reference Tauri desktop app** used internally for parity checks.

## Project Charter（设计定位）

1. **Service + Web** — Turn the **reference desktop model** into a **server-backed product** usable from the **browser**: CLIs and PTYs run on the machine hosting Orchestra; the UI is a web client (Vue), not Tauri + local IPC.
2. **Deliberate extension** — Add **workspace / working-directory switching**: browse and bind **server-side paths** per workspace so usage is not limited to a single fixed project root on the client.
3. **Everything else** — **Align with the reference desktop app** in behavior, contracts, and UX as much as possible: **port** features (via HTTP/WebSocket and server paths) rather than inventing alternate semantics. Implementation stacks differ (Go + Gin vs Rust/Tauri); the **product model** should stay recognizable to users familiar with that reference.

Chinese wording of the same charter: `docs/superpowers/specs/2026-03-29-orchestra-design.md` §1.3.

## Tech Stack

### Backend
- Go 1.21+
- Gin (HTTP framework)
- gorilla/websocket
- SQLite (storage)
- PTY (terminal emulation)

### Frontend
- Vue 3 + TypeScript
- Pinia (state management)
- Tailwind CSS
- xterm.js (terminal)

## Project Structure

```
Orchestra/
├── backend/           # Go backend
│   ├── cmd/           # Entry points
│   ├── internal/      # Internal modules
│   ├── pkg/           # Public utilities
│   └── configs/       # Configuration
├── frontend/          # Vue frontend
├── docs/              # Documentation
│   └── superpowers/
│       ├── specs/     # Design specifications
│       └── plans/     # Implementation plans
└── CLAUDE.md          # This file
```

## Development Commands

### Backend
```bash
cd backend
make build    # Build binary
make run      # Run server
make test     # Run tests
```

### Frontend
```bash
cd frontend
pnpm install  # Install dependencies
pnpm dev      # Development server
pnpm build    # Production build
pnpm test:e2e # Build + Playwright (set ORCHESTRA_API_URL if backend not on 127.0.0.1:8080)
```

### After code changes (local dev)

Pulling in fixes or editing the repo yourself: **restart both dev processes** so the running backend and browser bundle match the tree (Go has no hot reload; Vite HMR can miss edge cases such as env, proxy, or Pinia init).

1. **Backend** — stop the process (`Ctrl+C`), then from `backend` run `make run` again (or `make build` then run the binary if that is your flow).
2. **Frontend** — stop `pnpm dev` (`Ctrl+C`), then `pnpm dev` again from `frontend`.

If something still looks wrong, hard-refresh the browser or clear site data (see below).

### Clean redeploy / test without old server data

1. Stop the backend (`Ctrl+C` on `make run`).
2. From `backend`: `make reset-data` — deletes `data/orchestra.db` and SQLite `-wal` / `-shm` / journal files so the next `make run` starts with an empty database (migrations reapply).
3. Optional browser reset (avoids stale auth/settings against the new DB): remove localStorage keys `orchestra-settings`, `orchestra.auth`, `orchestra.user`, `orchestra.credentials` for your dev origin, or use DevTools → Application → Clear site data.

## Code Standards

- Go: gofmt, goimports
- TypeScript: ESLint, Prettier
- Commits: Conventional Commits

## API Endpoints

- `GET /health` - Health check
- `GET /api/ping` - Test connection
- `GET /ws/terminal/:sessionId` - Terminal WebSocket

## Members（工作区成员）

- 角色 `roleType`：`owner`、`admin`、`secretary`（**秘书**，产品侧对应参考桌面端的监工/协调语义）、`assistant`、`member`。创建成员见 `POST /api/workspaces/:id/members`。

## Configuration

Config file: `backend/configs/config.yaml`

Environment variables:
- `ORCHESTRA_ENCRYPTION_KEY` - API key encryption key (32+ bytes)
- `ORCHESTRA_CONFIG` - Custom config file path

## Reference Resources

- Design doc: `docs/superpowers/specs/2026-03-29-orchestra-design.md`
- Phase 1 plan: `docs/superpowers/plans/2026-03-29-orchestra-phase1-backend.md`