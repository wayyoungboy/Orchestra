# Codebase Concerns

**Analysis Date:** 2026-04-08

## CRITICAL -- Security Vulnerabilities

### Path Traversal in Filesystem Validator

**Severity:** Critical
**Files:** `backend/internal/filesystem/validator.go`

The `ValidatePath` function uses `strings.HasPrefix` for directory matching without a trailing separator check:

```go
// validator.go:33-40
if strings.HasPrefix(absPath, absAllowed) {
    return nil
}
```

If `/volumes/code` is an allowed path, then `/volumes/code2/secret` passes validation because the string `"/volumes/code2/secret"` starts with `"/volumes/code"`. This is a classic path traversal vulnerability.

**Fix:** Append `string(filepath.Separator)` to the allowed path before prefix matching, or use `strings.HasPrefix(absPath, absAllowed+string(filepath.Separator))` with an exact equality fallback.

### Workspace Browse Endpoint Bypasses Path Validation

**Severity:** Critical
**Files:** `backend/internal/api/handlers/workspace.go:236-259`

The `Browse` handler accepts a `subPath` query parameter and uses it directly without validating it against the workspace's configured path:

```go
// workspace.go:246-248
fullPath := ws.Path
if subPath != "" {
    fullPath = subPath  // User-controlled path used directly!
}
```

An authenticated user can browse any path on the server by passing `?path=/etc` regardless of workspace configuration. Combined with the path traversal bug above, this exposes arbitrary server filesystem contents.

**Fix:** Resolve the full path and verify it is a subdirectory of the workspace's configured base path.

### JWT Default Secret Hardcoded

**Severity:** Critical
**Files:** `backend/internal/security/jwt.go:29-37`

```go
// jwt.go:30-31
if secret == "" {
    secret = "default-secret-change-in-production"
}
```

If `ORCHESTRA_JWT_SECRET` is not explicitly set, the JWT signing key falls back to a well-known string. Any attacker who knows this default can forge valid JWT tokens for any user, gaining full admin access.

**Fix:** Panic or refuse to start if `JWTSecret` is empty in production mode. Never default to a known secret.

### Default Credentials Created Automatically

**Severity:** Critical
**Files:** `backend/cmd/server/main.go:130-148`

When auth is enabled and no users exist, a default user `orchestra/orchestra` is created with a well-known password. This user is auto-promoted to workspace owner. The credentials are logged to stdout.

**Fix:** Require an initial setup flow or environment-variable-provided admin credentials. Log the warning that default credentials were created and must be changed.

### Auth Disabled by Default

**Severity:** High
**Files:** `backend/internal/config/config.go:63`

```go
Auth: AuthConfig{
    Enabled: false,
    JWTSecret: "",
    ...
}
```

Authentication is disabled out of the box. If a user deploys Orchestra without reading the docs, the entire API is publicly accessible with no authentication required.

**Fix:** Default to `Enabled: true` and require explicit opt-out for development mode.

### Legacy Token Authentication Weakness

**Severity:** High
**Files:** `backend/internal/api/middleware/auth.go`

The auth middleware supports legacy token authentication alongside JWT:
- Query parameter `?token=...` (line 94) -- tokens in URLs are logged in server access logs, browser history, and proxy logs
- Simple string comparison with no expiry or rotation
- Stored as `ORCHESTRA_AUTH_TOKEN` env var, shared across all users

**Fix:** Deprecate legacy token auth. Remove query parameter token support entirely.

### No Rate Limiting on Login

**Severity:** High
**Files:** `backend/internal/api/handlers/auth.go:54-105`, `backend/internal/api/router.go:80-91`

The `/api/auth/login` endpoint has no rate limiting, brute force protection, or account lockout. An attacker can attempt unlimited password guesses.

**Fix:** Add rate limiting middleware (e.g., 5 attempts per minute per IP).

### API Key Handler Returns Decrypted Key Previews to All Authenticated Users

**Severity:** High
**Files:** `backend/internal/api/handlers/api_key.go:57-90`

The `List` endpoint returns `KeyPreview` values (first 8 + last 4 characters of decrypted keys) to any authenticated user. For a 40-character API key, this reveals 12 of 40 characters, significantly reducing brute force search space.

**Fix:** Restrict key visibility to workspace owners/admins only. Mask more aggressively.

### File Upload Content-Type Validation by Extension Only

**Severity:** Medium-High
**Files:** `backend/internal/api/handlers/attachment.go:366-403`

The `detectMimeType` function uses file extensions to determine MIME type rather than inspecting file headers (magic bytes). A malicious actor can upload an executable with a `.png` extension and it will be served with `image/png` content type.

**Fix:** Use `http.DetectContentType` or the `net/http` package's built-in detection to inspect actual file content. Add file extension allowlisting.

---

## HIGH -- Reliability & Data Integrity

### Encryptor Initialized But Never Used for API Keys

**Severity:** High
**Files:** `backend/cmd/server/main.go:110`

```go
_ = encryptor // TODO: use encryptor for API key encryption
```

The `KeyEncryptor` is created from `ORCHESTRA_ENCRYPTION_KEY` but never passed to any handler. The `APIKeyHandler` creates its own encryptor internally (using a hardcoded dev key if none is configured). This means:
1. The intended encryption key from config is never used
2. The TODO has been in the codebase for an unknown duration
3. Production API key encryption may be using the dev key `"dev-mode-encryption-key-32-bytes!!"`

**Fix:** Pass the config-level encryptor to `NewAPIKeyHandler` instead of letting it create its own.

### Deleted Core Components May Break Existing Deployments

**Severity:** High
**Git Status:** Multiple `D` entries

The following files were deleted in the current working tree:
- `backend/internal/terminal/pool.go`, `pool_test.go`, `pty.go`, `session.go` -- Old PTY terminal pool
- `backend/internal/chatbridge/bridge.go`, `screen.go`, `filter_noise.go`, `strip_tty.go` -- Old chat bridge
- `backend/internal/ws/terminal.go` -- Old WebSocket terminal handler
- `frontend/src/features/skills/SkillsPlaceholder.vue` -- UI component

These were replaced by A2A-based terminal handling, but:
1. Any active sessions using the old PTY pool will be broken
2. No migration path or deprecation warning is documented
3. Old test files (`pool_test.go`, `filter_noise_test.go`, `strip_tty_test.go`) were deleted -- reducing test coverage

**Fix:** Complete the A2A migration before deleting old code. Keep old code in a `deprecated/` directory until migration is verified.

### WebSocket Connections Have No Authorization

**Severity:** High
**Files:** `backend/internal/ws/gateway.go:64-102`

The `HandleTerminal` and `HandleChat` methods upgrade WebSocket connections without checking if the user has permission to access the specific session or workspace. The auth middleware (`WebSocketAuth`) only verifies that *some* valid token exists, not that the user has access to the target resource.

Any authenticated user can:
- Connect to any terminal session via `/ws/terminal/:sessionId`
- Connect to any workspace chat via `/ws/chat/:workspaceId`

**Fix:** Add resource-level authorization checks after WebSocket authentication.

### No Graceful Shutdown for Active Sessions

**Severity:** Medium-High
**Files:** `backend/cmd/server/main.go:97-110`

The shutdown sequence only logs a message and exits. Active A2A sessions, WebSocket connections, and SSE subscriptions are not gracefully terminated:
- A2A agent sessions are not notified of shutdown
- WebSocket clients receive abrupt connection closure
- No drain period for in-flight requests

**Fix:** Use `http.Server.Shutdown()` with a context timeout. Add cleanup logic for A2A pool and WebSocket hub.

### Message Deletion Is Not Transactional

**Severity:** Medium
**Files:** `backend/internal/api/handlers/conversation.go:407-423`

The `Delete` handler deletes messages first, then the conversation. If the conversation delete fails, messages are already gone with no rollback:

```go
h.msgRepo.DeleteByConversation(convID)  // No error check!
if err := h.convRepo.Delete(convID); err != nil {
    c.JSON(http.StatusInternalServerError, ...)
    return  // Messages are already deleted!
}
```

**Fix:** Use a database transaction to delete messages and conversation atomically.

---

## MEDIUM -- Scalability & Architecture

### SQLite Limits Multi-User/Concurrent Access

**Severity:** Medium
**Files:** `backend/internal/storage/database.go`, all repository files

SQLite is single-writer. Under concurrent write load (multiple agents sending messages, file uploads, status updates), write contention will cause `database is locked` errors. The current DSN string `_journal_mode=WAL` helps but does not eliminate the fundamental limitation.

No connection pool sizing is configured, and no read replicas or caching layer exists.

**Mitigation for now:** Works fine for single-user or small team deployments. For >10 concurrent writers, migrate to PostgreSQL.

### In-Memory A2A Pool Does Not Scale

**Severity:** Medium
**Files:** `backend/internal/a2a/pool.go`

The A2A session pool is entirely in-memory:
- All sessions are stored in a `map[string]*Session`
- No persistence across server restarts
- Cannot scale horizontally (multiple backend instances would have separate pools)

**Impact:** If the backend restarts, all active agent sessions are lost with no recovery path.

**Fix:** For single-instance deployments, add session persistence to SQLite. For multi-instance, use Redis as a session store.

### Global Singleton ChatHub

**Severity:** Medium
**Files:** `backend/internal/ws/chat.go:52-56`

```go
var GlobalChatHub = &ChatHub{...}
```

The `GlobalChatHub` is a package-level singleton. This makes testing difficult (state leaks between tests) and prevents running multiple independent hub instances. It is passed directly to handlers rather than injected as a dependency.

**Fix:** Inject `ChatHub` via dependency injection. Create a new instance in `main.go` and pass it through.

### Non-Cryptographic Random String Generation

**Severity:** Medium
**Files:** `backend/internal/ws/chat.go:208-219`

```go
func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().Nanosecond()%len(letters)]  // Not cryptographically random!
    }
    return string(b)
}
```

Client IDs are generated using `time.Now().Nanosecond()` modulo the alphabet length. This is predictable and can produce duplicate IDs if two clients connect within the same nanosecond. Combined with the timestamp-based prefix (`"chat-" + time.Now().Format(...)`), collisions are possible under load.

**Fix:** Use `crypto/rand` for client ID generation.

---

## MEDIUM -- Code Quality & Incomplete Features

### TODO Items Not Implemented

**Files and status:**

1. `backend/cmd/server/main.go:110` -- `_ = encryptor // TODO: use encryptor for API key encryption`
   - **Impact:** API key encryption may use dev key in production

2. `backend/internal/api/handlers/member.go:261` -- `// TODO: Broadcast to WebSocket clients`
   - **Impact:** Member presence updates are not broadcast to connected clients

3. `backend/internal/api/handlers/conversation.go:1122` -- `// TODO: Broadcast to WebSocket clients via chat gateway`
   - **Impact:** Agent status updates via `/api/internal/agent-status` are persisted but not pushed to frontend

### Conversation UpdateSettings Accepts Arbitrary Map

**Severity:** Medium
**Files:** `backend/internal/api/handlers/conversation.go:375-394`

```go
var req map[string]interface{}
if err := c.ShouldBindJSON(&req); err != nil { ... }
if err := h.convRepo.Update(convID, req); err != nil { ... }
```

The endpoint accepts any JSON object and passes it directly to the repository's `Update` method. There is no allowlist of permitted fields. An attacker could overwrite fields like `workspace_id`, `type`, or `created_at`.

**Fix:** Define a typed request struct with only the fields that should be updatable (e.g., `Pinned`, `Muted`, `Name`).

### Hardcoded Server URL in Secretary Forward Prompt

**Severity:** Medium
**Files:** `backend/internal/api/handlers/conversation.go:1006-1008`

```go
curl -X POST http://127.0.0.1:8080/api/internal/chat/send \
```

The secretary prompt instructs agents to use `http://127.0.0.1:8080`. If the server runs on a different port or address, agents will fail to send messages.

**Fix:** Pass the actual server URL via config to the prompt template.

### Silent Error Swallowing in Agent Forwarding

**Severity:** Medium
**Files:** `backend/internal/api/handlers/conversation.go:697-727`

The `forwardUserTextToAgent` function silently returns on multiple error conditions:
- `convRepo.GetByID` failure: silently returns
- `memberRepo.ListByWorkspace` failure: silently returns
- Individual `sess.SendUserMessage` errors: ignored with `_ =`

Users will send messages that appear to be delivered but never reach agents, with no indication of failure.

**Fix:** Log errors at minimum. Consider returning error information to the caller.

### No Input Validation on Conversation Create

**Severity:** Medium
**Files:** `backend/internal/api/handlers/conversation.go:199-232`

The `CreateConversationRequest` accepts any string for `Type` without validation. An invalid type could cause downstream issues.

**Fix:** Validate `Type` against known conversation types (`channel`, `dm`).

---

## MEDIUM -- State Management Concerns

### JWT Token Stored in localStorage (XSS Risk)

**Severity:** Medium
**Files:** `frontend/src/features/auth/authStore.ts:88`, `frontend/src/shared/api/client.ts:14`

```typescript
localStorage.setItem('orchestra.auth.token', token)
```

JWT tokens are stored in `localStorage`, which is accessible to any JavaScript running on the page. An XSS vulnerability in any dependency or custom code would allow token theft.

**Mitigation:** Use `httpOnly` cookies for token storage instead of localStorage. Alternatively, implement token rotation with short-lived access tokens and refresh tokens.

### Chat Store Polls Instead of Using WebSocket

**Severity:** Medium
**Files:** `frontend/src/features/chat/chatStore.ts:59-64`

```typescript
function startPolling() {
    stopPolling()
    pollingTimer = setInterval(async () => {
      await refreshAllData()
    }, 10000) // 10 seconds interval
}
```

The chat store uses HTTP polling every 10 seconds as the primary data sync mechanism, even though a WebSocket connection exists for real-time events. This causes:
- Unnecessary API load
- Up to 10-second delay between message send and display on other clients
- Race conditions between WebSocket events and poll results

**Fix:** Use WebSocket as the primary real-time channel. Use polling only for initial data load and reconnection recovery.

### Pinia Store State Not Persisted Across Reloads

**Severity:** Low-Medium
**Files:** `frontend/src/features/chat/chatStore.ts`, `frontend/src/features/terminal/terminalStore.ts`

Chat message history and terminal sessions are stored in volatile Pinia state. A page reload loses all in-memory messages and requires a full re-fetch from the server.

**Fix:** Consider persisting active conversation state in `sessionStorage` or implementing server-side message caching.

---

## LOW -- Code Style and Maintenance

### Untracked New Files Lack Tests

**Files:** Untracked (`??`) files from git status:
- `backend/internal/a2a/` -- Entire new A2A module (no test files detected)
- `backend/internal/api/handlers/api_key.go` -- No test file
- `backend/internal/api/handlers/attachment.go` -- No test file
- `backend/internal/api/handlers/task.go` -- No test file

The A2A module is a core component replacing the old PTY system but has zero test coverage. The old PTY tests (`pool_test.go`) were deleted.

**Fix:** Add unit tests for the A2A pool, session management, and new handlers.

### No API Versioning

All API routes use `/api/` prefix without versioning. Any breaking change to API contracts will break existing frontend deployments.

**Fix:** Use `/api/v1/` prefix for versioned routes.

### Swagger Documentation Exposed in Production

**Files:** `backend/internal/api/router.go:77`

```go
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

The Swagger UI is mounted without any access control. Any user with network access can view the full API documentation, including internal endpoints like `/api/internal/chat/send` and `/api/internal/tasks/*`.

**Fix:** Conditionally mount Swagger only in development mode.

### Database Close Not Deferred with Proper Error Handling

**Files:** `backend/cmd/server/main.go:54`

```go
defer db.Close()
```

The `db.Close()` error is silently discarded. In production, this could mask issues with database connection cleanup.

**Fix:** Use a proper shutdown handler that logs close errors.

---

## Summary by Priority

| Priority | Count | Key Issues |
|----------|-------|------------|
| Critical | 7 | Path traversal, default secrets, auth disabled by default, JWT fallback secret |
| High | 6 | Unused encryptor, deleted core components, WebSocket authorization gap, no rate limiting |
| Medium | 10 | SQLite scaling, in-memory sessions, TODO debt, silent errors, localStorage XSS |
| Low | 4 | Missing tests, no API versioning, Swagger exposure, deferred close |

**Recommended Immediate Actions:**
1. Fix path traversal in `validator.go` (Critical)
2. Remove or secure JWT default secret (Critical)
3. Enable auth by default or require explicit dev-mode flag (Critical)
4. Pass config encryptor to APIKeyHandler (High)
5. Add resource-level authorization to WebSocket handlers (High)
6. Add rate limiting to login endpoint (High)

---

*Concerns audit: 2026-04-08*
