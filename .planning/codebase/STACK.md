# Technology Stack

**Analysis Date:** 2026-04-08

## Languages

**Primary:**
- Go 1.25.0 (per `go.mod` directive) - Backend API server, WebSocket handlers, A2A agent protocol
- TypeScript 5.3.0 - Frontend application logic, type definitions
- Vue 3.4.0 SFC (Single File Components) - Frontend UI components

**Secondary:**
- SQL (SQLite dialect) - Database migrations and queries

## Runtime

**Backend:**
- Go standard library `net/http` (via Gin)
- No separate runtime - compiled binary

**Frontend:**
- Browser (ES2020+ target via Vite)
- Node.js for build tooling

**Package Manager:**
- Go modules (`go.mod` / `go.sum`) - lockfile present
- pnpm 10.33.0 - lockfile present (`pnpm-lock.yaml`)

## Backend Frameworks & Libraries

**Core HTTP Framework:**
- [Gin](https://github.com/gin-gonic/gin) v1.12.0 - Web framework, routing, middleware
  - File: `backend/internal/api/router.go`

**WebSocket:**
- [gorilla/websocket](https://github.com/gorilla/websocket) v1.5.3 - WebSocket upgrader and connection management
  - Files: `backend/internal/ws/gateway.go`, `backend/internal/ws/chat.go`, `backend/internal/ws/a2a_terminal.go`

**Database:**
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) v1.14.37 - CGo SQLite driver
  - File: `backend/internal/storage/database.go`
  - Uses `database/sql` standard library interface (no ORM)
  - WAL journal mode enabled (`_journal_mode=WAL`)
  - Foreign keys enabled (`_foreign_keys=on`)

**Authentication:**
- [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) v5.2.2 - JWT token generation and validation (HS256 signing)
  - File: `backend/internal/security/jwt.go`
- `golang.org/x/crypto` v0.49.0 - Password hashing (bcrypt)
  - File: `backend/internal/security/password.go`

**Agent-to-Agent (A2A) Protocol:**
- [a2aproject/a2a-go/v2](https://github.com/a2aproject/a2a-go) v2.1.0 - A2A protocol implementation for inter-agent communication
  - Files: `backend/internal/a2a/pool.go`, `backend/internal/a2a/session.go`, `backend/internal/a2a/registry.go`, `backend/internal/a2a/tool_handler.go`
  - Uses JSONRPC transport over HTTP
  - SSE (Server-Sent Events) for streaming task updates

**API Documentation:**
- [swaggo/swag](https://github.com/swaggo/swag) v1.16.6 - Swagger/OpenAPI doc generator
- [swaggo/gin-swagger](https://github.com/swaggo/gin-swagger) v1.6.1 - Gin middleware for Swagger UI
- [swaggo/files](https://github.com/swaggo/files) v1.0.1 - Embedded Swagger UI assets
  - Served at `GET /swagger/*any`

**Utilities:**
- [google/uuid](https://github.com/google/uuid) v1.6.0 - UUID generation
- [oklog/ulid/v2](https://github.com/oklog/ulid) v2.1.1 - ULID generation (used via `pkg/utils`)
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) v3.0.1 - YAML config parsing

## Frontend Frameworks & Libraries

**Core:**
- [Vue](https://vuejs.org/) v3.4.0 - UI framework (Composition API)
- [vue-router](https://router.vuejs.org/) v4.3.0 - Client-side routing with history mode
  - File: `frontend/src/app/router.ts`
- [pinia](https://pinia.vuejs.org/) v2.1.0 - State management (composable stores)
  - Stores: `frontend/src/features/*/`, `frontend/src/stores/`

**UI & Styling:**
- [Tailwind CSS](https://tailwindcss.com/) v3.4.0 - Utility-first CSS framework
  - Config: `frontend/tailwind.config.js`
  - Custom theme with CSS variables for colors (`--color-primary`, etc.)
  - Custom fonts: Be Vietnam Pro (sans), JetBrains Mono (mono)
- PostCSS v8.4.0 + Autoprefixer v10.4.0 - CSS processing

**Terminal:**
- [@xterm/xterm](https://github.com/xtermjs/xterm.js) v5.3.0 - Terminal emulator
- @xterm/addon-fit v0.10.0 - Auto-resize terminal
- @xterm/addon-canvas v0.7.0 - Canvas-based renderer
- @xterm/addon-search v0.15.0 - Terminal search
- @xterm/addon-web-links v0.11.0 - Clickable links in terminal

**HTTP Client:**
- [axios](https://axios-http.com/) v1.6.0 - HTTP client with interceptors
  - File: `frontend/src/shared/api/client.ts`
  - Base URL: `/api`, timeout: 30s
  - Auto-attaches `Authorization: Bearer` token from `localStorage`

**Markdown Rendering:**
- [marked](https://marked.js.org/) v17.0.6 - Markdown parsing for chat messages

**Internationalization:**
- [vue-i18n](https://vue-i18n.intlify.dev/) v9.14.5 - i18n support
  - File: `frontend/src/i18n/index.ts`

## Build Tools

**Backend:**
- Make - Build orchestration (`backend/Makefile`)
  - `make build` - Compiles to `bin/orchestra`
  - `make run` - Runs via `go run ./cmd/server`
  - `make test` - Runs `go test -v ./...`
  - `make reset-data` - Wipes SQLite database and WAL files
  - `make clean` - Removes binary and data files
- No hot reload configured (manual restart required)

**Frontend:**
- [Vite](https://vitejs.dev/) v5.0.0 - Build tool and dev server
  - Config: `frontend/vite.config.ts`
  - Dev server port: 5173
  - Proxies `/api` and `/ws` to `http://localhost:8080`
  - Path alias: `@` resolves to `src/`
- [vue-tsc](https://github.com/vuejs/language-tools) v2.0.0 - TypeScript type checking for Vue SFCs
- [ESLint](https://eslint.org/) v8.56.0 + eslint-plugin-vue v9.20.0 - Linting
  - TypeScript parser: @typescript-eslint/parser v6.19.0

**Testing:**
- [Playwright](https://playwright.dev/) v1.58.2 - E2E testing
  - Config: `frontend/playwright.config.ts`
  - Test directory: `frontend/tests/e2e/`
  - Browser: Chromium only
  - Starts Vite preview server automatically
  - `pnpm test:e2e` - Build + run tests
  - `pnpm test:e2e:ui` - Run with Playwright UI

## Database

**Engine:** SQLite 3 (via `mattn/go-sqlite3` CGo driver)

**Location:** `./data/orchestra.db` (configurable via `storage.database` in config)

**Migration System:** Custom file-based migrations
- Directory: `backend/internal/storage/migrations/`
- Migration table: `schema_migrations` (version TEXT, applied_at INTEGER)
- Migrations applied sequentially by filename order
- No ORM - raw SQL via `database/sql`
- Current migrations:
  - `001_init.sql` - Core tables: workspaces, members, conversations, messages, workflows, settings, api_keys
  - `002_conversation_reads.sql` - Conversation read tracking
  - `003_users.sql` - User authentication table
  - `004_attachments.sql` - File attachments
  - `005_tasks.sql` - Task management
  - `006_acp_support.sql` - ACP (Agent Communication Protocol) support
  - `007_api_keys_updated_at.sql` - API key timestamps
  - `008_a2a_support.sql` - A2A agent registry and task tracking

**Schema approach:** TEXT primary keys (ULID/UUID format), INTEGER timestamps (Unix epoch), foreign keys with ON DELETE CASCADE

## Configuration

**Config Files:**
- `backend/configs/config.yaml` - YAML configuration
- Environment variable override: `ORCHESTRA_CONFIG` for custom path

**Config Loading:**
- File: `backend/internal/config/loader.go`
- Merges YAML config with environment variables
- Environment variables prefixed with `ORCHESTRA_`

**Key Environment Variables:**
- `ORCHESTRA_ENCRYPTION_KEY` - API key encryption key (32+ bytes, AES-GCM)
- `ORCHESTRA_AUTH_TOKEN` - Legacy auth token (backward compatibility)
- `ORCHESTRA_AUTH_DISABLED` - Set to `true` to bypass auth
- `ORCHESTRA_JWT_SECRET` - JWT signing secret

## Logging

**Approach:** Go standard library `log` package
- No structured logging framework
- Package-prefixed log lines (e.g., `[ChatHub]`, `[a2a-ws]`, `[a2a-pool]`)
- Files: Throughout `backend/internal/`

## Middleware

**Backend middleware** (defined in `backend/internal/api/middleware/`):
- `Logger()` - Custom request logging
- `CORS(allowedOrigins)` - Custom CORS with wildcard subdomain support
- `Auth(authConfig)` - JWT + legacy token authentication
- `WebSocketAuth(authConfig)` - WebSocket-optimized auth (query parameter support)
- `gin.Recovery()` - Panic recovery (Gin built-in)

## Platform Requirements

**Development:**
- Go 1.21+ (1.25.0 in go.mod)
- Node.js (version per pnpm lock)
- pnpm 10.33.0+
- CGo toolchain (required for go-sqlite3)
- SQLite3 development headers

**Production:**
- Compiled Go binary (no runtime dependencies beyond libc for CGo)
- Static file serving for frontend (or separate CDN)
- SQLite file persistence on disk
- Workspace directories for agent working directories

---

*Stack analysis: 2026-04-08*
