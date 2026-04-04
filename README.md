# Orchestra

Multi-Agent Collaboration Platform - A web-based system for orchestrating multiple AI agents (Claude Code, Gemini CLI, Aider, etc.) running in parallel.

## Features

- **Multi-Agent Terminal Management**: Run multiple AI agents in parallel terminals
- **Real-time Collaboration**: Chat interface with @mentions and member management
- **Workspace Management**: Create and switch between multiple workspaces
- **Member Roles**: Support for Owner, Admin, Assistant, and Member roles
- **Modern UI**: Built with Vue 3, TypeScript, and Tailwind CSS

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

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm or pnpm

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Build
make build

# Run
make run
```

The backend will start on `http://localhost:8080`

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Development server
npm run dev

# Production build
npm run build
```

The frontend will start on `http://localhost:5173`

## Project Structure

```
Orchestra/
├── backend/           # Go backend
│   ├── cmd/           # Entry points
│   ├── internal/      # Internal modules
│   │   ├── api/       # HTTP handlers
│   │   ├── storage/   # Database layer
│   │   ├── terminal/  # PTY management
│   │   └── ws/        # WebSocket handling
│   └── configs/       # Configuration
├── frontend/          # Vue frontend
│   ├── src/
│   │   ├── app/       # App setup, router
│   │   ├── features/  # Feature modules
│   │   └── shared/    # Shared components
│   └── public/
├── docs/              # Documentation
└── README.md
```

## API Endpoints

### REST API
- `GET /health` - Health check
- `GET /api/ping` - Test connection
- `GET /api/workspaces` - List workspaces
- `POST /api/workspaces` - Create workspace
- `GET /api/workspaces/:id/members` - List members
- `POST /api/workspaces/:id/members` - Add member
- `POST /api/terminals` - Create terminal session
- `DELETE /api/terminals/:id` - Close terminal session

### WebSocket
- `GET /ws/terminal/:sessionId` - Terminal WebSocket connection

## Configuration

Configuration file: `backend/configs/config.yaml`

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
    - claude
    - gemini
  allowed_paths:
    - ~/projects
  allowed_origins:
    - "http://localhost:5173"

storage:
  database: "./data/orchestra.db"
```

## Development

### Code Standards
- Go: `gofmt`, `goimports`
- TypeScript: ESLint, Prettier
- Commits: Conventional Commits

### Running Tests

```bash
# Backend
cd backend && make test

# Frontend
cd frontend && npm run test
```

## License

MIT License

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.