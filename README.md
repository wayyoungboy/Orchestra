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
- **Go 1.21+** - Core runtime
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
- Go 1.21+
- Node.js 18+ (pnpm recommended)

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
cd backend
make reset-data  # Deletes database and WAL files

# Optional: clear browser localStorage
# DevTools → Application → Clear site data
# Or remove keys: orchestra-settings, orchestra.auth, orchestra.user
```

### After Code Changes

Restart both dev processes after pulling changes:

1. **Backend**: Stop (`Ctrl+C`) → `make run`
2. **Frontend**: Stop (`Ctrl+C`) → `pnpm dev`

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
| `/api/ping` | GET | Test connection |
| `/api/workspaces` | GET | List all workspaces |
| `/api/workspaces` | POST | Create workspace |
| `/api/workspaces/:id` | GET | Get workspace details |
| `/api/workspaces/:id/members` | GET | List workspace members |
| `/api/workspaces/:id/members` | POST | Add member to workspace |
| `/api/workspaces/:id/members/:mid` | PUT | Update member |
| `/api/workspaces/:id/members/:mid` | DELETE | Remove member |
| `/api/terminals` | POST | Create terminal session |
| `/api/terminals/:id` | DELETE | Close terminal session |
| `/api/browse` | GET | Browse server paths |
| `/api/conversations/:workspaceId` | GET | Get chat history |
| `/api/conversations/:workspaceId/messages` | POST | Send chat message |
| `/api/tasks` | POST | Create task |
| `/api/tasks/:id` | PATCH | Update task status |
| `/api/tasks/:id/start` | POST | Start task |
| `/api/tasks/:id/complete` | POST | Complete task |
| `/api/tasks/:id/fail` | POST | Fail task |
| `/api/keys` | GET/POST/DELETE | API key management |
| `/api/keys/:id/test` | POST | Test API key |

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

## Development

### Code Standards
- Go: `gofmt`, `goimports`
- TypeScript: ESLint, Prettier
- Commits: Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)

### Running Tests

```bash
# Current MVP verification gate
./scripts/verify-mvp.sh

# Backend unit tests
cd backend && make test

# Frontend unit tests
cd frontend && pnpm test

# E2E tests (requires running backend)
cd frontend && pnpm test:e2e

# Focused agent terminal runtime E2E (requires backend and tmux)
cd frontend && pnpm test:e2e:terminal

# E2E with custom backend URL
ORCHESTRA_API_URL=http://your-server:8080 pnpm test:e2e
```

### Make Commands

```bash
# Backend Makefile targets
make build      # Build binary
make run        # Run server
make test       # Run tests
make reset-data # Clean database
make clean      # Remove build artifacts
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
