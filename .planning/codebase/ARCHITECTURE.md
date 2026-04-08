# Architecture

**Analysis Date:** 2026-04-08

## Pattern Overview

**Overall:** Client-server architecture with a Go backend (Gin HTTP + WebSocket) serving a Vue 3 SPA. The backend manages multi-agent collaboration sessions using the A2A (Agent-to-Agent) protocol, with SQLite for persistent storage.

**Key Characteristics:**
- **Gin-based REST API** with JWT + legacy token authentication
- **WebSocket gateway** for real-time terminal (A2A) and chat communication
- **Repository pattern** for data access, using `database/sql` with SQLite
- **A2A protocol** for agent communication (HTTP-based, replaces old stdin/stdout PTY approach)
- **Feature-based frontend** organized in `frontend/src/features/`
- **YAML configuration** loaded from `backend/configs/config.yaml`

## Layers

### Presentation (Frontend)
- **Purpose:** Vue 3 SPA providing workspace management, chat, terminal, member management, settings, and tasks
- **Location:** `frontend/src/`
- **Contains:** Vue components, Pinia stores, WebSocket client wrappers, API client
- **Depends on:** Axios-based API client, WebSocket classes, Vue Router
- **Used by:** Browser users

### API Layer (Backend)
- **Purpose:** HTTP request routing, authentication middleware, request/response handling
- **Location:** `backend/internal/api/`
- **Contains:** Gin router setup (`router.go`), middleware (`middleware/`), HTTP handlers (`handlers/`)
- **Depends on:** Repositories, A2A pool, WebSocket gateway, config
- **Used by:** Frontend SPA via REST and WebSocket

### WebSocket Gateway
- **Purpose:** Real-time bidirectional communication for terminal sessions and chat
- **Location:** `backend/internal/ws/`
- **Contains:** `Gateway` (connection manager), `A2ATerminalHandler` (terminal WS proxy), `ChatHandler` + `ChatHub` (chat broadcast)
- **Depends on:** A2A pool, gorilla/websocket
- **Used by:** Frontend via `/ws/terminal/:sessionId` and `/ws/chat/:workspaceId`

### A2A Protocol Layer
- **Purpose:** Agent-to-Agent communication via HTTP/SSE
- **Location:** `backend/internal/a2a/`
- **Contains:** `Pool` (session management), `Session` (individual agent sessions with SSE subscription), `AgentRegistry`, `ToolHandler`
- **Depends on:** `a2aproject/a2a-go` library
- **Used by:** WebSocket gateway, terminal handlers, conversation handlers

### Data Access (Repository)
- **Purpose:** Type-safe data access over SQLite
- **Location:** `backend/internal/storage/repository/`
- **Contains:** Interface definitions (`interface.go`) and implementations for each entity
- **Depends on:** `database/sql`, `models`
- **Used by:** API handlers

### Storage
- **Purpose:** Database connection, migrations, schema management
- **Location:** `backend/internal/storage/`
- **Contains:** `Database` wrapper (`database.go`), migration runner, migration SQL files in `migrations/`
- **Depends on:** `mattn/go-sqlite3`
- **Used by:** Repository layer

### Business Logic (Models + Handlers)
- **Purpose:** Domain entities and request processing
- **Location:** `backend/internal/models/` and `backend/internal/api/handlers/`
- **Contains:** Model structs, handler structs with CRUD logic
- **Depends on:** Repositories, A2A pool, security, filesystem
- **Used by:** API layer

### Security
- **Purpose:** Authentication, authorization, cryptography, path whitelisting
- **Location:** `backend/internal/security/`
- **Contains:** JWT generation/validation (`jwt.go`), password hashing (`password.go`), encryption (`crypto.go`), command/path whitelist (`whitelist.go`)
- **Used by:** Middleware, auth handler, API key handler

## Data Flow

### HTTP Request Flow
1. **Client** sends HTTP request to `/api/...` endpoint
2. **CORS middleware** (`middleware/cors.go`) validates origin and sets headers
3. **Logger middleware** (`middleware/logger.go`) logs request details
4. **Auth middleware** (`middleware/auth.go`) validates JWT or legacy token; sets `userId` and `username` in Gin context
5. **Handler** (`handlers/*.go`) processes request, calls repository methods
6. **Repository** executes SQL against SQLite via `database/sql`
7. **Handler** returns JSON response

### WebSocket Terminal Flow
1. **Client** connects to `/ws/terminal/:sessionId` with auth token in query param
2. **WebSocket auth middleware** validates token
3. **Gateway** (`ws/gateway.go`) upgrades HTTP to WebSocket, delegates to `A2ATerminalHandler`
4. **A2ATerminalHandler** finds session in A2A `Pool`, starts read/write loops
5. **Write loop** listens on `Session.OutputChan` (A2A events) and `Session.chatStream` channel, converts ACP messages to WS JSON, sends to client
6. **Read loop** parses client JSON messages (`user_message`, `tool_result`, `close`), calls `Session.SendUserMessage()` or `Session.SendToolResult()`
7. **A2A Session** sends messages to agent via HTTP, subscribes to SSE for streaming responses

### WebSocket Chat Flow
1. **Client** connects to `/ws/chat/:workspaceId`
2. **Gateway** delegates to `ChatHandler`
3. **ChatHandler** registers client in `GlobalChatHub` (workspace-scoped subscriptions)
4. **Hub** broadcasts events (`new_message`, `message_status`, `unread_sync`) to all clients in the workspace
5. **Conversation handler** (`handlers/conversation.go`) calls `hub.BroadcastToWorkspace()` when messages are saved

### Filesystem Browse Flow
1. **Client** requests `GET /api/workspaces/:id/browse?path=...`
2. **Workspace handler** uses `filesystem.Browser.ListDir()`
3. **Browser** calls `Validator.ValidatePath()` to check path is within allowed paths
4. **Browser** returns `FileInfo` entries (name, path, isDir, size, modTime, mode)

## Key Abstractions

### A2A Session (`backend/internal/a2a/session.go`)
- **Purpose:** Wraps a single agent communication session with A2A protocol
- **Pattern:** Manages SSE subscription lifecycle with exponential backoff reconnect, converts A2A events to ACP message format for frontend compatibility
- **Key methods:** `SendUserMessage()`, `SendToolResult()`, `subscribeToTask()`, `Release()`

### Repository Pattern (`backend/internal/storage/repository/`)
- **Purpose:** Abstract data access behind Go interfaces
- **Interface file:** `backend/internal/storage/repository/interface.go`
- **Repositories:** `WorkspaceRepository`, `MemberRepository`, `TaskRepository`, `APIKeyRepository`, plus user, conversation, message, attachment repos
- **Pattern:** Constructor takes `*sql.DB`, methods take `context.Context`, return typed results or errors

### Handler Pattern (`backend/internal/api/handlers/`)
- **Purpose:** Struct-based handlers with dependency injection
- **Pattern:** Constructor accepts repositories and external services; methods are `func(h *Handler) c *gin.Context`
- **Example:** `NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, a2aPool, chatHub)`

### ChatHub (`backend/internal/ws/chat.go`)
- **Purpose:** Pub/sub hub for chat WebSocket connections
- **Pattern:** Workspace-scoped subscriptions with `workspaceSubs map[string]map[string]struct{}`, broadcast via non-blocking `select` with drop-on-full behavior
- **Global singleton:** `var GlobalChatHub = &ChatHub{...}`

## Entry Points

### Backend Entry Point (`backend/cmd/server/main.go`)
- **Triggers:** `make run` or direct binary execution
- **Responsibilities:**
  1. Load config from YAML (`config.Load()`)
  2. Initialize SQLite database with WAL mode and foreign keys
  3. Run file-based migrations
  4. Initialize security whitelist and key encryptor
  5. Create default user if auth enabled and no users exist
  6. Initialize A2A registry and pool
  7. Create WebSocket gateway
  8. Setup Gin router with all routes and middleware
  9. Start HTTP server in goroutine
  10. Wait for SIGINT/SIGTERM for graceful shutdown

### Frontend Entry Point (`frontend/src/main.ts`)
- **Triggers:** Vite dev server or production build
- **Responsibilities:** Create Vue app, install Pinia, i18n, Router, mount to `#app`

## Error Handling

**Strategy:**
- **Backend:** Handlers return HTTP status codes with JSON `{"error": "message"}`. Panics caught by `gin.Recovery()`. WebSocket handlers log errors and close connection gracefully.
- **Frontend:** Axios interceptor catches API errors and shows toast notifications via `notifyUserError`. Configurable per-request with `skipErrorToast`. Terminal attach probe 404s are silently suppressed.

**Patterns:**
- `c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})` — standard error response
- `log.Printf("[component] error: %v", err)` — structured logging with component prefix
- `notifyUserError(label, error)` — frontend error toasts

## Cross-Cutting Concerns

**Logging:** Standard `log` package with component-prefixed format (e.g., `[ChatHub]`, `[a2a-ws]`, `[a2a-pool]`). Request logging via custom middleware.

**Validation:** Gin `binding` tags on request structs. Path validation via `filesystem.Validator` against allowed paths whitelist. Command validation via `security.Whitelist`.

**Authentication:** Dual-mode — JWT (Bearer token) and legacy token (`ORCHESTRA_AUTH_TOKEN`). Can be disabled via `ORCHESTRA_AUTH_DISABLED=true`. WebSocket auth via query parameter `?token=...`.

**CORS:** Custom middleware supporting exact origin match and wildcard subdomain (`*.example.com`). Preflight requests handled with 204 response.

**Configuration:** YAML file at `backend/configs/config.yaml` with env var overrides. Structured as `Server`, `Terminal`, `Security`, `Storage`, `Auth` sections.

**i18n:** Vue I18n with `en.json` and `zh.json` locale files. Locale controlled by settings store.

---

*Architecture analysis: 2026-04-08*
