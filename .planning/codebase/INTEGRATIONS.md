# External Integrations

**Analysis Date:** 2026-04-08

## HTTP API Routes

**Base URL:** `http://localhost:8080` (default, configurable via `server.http_addr`)

**Router:** `backend/internal/api/router.go`

### Health & Infrastructure

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | None | Health check |
| GET | `/swagger/*any` | None | Swagger API documentation |

### Authentication (`/api/auth/`)

| Method | Path | Auth | Handler |
|--------|------|------|---------|
| GET | `/api/auth/config` | None | `GetAuthConfig` - Returns whether auth is enabled |
| POST | `/api/auth/login` | None | `Login` - Username/password authentication, returns JWT |
| POST | `/api/auth/validate` | None | `ValidateToken` - Validates a JWT token |
| GET | `/api/auth/me` | JWT | `GetCurrentUser` - Returns authenticated user info |
| POST | `/api/auth/register` | None | `Register` - Creates new user (only if `allow_registration: true`) |

Handler file: `backend/internal/api/handlers/auth.go`

### Workspaces (`/api/workspaces/`)

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/workspaces` | `List` - List all workspaces |
| POST | `/api/workspaces` | `Create` - Create new workspace |
| GET | `/api/workspaces/:id` | `Get` - Get workspace details |
| PUT | `/api/workspaces/:id` | `Update` - Update workspace |
| DELETE | `/api/workspaces/:id` | `Delete` - Delete workspace |
| GET | `/api/workspaces/:id/browse` | `Browse` - Browse filesystem at workspace path |
| GET | `/api/workspaces/:id/search` | `Search` - Search workspace path |
| POST | `/api/workspaces/validate-path` | `ValidatePath` - Validate a server path |
| GET | `/api/browse` | `BrowseRoot` - Browse server root |

Handler file: `backend/internal/api/handlers/workspace.go`

### Members (`/api/workspaces/:id/members/`)

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/workspaces/:id/members` | `List` - List members in workspace |
| GET | `/api/workspaces/:id/members/:memberId` | `Get` - Get member details |
| POST | `/api/workspaces/:id/members` | `Create` - Add new member (agent/CLI) |
| PUT | `/api/workspaces/:id/members/:memberId` | `Update` - Update member config |
| DELETE | `/api/workspaces/:id/members/:memberId` | `Delete` - Remove member |
| DELETE | `/api/workspaces/:id/members/:memberId/conversations` | `DeleteConversationsForMember` |
| POST | `/api/workspaces/:id/members/:memberId/presence` | `UpdatePresence` |
| GET | `/api/workspaces/:id/members/:memberId/terminal-session` | `GetSessionForMember` |
| POST | `/api/workspaces/:id/members/:memberId/terminal-session` | `GetOrCreateSessionForMember` |

Handler files: `backend/internal/api/handlers/member.go`, `backend/internal/api/handlers/terminal.go`

### Terminal Sessions

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/terminals` | `CreateSession` - Create new terminal session for member |
| DELETE | `/api/terminals/:sessionId` | `DeleteSession` - Close terminal session |
| GET | `/api/workspaces/:id/terminal-sessions` | `ListWorkspaceTerminalSessions` |

Handler file: `backend/internal/api/handlers/terminal.go`

### Conversations (`/api/workspaces/:id/conversations/`)

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/workspaces/:id/conversations` | `List` - List conversations |
| GET | `/api/workspaces/:id/conversations/:convId` | `GetConversation` |
| POST | `/api/workspaces/:id/conversations` | `Create` - New conversation |
| PUT | `/api/workspaces/:id/conversations/:convId` | `UpdateSettings` |
| DELETE | `/api/workspaces/:id/conversations/:convId` | `Delete` |
| DELETE | `/api/workspaces/:id/conversations/:convId/messages` | `ClearMessages` |
| DELETE | `/api/workspaces/:id/conversations/:convId/messages/:messageId` | `DeleteMessage` |
| GET | `/api/workspaces/:id/conversations/:convId/messages` | `GetMessages` |
| POST | `/api/workspaces/:id/conversations/:convId/messages` | `SendMessage` - Send message (triggers A2A agent) |
| POST | `/api/workspaces/:id/conversations/:convId/read` | `MarkConversationRead` |
| POST | `/api/workspaces/:id/conversations/read-all` | `MarkAllConversationsRead` |
| PUT | `/api/workspaces/:id/conversations/:convId/members` | `SetConversationMembers` |

Handler file: `backend/internal/api/handlers/conversation.go`

### Internal API (AI Assistant Integration)

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/internal/chat/send` | `InternalChatSend` - AI agent sends messages to conversation |
| POST | `/api/internal/agent-status` | `UpdateAgentStatus` - AI agent reports status |

### Task API (Secretary Coordination)

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/internal/tasks/create` | `CreateTask` |
| POST | `/api/internal/tasks/start` | `StartTask` |
| POST | `/api/internal/tasks/complete` | `CompleteTask` |
| POST | `/api/internal/tasks/fail` | `FailTask` |
| GET | `/api/internal/workloads/list` | `ListWorkloads` |
| GET | `/api/workspaces/:id/tasks` | `ListTasks` |
| GET | `/api/workspaces/:id/tasks/:taskId` | `GetTask` |
| GET | `/api/workspaces/:id/tasks/my-tasks` | `GetMyTasks` |

Handler file: `backend/internal/api/handlers/task.go`

### Attachments

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/workspaces/:id/attachments` | `ListAttachments` |
| POST | `/api/workspaces/:id/conversations/:convId/attachments` | `UploadAttachment` |
| GET | `/api/workspaces/:id/attachments/:attachmentId` | `DownloadAttachment` |
| GET | `/api/workspaces/:id/attachments/:attachmentId/info` | `GetAttachmentInfo` |
| DELETE | `/api/workspaces/:id/attachments/:attachmentId` | `DeleteAttachment` |

Handler file: `backend/internal/api/handlers/attachment.go`

### API Keys

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/api-keys` | `List` - List stored API keys |
| GET | `/api/api-keys/provider/:provider` | `GetByProvider` |
| POST | `/api/api-keys` | `Create` - Store new API key (encrypted) |
| DELETE | `/api/api-keys/:id` | `Delete` |
| POST | `/api/api-keys/test` | `Test` - Test API key connectivity |

Handler file: `backend/internal/api/handlers/api_key.go`

## WebSocket Endpoints

**Gateway:** `backend/internal/ws/gateway.go`

### Terminal WebSocket

- **Endpoint:** `GET /ws/terminal/:sessionId`
- **Auth:** JWT token or legacy token via query parameter `?token=...`
- **Purpose:** Bidirectional communication with A2A agent sessions
- **Message protocol:** JSON

**Client-side:** `frontend/src/shared/socket/terminal.ts` - `TerminalSocket` class

**Client messages (ws -> server):**
```typescript
{ type: "input", data: string }        // User input to agent
{ type: "resize", cols: number, rows: number }  // Terminal resize
{ type: "close" }                       // Close session
```

**Server messages (server -> ws):**
```typescript
{ type: "connected", sessionId: string }
{ type: "assistant_message", content: string }
{ type: "tool_use", toolName: string, toolInput: any, toolUseID: string }
{ type: "result", message: string, costUSD: number, durationMs: number }
{ type: "error", error: string }
{ type: "status", status: string }
{ type: "exit", code: number }
```

**Handler:** `backend/internal/ws/a2a_terminal.go` - `A2ATerminalHandler`

### Chat WebSocket

- **Endpoint:** `GET /ws/chat/:workspaceId`
- **Auth:** JWT token or legacy token via query parameter `?token=...`
- **Purpose:** Real-time chat message broadcasting across workspace clients
- **Message protocol:** JSON

**Client-side:** `frontend/src/shared/socket/chat.ts` - `ChatSocket` class (singleton via `getChatSocket()`)

**Server events (server -> ws):**
```typescript
{
  type: "new_message" | "message_status" | "unread_sync",
  workspaceId: string,
  conversationId?: string,
  messageId?: string,
  senderId?: string,
  senderName?: string,
  content?: string,
  createdAt?: number,
  isAi?: boolean,
  status?: string,
  unreadCount?: number
}
```

**Hub:** `backend/internal/ws/chat.go` - `ChatHub` (global singleton) manages pub/sub by workspace

**Heartbeat:** Client sends `{ type: "ping" }` every 30 seconds; server responds with WebSocket Ping frames every 30 seconds

## A2A (Agent-to-Agent) Protocol Integration

**What it is:** A2A replaces the legacy stdin/stdout PTY approach for agent communication. Agents run as independent HTTP services that communicate via JSONRPC over HTTP with SSE streaming.

**Key files:**
- `backend/internal/a2a/pool.go` - Manages A2A session lifecycle
- `backend/internal/a2a/session.go` - Individual agent session (send messages, subscribe to SSE)
- `backend/internal/a2a/registry.go` - Agent URL registry
- `backend/internal/a2a/tool_handler.go` - Tool execution handler
- `backend/internal/a2a/messages.go` - Message type definitions

**How it works:**
1. Member has `a2a_agent_url` configured (stored in SQLite `members` table)
2. `A2A.Pool.Acquire()` creates an `a2aclient.Client` for the agent
3. Client resolves agent card via `agentcard.NewResolver` or creates client directly from endpoint
4. Messages sent via `Client.SendMessage()` with `a2a.NewMessage(a2a.MessageRoleUser, ...)`
5. Long-running tasks return `a2a.Task` - server subscribes to SSE stream for updates
6. SSE events converted to internal `ACPMessage` format and relayed via WebSocket
7. Tool use results sent back via `SendToolResult()`

**Supported agent types:** Claude Code, Gemini CLI, Codex, Qwen, Aider, Cursor-Agent (configured in `security.allowed_commands` in `config.yaml`)

## PTY / Terminal Integration

**Current approach:** A2A protocol (HTTP-based) replaces direct PTY management. The old PTY files (`backend/internal/terminal/`) have been deleted.

**Session lifecycle:**
- Sessions created via `POST /api/terminals` or `POST /api/workspaces/:id/members/:memberId/terminal-session`
- Sessions tracked in `A2A.Pool` with idle timeout cleanup (default 30 minutes)
- Max sessions configurable via `terminal.max_sessions` (default 10)
- Sessions released via `DELETE /api/terminals/:sessionId` or WebSocket `close` message

**File:** `backend/internal/api/handlers/terminal.go`

## File System Access

**What it does:** Server-side filesystem browsing scoped to workspace paths.

**Key files:**
- `backend/internal/filesystem/browser.go` - Directory listing, path validation
- `backend/internal/filesystem/validator.go` - Path whitelist enforcement

**How it works:**
1. `AllowedPaths` configured in `config.yaml` (default: `["~"]`, i.e., home directory)
2. `PathValidator` checks that requested paths are within allowed prefixes
3. `Browser.ListDir()` returns file/directory metadata (name, path, size, mod time, mode)
4. `ValidatePath()` performs comprehensive checks: existence, readability, writability

**Security:**
- Path traversal prevention via prefix matching
- Writability tested by creating a temp file (`.orchestra_write_test`)
- No file upload/download through browser endpoint (only metadata)

**Endpoints:**
- `GET /api/workspaces/:id/browse` - Browse at workspace path
- `GET /api/browse` - Browse server root
- `POST /api/workspaces/validate-path` - Validate a path

## Authentication Mechanisms

**Dual-mode authentication:** JWT + legacy token fallback

**JWT Authentication:**
- Algorithm: HS256
- Claims: `userId` (string), `username` (string), standard `exp`, `iat`, `iss`
- Issuer: `orchestra`
- Default expiration: 24 hours
- Secret from config: `auth.jwt_secret` or env var `ORCHESTRA_JWT_SECRET`
- Token stored in frontend: `localStorage.getItem('orchestra.auth.token')`

**Legacy Token Authentication:**
- Single shared token from env var `ORCHESTRA_AUTH_TOKEN`
- Accepted via: `Authorization` header, `X-Auth-Token` header, or `?token=` query param
- WebSocket connections use query parameter (browser WebSocket API limitations)

**Auth bypass:**
- `ORCHESTRA_AUTH_DISABLED=true` disables auth entirely
- When auth disabled, login returns mock user `anonymous`
- All API routes still pass through middleware but skip validation

**Password hashing:** bcrypt via `golang.org/x/crypto/bcrypt`

**Files:**
- `backend/internal/security/jwt.go` - JWT config, generate, validate
- `backend/internal/security/password.go` - bcrypt hash/verify
- `backend/internal/api/middleware/auth.go` - Gin middleware
- `frontend/src/features/auth/authStore.ts` - Frontend auth state

## Data Storage

**Database:** SQLite (local file, no external database server)
- Path: `./data/orchestra.db` (configurable)
- Client: `database/sql` + `mattn/go-sqlite3` (CGo driver)
- No ORM - raw SQL with repository pattern
- Repository files: `backend/internal/storage/repository/`

**File Storage:**
- Upload directory: `./uploads` (configurable via `server.upload_dir`)
- Workspace directories: `./workspaces` (configurable via `storage.workspaces`)
- No cloud file storage (S3, etc.) detected

**Caching:** None - no Redis or in-memory cache beyond Go process memory

## Environment Configuration

**Required environment variables:**
- `ORCHESTRA_ENCRYPTION_KEY` - For API key encryption (AES-GCM, 32+ bytes)
- `ORCHESTRA_JWT_SECRET` - JWT signing secret (required when `auth.enabled: true`)
- `ORCHESTRA_AUTH_TOKEN` - Legacy auth token (alternative to JWT)
- `ORCHESTRA_AUTH_DISABLED` - Set to `true` to disable auth
- `ORCHESTRA_CONFIG` - Custom config file path

**Config file:** `backend/configs/config.yaml`

**Config loading:** `backend/internal/config/loader.go` - Loads YAML then overrides with `ORCHESTRA_*` env vars

## CI/CD & Deployment

**Hosting:** Self-hosted (no cloud platform detected)
- Backend: Compiled Go binary
- Frontend: Static files served by Gin or separate web server

**CI Pipeline:** None configured (no `.github/workflows/`, `.gitlab-ci.yml`, etc.)

**E2E Testing:**
- Playwright tests in `frontend/tests/e2e/`
- Requires backend running on `ORCHESTRA_API_URL` (default `http://127.0.0.1:8080`)
- Configurable via `ORCHESTRA_E2E_BASE_URL`

## Webhooks & Callbacks

**Incoming:** None (no webhook endpoints detected)

**Outgoing:**
- A2A agent communication: HTTP POST + SSE subscription to agent URLs
  - Agents must expose A2A-compliant endpoints (JSONRPC over HTTP)
  - SSE used for streaming task updates from agents to Orchestra

## Monitoring & Observability

**Error Tracking:** None - errors logged via Go `log` package only

**Logs:** Standard output (`os.Stdout`)
- Request logging via custom `middleware.Logger()`
- Structured package prefixes: `[ChatHub]`, `[a2a-ws]`, `[a2a-pool]`, etc.
- No log rotation or aggregation

**Swagger Docs:** Available at `/swagger/*any` during development

---

*Integration audit: 2026-04-08*
