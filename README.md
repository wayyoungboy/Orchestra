# Orchestra

Multi-Agent Collaboration Platform - A web-based system for orchestrating multiple AI agents (Claude Code, Gemini CLI, Aider, etc.) running in parallel with real-time chat and workspace management.

[中文文档](README_CN.md)

## Product Direction

Orchestra's near-term roadmap is anchored on one MVP loop:

```text
Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state
```

Reference desktop behavior and Golutra specifications are useful guides, but they are not treated as full parity backlogs. Work that improves this loop has priority over broad feature porting.

Agent execution is CLI-first for the MVP: PTY + tmux is the baseline for durable local sessions and inspectable terminal state, ACP is used as a structured protocol enhancement where available, skills package agent capabilities, and A2A is deferred until the local CLI loop is stable. Member-level CLI/ACP configuration is the source of truth when creating terminal sessions, so users do not need to re-enter commands after adding an assistant.

The Members page now exposes that source of truth directly: configured assistants and secretaries show existing backend session state, can start or reuse their agent session from the member card, and failed starts surface inline instead of disappearing behind chat dispatch. The Agent Sessions page lists active workspace sessions across members, loads the current tmux pane snapshot into an xterm.js terminal surface, follows live output, propagates terminal resize events back to tmux, supports direct keystroke forwarding plus controlled line input, and can terminate stale sessions.

## Acknowledgments

The design concepts of this project were inspired by [golutra](https://github.com/golutra/golutra). Special thanks to the project and its author [seeksky](https://github.com/seekskyworld) for the inspiration. All code in this project is independently implemented.

## Features

- **Tmux-Backed Agent Sessions**: AI agent terminals run inside tmux sessions — processes survive backend restarts with automatic session recovery on startup
- **Multi-Agent Terminal Management**: Run multiple AI agent terminals in parallel, each with independent tmux sessions and PTY streams
- **Member-Level Agent Launch**: Check, start, or reuse an assistant/secretary backend session from the member card using the saved CLI/ACP command
- **Agent Session Inspection**: View active workspace agent sessions, owning members, current tmux pane snapshots, xterm.js live output with tmux resize propagation, direct keystroke forwarding, controlled line input, and stale-session termination from a dedicated navigation tab
- **ACP Support**: Structured JSON communication with AI agents for reliable message exchange
- **Provider Abstraction**: Pluggable AI provider support (Claude, Gemini) with unified command interface
- **Native Tool Calling**: AI agents can call Orchestra tools directly (task management, chat, status updates)
- **Task Management**: Full task lifecycle (create/start/complete/fail) with optimistic locking and Kanban view
- **Agent Completion Notifications**: Assistant/secretary replies create unread completion notifications and live browser toasts for conversation participants
- **Internal Chat Routing**: Secretary-to-assistant task forwarding, @mentions, and auto-forwarding of results
- **Outbox Pattern**: Reliable async message delivery with retry and dead-letter handling
- **Event Bus**: Internal pub/sub system for decoupled component communication
- **Skills System**: CLI-based skill management (install/uninstall/list) for extending AI agents
- **Real-time Collaboration Chat**: Built-in chat interface with @mentions for directing messages to specific members
- **Workspace Management**: Create and switch between multiple workspaces with configurable server-side paths
- **Member Roles**: Role-based permissions (Owner, Admin, Secretary, Assistant, Member)
- **Secretary Coordination**: Coordinator role for task distribution and multi-agent orchestration
- **Path Browser**: Browse and select server-side directories for each workspace
- **API Key Management**: Per-member API keys with encrypted storage and test endpoints
- **Modern Soft-Light Glass UI**: Clean, modern interface built with Vue 3, TypeScript, and Tailwind CSS
- **WebSocket Terminal Streaming**: Real-time terminal output via WebSocket with ANSI color support
- **Language & Theme**: English/Chinese language switching and light/dark theme toggle
- **i18n Support**: Internationalization support

## Screenshots

> Screenshots coming soon - see the live demo at `http://localhost:5173` after setup

## Tech Stack

### Backend
- **Go 1.25+** - Core runtime
- **Gin** - HTTP framework
- **gorilla/websocket** - WebSocket handling
- **SQLite** - Persistent storage
- **PTY** - Terminal emulation (via `creack/pty`)

### Frontend
- **Vue 3 + TypeScript** - UI framework
- **Pinia** - State management
- **Tailwind CSS v4** - Styling
- **xterm.js** - Terminal rendering
- **vue-i18n** - Internationalization

## Quick Start

### Prerequisites
- Go 1.25+
- Node.js 18+ (pnpm recommended)
- tmux (required for agent sessions and the MVP verification gate)

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Build
make build

# Run (starts on http://localhost:8080)
make run
```

From the repository root, the same backend server can be started with:

```bash
make backend-run
# or
make run
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
pnpm install

# Development server (starts on http://localhost:5173)
pnpm dev

# Production build
pnpm build
```

### Clean Reset (for testing)

If you need to start fresh without old data:

```bash
# Stop backend first (Ctrl+C)
make reset-data  # From the repository root; deletes database and WAL files

# Equivalent backend-only command:
# cd backend && make reset-data

# Optional: clear browser localStorage
# DevTools → Application → Clear site data
# Or remove keys: orchestra-settings, orchestra.auth, orchestra.user
```

### After Code Changes

Restart both dev processes after pulling changes:

1. **Backend**: Stop (`Ctrl+C`) → `make run` from the repository root, or `cd backend && make run`
2. **Frontend**: Stop (`Ctrl+C`) → `make frontend-dev` from the repository root, or `cd frontend && pnpm dev`

> Note: Go has no hot reload; Vite HMR may miss some edge cases.

## Project Structure

```
Orchestra/
├── backend/              # Go backend
│   ├── cmd/              # Entry points (server, cli)
│   ├── internal/         # Internal modules
│   │   ├── a2a/          # Agent session management (pool, sessions, tools)
│   │   ├── api/          # HTTP handlers & router
│   │   ├── chatbridge/   # Terminal-to-chat bridge
│   │   ├── cli/          # CLI commands (skills, providers)
│   │   ├── config/       # Configuration loader
│   │   ├── eventbus/     # Internal pub/sub system
│   │   ├── filesystem/   # Path browser service
│   │   ├── messagequeue/ # Async message queue
│   │   ├── models/       # Data models
│   │   ├── outbox/       # Reliable async delivery (outbox pattern)
│   │   ├── persist/      # Atomic JSON saver with coalescing
│   │   ├── provider/     # AI provider abstraction (Claude, Gemini)
│   │   ├── security/     # Auth, encryption, whitelist
│   │   ├── storage/      # SQLite repository + migrations
│   │   ├── supervisor/   # Session lifecycle management
│   │   ├── tmux/         # Tmux-backed session engine (persistence + recovery)
│   │   └── ws/           # WebSocket handlers (terminal, chat)
│   ├── pkg/              # Public utilities
│   ├── configs/          # Configuration files
│   └── Makefile          # Build commands
├── frontend/             # Vue frontend
│   ├── src/
│   │   ├── app/          # App setup, router, i18n
│   │   ├── assets/       # CSS, static assets
│   │   ├── features/     # Feature modules
│   │   │   ├── auth/     # Authentication
│   │   │   ├── chat/     # Chat interface
│   │   │   ├── members/  # Member management
│   │   │   ├── settings/ # Settings page
│   │   │   ├── tasks/    # Task Kanban board
│   │   │   ├── terminal/ # Terminal workspace
│   │   │   └── workspace/# Workspace selection
│   │   └── shared/       # Shared components, API, utils, bridge
│   └── public/
├── docs/                 # Documentation
│   ├── ARCHITECTURE.md   # System architecture
│   └── superpowers/      # Specs and plans
├── CLAUDE.md             # Project instructions
├── README.md             # This file
└── README_CN.md          # Chinese documentation
```

## Member Roles

| Role | Description |
|------|-------------|
| **Owner** | Full control over workspace and members |
| **Admin** | Can manage members and workspace settings |
| **Secretary** | Coordinator role (monitoring/orchestration semantics) |
| **Assistant** | Can participate in chat and use terminals |
| **Member** | Basic participant with limited permissions |

## API Endpoints

### REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/swagger/*any` | GET | Swagger UI and OpenAPI assets |
| `/api/auth/config` | GET | Read auth mode and registration settings |
| `/api/auth/login` | POST | Log in and issue an auth token |
| `/api/auth/validate` | POST | Validate an auth token |
| `/api/auth/me` | GET | Read the current authenticated user |
| `/api/auth/register` | POST | Register a user when registration is enabled |
| `/api/workspaces` | GET/POST | List or create workspaces |
| `/api/workspaces/validate-path` | POST | Validate a server-side workspace path |
| `/api/workspaces/:id` | GET/PUT/DELETE | Read, update, or delete a workspace |
| `/api/browse` | GET | Browse server paths |
| `/api/workspaces/:id/browse` | GET | Browse paths for a workspace |
| `/api/workspaces/:id/search` | GET | Search paths inside a workspace |
| `/api/workspaces/:id/members` | GET/POST | List or add workspace members |
| `/api/workspaces/:id/members/:memberId` | GET/PUT/DELETE | Read, update, or remove a member |
| `/api/workspaces/:id/members/:memberId/conversations` | DELETE | Delete conversations for a member |
| `/api/workspaces/:id/members/:memberId/presence` | POST | Update member presence/activity |
| `/api/workspaces/:id/members/:memberId/terminal-session` | GET/POST | Check, start, or reuse a member agent session |
| `/api/workspaces/:id/terminal-sessions` | GET | List workspace agent sessions |
| `/api/terminals` | POST | Create terminal session |
| `/api/terminals/:sessionId/snapshot` | GET | Read current terminal pane snapshot |
| `/api/terminals/:sessionId` | DELETE | Close terminal session |
| `/api/workspaces/:id/conversations` | GET/POST | List or create conversations |
| `/api/workspaces/:id/conversations/:convId` | GET/PUT/DELETE | Read, update, or delete a conversation |
| `/api/workspaces/:id/conversations/direct` | POST | Create or reuse a direct message |
| `/api/workspaces/:id/conversations/:convId/members` | PUT | Replace conversation membership |
| `/api/workspaces/:id/conversations/:convId/messages` | GET/POST | List or send conversation messages |
| `/api/workspaces/:id/conversations/:convId/messages` | DELETE | Clear conversation messages |
| `/api/workspaces/:id/conversations/:convId/messages/:messageId` | DELETE | Delete a single message |
| `/api/workspaces/:id/conversations/:convId/read` | POST | Mark a conversation read |
| `/api/workspaces/:id/conversations/read-all` | POST | Mark all workspace conversations read |
| `/api/internal/chat/send` | POST | Internal AI result message API |
| `/api/internal/agent-status` | POST | Internal agent status update API |
| `/api/internal/tasks/create` | POST | Create task from an agent |
| `/api/internal/tasks/assign` | POST | Assign a task to an agent |
| `/api/internal/tasks/start` | POST | Start task from an agent |
| `/api/internal/tasks/complete` | POST | Complete task from an agent |
| `/api/internal/tasks/fail` | POST | Fail task from an agent |
| `/api/internal/tasks/cancel` | POST | Cancel task from an agent |
| `/api/internal/tasks/list` | GET | List tasks for a secretary |
| `/api/internal/workloads/list` | GET | List agent workloads |
| `/api/workspaces/:id/tasks` | GET | List workspace tasks |
| `/api/workspaces/:id/tasks/:taskId` | GET | Get task details |
| `/api/workspaces/:id/tasks/my-tasks` | GET | List tasks assigned to the current agent |
| `/api/workspaces/:id/tasks/:taskId/cancel` | POST | Cancel task |
| `/api/workspaces/:id/attachments` | GET | List workspace attachments |
| `/api/workspaces/:id/conversations/:convId/attachments` | POST | Upload a conversation attachment |
| `/api/workspaces/:id/attachments/:attachmentId` | GET/DELETE | Download or delete an attachment |
| `/api/workspaces/:id/attachments/:attachmentId/info` | GET | Read attachment metadata |
| `/api/api-keys` | GET/POST | List or create API keys |
| `/api/api-keys/provider/:provider` | GET | Get the API key for a provider |
| `/api/api-keys/:id` | DELETE | Delete API key |
| `/api/api-keys/test` | POST | Test API key |
| `/api/workspaces/:id/notifications` | GET | List workspace notifications |
| `/api/workspaces/:id/notifications/badge` | GET | Read notification badge counts |
| `/api/workspaces/:id/notifications/:notifId/read` | POST | Mark a notification read |
| `/api/workspaces/:id/notifications/read-all` | POST | Mark all workspace notifications read |
| `/api/workspaces/:id/outbox` | GET | List workspace outbox delivery diagnostics |

### WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws/terminal/:sessionId` | Terminal I/O stream |
| `/ws/chat/:workspaceId` | Chat message stream |

## Configuration

Configuration file: `backend/configs/config.yaml`

Environment variables:
- `ORCHESTRA_ENCRYPTION_KEY` - API key encryption key (32+ bytes)
- `ORCHESTRA_CONFIG` - Custom config file path

### Default Configuration

```yaml
server:
  http_addr: ":8080"

terminal:
  max_sessions: 10
  idle_timeout: 30m

security:
  allowed_commands:
    - /bin/bash
    - /bin/zsh
    - /bin/cat       # Local smoke-test agent
    - claude        # Claude Code CLI
    - gemini        # Gemini CLI
    - aider         # Aider
  allowed_paths:
    - ~/projects    # Restrict path browsing
  allowed_origins:
    - "http://localhost:5173"

storage:
  database: "./data/orchestra.db"
```

`security.allowed_commands` is enforced when a member agent session is started from `/api/terminals` or `/api/workspaces/:id/members/:memberId/terminal-session`. Saved member configuration can still be edited, but disallowed commands are rejected before tmux startup and the Members page shows the backend error inline on the member card.

## Development

### Code Standards
- Go: `gofmt`, `goimports`
- TypeScript: ESLint, Prettier
- Commits: Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)

### Running Tests

```bash
# Standard root-level shortcuts
make verify
make verify-focused

# Current MVP verification gate (backend tests, frontend build/unit tests, and focused spec typecheck)
./scripts/verify-mvp.sh

# Start a temporary backend and run every focused browser MVP E2E locally
./scripts/run-focused-e2e-local.sh

# On GitHub Actions, run the CI workflow manually with
# "Run the full focused browser MVP E2E gate" enabled to execute make verify-focused remotely.

# Include every focused browser MVP E2E in one gate (requires backend and tmux)
ORCHESTRA_RUN_ALL_FOCUSED_E2E=1 ./scripts/verify-mvp.sh

# Include the first-screen workspace onboarding browser flow in the gate (requires backend)
ORCHESTRA_RUN_WORKSPACE_ONBOARDING_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser terminal E2E in the gate (requires backend and tmux)
ORCHESTRA_RUN_TERMINAL_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP chat flow in the gate (requires backend)
ORCHESTRA_RUN_MVP_CHAT_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP direct-message flow in the gate (requires backend and tmux)
ORCHESTRA_RUN_MVP_DM_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP unread sync flow in the gate (requires backend)
ORCHESTRA_RUN_MVP_UNREAD_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP agent-completion notification flow in the gate (requires backend)
ORCHESTRA_RUN_MVP_NOTIFICATION_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP dispatch diagnostics flow in the gate (requires backend and tmux)
ORCHESTRA_RUN_MVP_DISPATCH_DIAGNOSTICS_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser member session flow in the gate (requires backend and tmux)
ORCHESTRA_RUN_MVP_MEMBER_SESSION_E2E=1 ./scripts/verify-mvp.sh

# Include the focused browser MVP task flow in the gate (requires backend)
ORCHESTRA_RUN_MVP_TASK_E2E=1 ./scripts/verify-mvp.sh

# Backend unit tests
cd backend && make test

# Backend Go formatting check
make backend-format-check

# Backend static analysis
make backend-vet

# Focused backend terminal API runtime smoke (requires tmux)
cd backend && go test ./internal/api -run TestTerminalRuntimeAPIWorkspaceMemberSessionLifecycle -count=1

# Focused backend result-return loop (requires tmux)
cd backend && go test ./internal/api -run TestAssistantResultCompletesTaskAndForwardsToSecretary -count=1

# Frontend unit tests
cd frontend && pnpm test

# Frontend lint check
cd frontend && pnpm lint:check

# E2E tests (requires running backend)
cd frontend && pnpm test:e2e

# Focused MVP chat browser flow (requires running backend)
cd frontend && pnpm test:e2e:mvp-chat

# Focused MVP direct-message browser flow (requires running backend and tmux)
cd frontend && pnpm test:e2e:mvp-dm

# Focused MVP unread sync browser flow (requires running backend)
cd frontend && pnpm test:e2e:mvp-unread

# Focused MVP agent-completion notification browser flow (requires running backend)
cd frontend && pnpm test:e2e:mvp-notification

# Focused MVP dispatch diagnostics browser flow (requires running backend and tmux)
cd frontend && pnpm test:e2e:mvp-dispatch-diagnostics

# Focused member-card agent session browser flow (requires backend and tmux)
cd frontend && pnpm test:e2e:mvp-member-session

# Focused MVP task browser flow (requires running backend)
cd frontend && pnpm test:e2e:mvp-task

# Focused agent terminal runtime E2E (requires backend and tmux)
cd frontend && pnpm test:e2e:terminal

# E2E with custom backend URL
ORCHESTRA_API_URL=http://your-server:8080 pnpm test:e2e
```

The focused E2E runners clear inherited HTTP/SOCKS proxy variables for local browser/API traffic and append `127.0.0.1`, `localhost`, and `::1` to `NO_PROXY`, so localhost verification is not affected by a global development proxy.

### Make Commands

```bash
# Root Makefile targets
make verify           # Backend tests, frontend build/unit tests, focused spec typecheck
make verify-focused   # Temporary backend + all focused browser MVP E2E
make run              # Start backend API server (root alias)
make test             # Run backend tests (root alias)
make build            # Build backend server binary (root alias)
make reset-data       # Reset backend SQLite data (root alias)
make backend-run      # Start backend API server
make backend-test     # Run backend tests
make backend-reset    # Reset backend SQLite data
make backend-format-check # Check backend Go formatting
make backend-vet      # Run go vet for the backend
make frontend-install # Install frontend dependencies
make frontend-dev     # Start frontend dev server
make frontend-build   # Build frontend
make frontend-test    # Run frontend unit tests
make frontend-lint    # Run frontend lint checks

# Backend-only Makefile targets from ./backend (also reachable through the root aliases above)
make build            # Build binary
make run              # Run server
make test             # Run tests
make reset-data       # Clean database
make clean            # Remove build artifacts
```

## Roadmap

- [ ] Member presence indicators (real-time)
- [ ] Workspace templates
- [ ] Export chat transcripts
- [ ] E2E encryption for inter-agent messages

## License

MIT License

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Follow code standards (gofmt, ESLint)
4. Write tests for new features
5. Submit a PR with clear description

---

Built with ❤️ for AI-assisted development workflows.
