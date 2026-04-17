# Orchestra

Multi-Agent Collaboration Platform - A web-based system for orchestrating multiple AI agents (Claude Code, Gemini CLI, Aider, etc.) running in parallel with real-time chat and workspace management.

[中文文档](README_CN.md)

## Acknowledgments

The design concepts of this project were inspired by [golutra](https://github.com/golutra/golutra). Special thanks to the project and its author [seeksky](https://github.com/seekskyworld) for the inspiration. All code in this project is independently implemented.

## Features

- **Multi-Agent Terminal Management**: Run multiple AI agent terminals in parallel, each with independent PTY sessions
- **A2A Protocol**: Agent-to-Agent communication with tool calling, task delegation, and structured JSON messaging
- **ACP Support**: Structured JSON communication with AI agents (stdin/stdout) for reliable message exchange
- **Native Tool Calling**: AI agents can call Orchestra tools directly (task management, chat, status updates)
- **Task Management**: Full task lifecycle (create/start/complete/fail) with optimistic locking and Kanban view
- **Internal Chat Routing**: Secretary-to-assistant task forwarding, @mentions, and auto-forwarding of results
- **Loop Detection**: Message depth tracking and parent references to prevent infinite agent loops
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
│   ├── cmd/              # Entry points (main.go)
│   ├── internal/         # Internal modules
│   │   ├── a2a/          # Agent-to-Agent protocol (runner, pool, sessions, tools)
│   │   ├── acp/          # ACP protocol implementation
│   │   ├── api/          # HTTP handlers & router
│   │   ├── chatbridge/   # Terminal-to-chat bridge
│   │   ├── config/       # Configuration loader
│   │   ├── filesystem/   # Path browser service
│   │   ├── models/       # Data models
│   │   ├── storage/      # SQLite repository layer + migrations
│   │   ├── terminal/     # PTY management
│   │   └── ws/           # WebSocket handlers
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
│   │   │   ├── terminal/ # Terminal workspace
│   │   │   └── workspace/# Workspace selection
│   │   └── shared/       # Shared components, API, utils
│   └── public/
├── docs/                 # Documentation
│   ├── ARCHITECTURE.md   # System architecture
│   ├── ACP-INTEGRATION.md# ACP protocol docs
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
# Backend unit tests
cd backend && make test

# Frontend unit tests
cd frontend && pnpm test

# E2E tests (requires running backend)
cd frontend && pnpm test:e2e

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

- [ ] Terminal session persistence and reconnection
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