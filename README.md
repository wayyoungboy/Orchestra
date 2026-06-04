# Orchestra

**A web cockpit for running Claude Code, Gemini CLI, Aider, and other coding agents together.**

Orchestra helps you coordinate multiple terminal-based AI agents from one browser workspace: live terminal streaming, team chat, task tracking, agent-to-agent communication, and role-based coordination.

[中文文档](README_CN.md) · [Launch plan](docs/LAUNCH_PLAN.md) · [Demo script](docs/DEMO_SCRIPT.md)

---

## Why Orchestra?

Most coding agents are powerful, but they still work like isolated terminals. Once you use more than one agent, the workflow gets messy:

- Claude Code is in one terminal.
- Gemini CLI or Aider is in another.
- Task state lives in your head or a separate board.
- Handoffs happen through copy-paste.
- Useful context is hard to coordinate across agents.

Orchestra turns that into one coordinated cockpit.

> Start multiple agents, route messages through chat, coordinate handoffs, and track implementation work from a shared workspace.

## What you can do

| Need | Orchestra provides |
| --- | --- |
| Run multiple coding agents | Claude Code, Gemini CLI, Aider, and configurable providers |
| Watch work in real time | WebSocket terminal streaming with ANSI output |
| Coordinate handoffs | Workspace chat, @mentions, and Secretary-to-assistant routing |
| Track execution | Task lifecycle and Kanban-style workflow |
| Let agents talk to each other | A2A protocol, tool calling, and structured JSON messaging |
| Prevent agent loops | Message depth tracking and parent references |
| Manage teams/workspaces | Roles, API keys, path browser, and workspace settings |

## Demo

A short demo is the most important next asset for this project.

- Recording guide: [`docs/DEMO_SCRIPT.md`](docs/DEMO_SCRIPT.md)
- Recommended GIF path after recording: `docs/assets/orchestra-demo.gif`

Suggested 60-second demo flow:

1. Open one workspace in Orchestra.
2. Start Claude Code, Gemini CLI, and Aider sessions.
3. Create a task and assign it through chat.
4. Show live terminal output from one agent.
5. Route the result to another agent for review.
6. Move the task across the board.

## Features

- **Multi-agent terminal management**: Run multiple AI agent terminals in parallel, each with independent PTY sessions.
- **A2A protocol**: Agent-to-agent communication with tool calling, task delegation, and structured JSON messaging.
- **ACP support**: Structured JSON communication with AI agents over stdin/stdout for reliable message exchange.
- **Native tool calling**: Agents can call Orchestra tools directly for task management, chat, and status updates.
- **Task management**: Create, start, complete, fail, and track work with optimistic locking and a Kanban view.
- **Internal chat routing**: Secretary-to-assistant task forwarding, @mentions, and auto-forwarding of results.
- **Loop detection**: Message depth tracking and parent references to prevent infinite agent loops.
- **Real-time collaboration chat**: Built-in chat interface with @mentions for directing messages to specific members.
- **Workspace management**: Create and switch between multiple workspaces with configurable server-side paths.
- **Member roles**: Owner, Admin, Secretary, Assistant, and Member roles.
- **Path browser**: Browse and select server-side directories for each workspace.
- **API key management**: Per-member API keys with encrypted storage and test endpoints.
- **Modern UI**: Vue 3, TypeScript, Tailwind CSS, language switching, and light/dark theme support.

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm recommended
- Optional: Claude Code, Gemini CLI, or Aider installed locally if you want to run those providers

### 1. Start the backend

```bash
cd backend

go mod download
make build
make run
```

Backend starts at `http://localhost:8080`.

### 2. Start the frontend

```bash
cd frontend

pnpm install
pnpm dev
```

Frontend starts at `http://localhost:5173`.

### 3. Open Orchestra

Visit:

```text
http://localhost:5173
```

Create a workspace, choose a server-side project path, and start agent terminals from the browser.

## Tech Stack

### Backend

- **Go 1.21+** - Core runtime
- **Gin** - HTTP framework
- **gorilla/websocket** - WebSocket handling
- **SQLite** - Persistent storage
- **PTY** - Terminal emulation via `creack/pty`

### Frontend

- **Vue 3 + TypeScript** - UI framework
- **Pinia** - State management
- **Tailwind CSS v4** - Styling
- **xterm.js** - Terminal rendering
- **vue-i18n** - Internationalization

## Project Structure

```text
Orchestra/
├── backend/              # Go backend
│   ├── cmd/              # Entry points
│   ├── internal/         # Internal modules
│   │   ├── a2a/          # Agent-to-Agent protocol
│   │   ├── acp/          # ACP protocol implementation
│   │   ├── api/          # HTTP handlers and router
│   │   ├── chatbridge/   # Terminal-to-chat bridge
│   │   ├── config/       # Configuration loader
│   │   ├── filesystem/   # Path browser service
│   │   ├── models/       # Data models
│   │   ├── storage/      # SQLite repository layer and migrations
│   │   ├── terminal/     # PTY management
│   │   └── ws/           # WebSocket handlers
├── frontend/             # Vue frontend
├── docs/                 # Documentation
│   ├── ARCHITECTURE.md
│   ├── ACP-INTEGRATION.md
│   ├── DEMO_SCRIPT.md
│   ├── LAUNCH_PLAN.md
│   └── superpowers/
├── CLAUDE.md
├── README.md
└── README_CN.md
```

## Member Roles

| Role | Description |
| --- | --- |
| **Owner** | Full control over workspace and members |
| **Admin** | Can manage members and workspace settings |
| **Secretary** | Coordinator role for monitoring and orchestration |
| **Assistant** | Can participate in chat and use terminals |
| **Member** | Basic participant with limited permissions |

## Configuration

Configuration file:

```text
backend/configs/config.yaml
```

Environment variables:

- `ORCHESTRA_ENCRYPTION_KEY` - API key encryption key, 32+ bytes
- `ORCHESTRA_CONFIG` - Custom config file path

Default allowed commands include common shells and coding agents:

```yaml
security:
  allowed_commands:
    - /bin/bash
    - /bin/zsh
    - claude
    - gemini
    - aider
  allowed_paths:
    - ~/projects
  allowed_origins:
    - "http://localhost:5173"
```

## API Overview

### REST API

| Endpoint | Method | Description |
| --- | --- | --- |
| `/health` | GET | Health check |
| `/api/ping` | GET | Test connection |
| `/api/workspaces` | GET/POST | List or create workspaces |
| `/api/workspaces/:id` | GET | Get workspace details |
| `/api/workspaces/:id/members` | GET/POST | List or add workspace members |
| `/api/workspaces/:id/members/:mid` | PUT/DELETE | Update or remove a member |
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
| --- | --- |
| `/ws/terminal/:sessionId` | Terminal I/O stream |
| `/ws/chat/:workspaceId` | Chat message stream |

## Development

### Code standards

- Go: `gofmt`, `goimports`
- TypeScript: ESLint, Prettier
- Commits: Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)

### Running tests

```bash
# Backend unit tests
cd backend && make test

# Frontend unit tests
cd frontend && pnpm test

# E2E tests, requires running backend
cd frontend && pnpm test:e2e
```

## Launch Roadmap

See [`docs/LAUNCH_PLAN.md`](docs/LAUNCH_PLAN.md) for a focused launch checklist.

Near-term priorities:

- [ ] Add a short demo GIF to the README hero area.
- [ ] Add a one-command local setup path.
- [ ] Add sample workflows for Claude Code, Gemini CLI, and Aider.
- [ ] Add more screenshots for terminal sessions, chat routing, and task board.

## Acknowledgments

The design concepts of this project were inspired by [golutra](https://github.com/golutra/golutra). Special thanks to the project and its author [seeksky](https://github.com/seekskyworld) for the inspiration. All code in this project is independently implemented.

## Contributing

Contributions are welcome. Please:

1. Fork the repository.
2. Create a feature branch.
3. Follow code standards.
4. Write tests for new features.
5. Submit a PR with a clear description.

## License

MIT License

---

Built with ❤️ for AI-assisted development workflows.
