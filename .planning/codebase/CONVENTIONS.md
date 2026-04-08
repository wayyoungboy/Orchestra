# Coding Conventions

**Analysis Date:** 2026-04-08

## Language Overview

This is a dual-language project: Go backend with a Vue 3 + TypeScript frontend.

## Go Coding Style

### File Naming

- All `.go` files use **snake_case** names matching their package purpose: `health.go`, `auth.go`, `workspace.go`, `conversation.go`, `terminal.go`, `member.go`, `cors_test.go`, `gateway_test.go`.
- Test files follow the convention `*_test.go` (e.g., `gateway_test.go`, `database_test.go`).

### Package Organization

- **Entry point:** `backend/cmd/server/main.go`
- **`internal/`** — all private modules, not importable externally:
  - `internal/api/handlers/` — HTTP handler structs (one per domain: `auth.go`, `workspace.go`, `member.go`, `conversation.go`, `terminal.go`)
  - `internal/api/middleware/` — Gin middleware (`cors.go`, `auth.go`)
  - `internal/api/router.go` — route setup via `SetupRouter()`
  - `internal/models/` — domain structs (`member.go`, `workspace.go`)
  - `internal/config/` — YAML config loading and defaults (`config.go`, `loader.go`)
  - `internal/storage/` — database connection and migrations (`database.go`)
  - `internal/storage/repository/` — interface definitions (`interface.go`) + SQL implementations (`workspace.go`, `member.go`, `message.go`, etc.)
  - `internal/ws/` — WebSocket gateway (`gateway.go`)
  - `internal/a2a/` — Agent-to-Agent protocol layer
  - `internal/security/` — crypto, JWT, whitelist validation
  - `internal/filesystem/` — path validation and file browsing
- **`pkg/`** — public utility packages (currently `pkg/utils/`)

### Naming Conventions

**Struct types:** `PascalCase` nouns — `AuthHandler`, `WorkspaceHandler`, `MemberRole`, `HealthResponse`.

**Constructor functions:** `New` prefix + PascalCase — `NewAuthHandler()`, `NewWorkspaceRepository()`, `NewKeyEncryptor()`.

**Methods:** `camelCase` verbs — `ValidateCommand()`, `Encrypt()`, `GetByUsername()`.

**Interfaces:** defined in `internal/storage/repository/interface.go` with descriptive names: `WorkspaceRepository`, `MemberRepository`, `TaskRepository`, `APIKeyRepository`. Interface methods follow `Verb` pattern: `Create()`, `GetByID()`, `List()`, `Update()`, `Delete()`.

**Constants/enums:** Custom type with `const` block using `PascalCase`:
```go
type MemberRole string

const (
    RoleOwner     MemberRole = "owner"
    RoleAdmin     MemberRole = "admin"
    RoleSecretary MemberRole = "secretary"
    RoleAssistant MemberRole = "assistant"
    RoleMember    MemberRole = "member"
)
```

**Variables:** `camelCase` — `workspaces`, `wsRepo`, `authHandler`.

### Go Import Organization

1. Standard library first (`net/http`, `time`, `context`)
2. External packages second (`github.com/gin-gonic/gin`, `github.com/google/uuid`)
3. Internal packages last (`github.com/orchestra/backend/internal/...`)

Each group separated by a blank line. Uses full module path `github.com/orchestra/backend/internal/...` for all internal imports — no path aliases.

### Go Error Handling Patterns

**Handler pattern:** Return HTTP error JSON on every error path, never panic (except init-time):
```go
func (h *WorkspaceHandler) List(c *gin.Context) {
    workspaces, err := h.repo.List(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // nil guard for empty results
    if workspaces == nil {
        workspaces = []*models.Workspace{}
    }
    c.JSON(http.StatusOK, workspaces)
}
```

**Error responses:** Always use `gin.H{"error": "message"}` with appropriate HTTP status codes.

**Sentinel errors:** Defined as package-level vars (e.g., `ErrInvalidKey` in `security/`).

**Repository pattern:** All repository methods accept `context.Context` as first parameter and return `error` as last return value.

### Swagger Annotations

Handlers use swaggo annotations for OpenAPI documentation:
```go
// @Summary Health check
// @Description Check if the server is running
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
```

## TypeScript / Vue Coding Style

### File Naming

- **TypeScript files:** `camelCase.ts` for utilities, `PascalCase.ts` for components and stores — `chatStore.ts`, `authStore.ts`, `client.ts`, `notifyError.ts`
- **Vue SFCs:** `PascalCase.vue` — `ChatInterface.vue`, `ChatInput.vue`, `MessagesList.vue`, `LoginPage.vue`
- **Test files:** `kebab-case.spec.ts` — `chat.spec.ts`, `members.spec.ts`, `errors.spec.ts`

### TypeScript Configuration

**Location:** `frontend/tsconfig.json`

- **Strict mode:** enabled (`"strict": true`)
- **No unused variables:** enabled (`"noUnusedLocals": true`, `"noUnusedParameters": true`)
- **Module:** ESNext with bundler resolution
- **Path alias:** `@/*` maps to `src/*` (configured in both `tsconfig.json` and `vite.config.ts`)
- **JSX:** preserve (for Vue SFC compilation)
- **Target:** ES2020

### Vue SFC Structure

Single-File Components follow this order:
1. `<template>` — template markup
2. `<script setup lang="ts">` — Composition API with TypeScript
3. `<style scoped>` — component-scoped CSS (Tailwind utility classes primarily)

### Pinia Store Convention

**File location:** `src/features/{domain}/{name}Store.ts` (e.g., `src/features/chat/chatStore.ts`, `src/features/auth/authStore.ts`)

**Pattern:** Composition API stores with `defineStore`:
```typescript
export const useChatStore = defineStore('chat', () => {
  // Reactive state
  const conversations = ref<Conversation[]>([])
  const loading = ref(false)

  // Computed
  const activeConversation = computed(() =>
    conversations.value.find(c => c.id === activeConversationId.value) || null
  )

  // Actions (plain async functions)
  async function loadConversations(wsId: string) { ... }

  return { conversations, loading, activeConversation, loadConversations }
})
```

**HMR support:** All stores include hot module replacement:
```typescript
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useChatStore, import.meta.hot))
}
```

### Type Organization

**Shared types** live in `src/shared/types/`:
- `chat.ts` — Conversation and message types
- `member.ts` — Member types  
- `workspace.ts` — Workspace types
- `terminal.ts` — Terminal types
- `settings.ts` — Settings types

**Index barrel file:** `src/shared/types/index.ts` re-exports all types.

### API Client Convention

**Location:** `src/shared/api/client.ts`

- Single Axios instance with base URL `/api`
- Request interceptor attaches Bearer token from `localStorage`
- Response interceptor shows error toasts via `notifyUserError()`
- Special handling: 404 on terminal session probe is silently skipped; 401 on login does not toast
- Custom `skipErrorToast` config option available on request config

### Error Handling (Frontend)

**Pattern:** `try/catch` with `notifyUserError()` utility:
```typescript
try {
  await client.post(`/workspaces/${wsId}/conversations/${convId}/messages`, payload)
} catch (e) {
  notifyUserError('Failed to send message', e)
}
```

**Silent failures:** Polling operations catch and log warnings instead of showing user errors:
```typescript
catch (e) { /* silent during polling */ }
catch (e) { console.warn('Polling tick failed', e) }
```

### API Route Naming

Backend uses RESTful conventions with namespace prefixes:
- `/api/auth/*` — authentication endpoints
- `/api/workspaces/:id/*` — workspace CRUD
- `/api/workspaces/:id/members/*` — member management
- `/api/workspaces/:id/conversations/:convId/*` — conversation operations
- `/api/internal/*` — internal AI assistant API
- `/ws/terminal/:sessionId` — terminal WebSocket
- `/ws/chat/:workspaceId` — chat WebSocket

## Import Organization (Frontend)

1. External packages first (`axios`, `vue`, `pinia`, `@xterm/xterm`)
2. Shared modules via alias (`@/shared/api/client`, `@/shared/notifyError`, `@/shared/types/chat`)
3. Feature modules (`@/features/auth/authStore`, `@/features/workspace/projectStore`)
4. Relative imports for local components

## Configuration Patterns

### Backend Config (`backend/configs/config.yaml`)

YAML configuration loaded at startup with environment variable overrides:
- `ORCHESTRA_ENCRYPTION_KEY` — API key encryption key (32+ bytes)
- `ORCHESTRA_CONFIG` — Custom config file path

Default values defined in `backend/internal/config/config.go` via `Default()` function.

### Environment Variables

- `ORCHESTRA_ENCRYPTION_KEY` — encryption key for API keys (required for production)
- `ORCHESTRA_E2E_BASE_URL` — E2E test base URL override
- `ORCHESTRA_API_URL` — API URL for E2E backend tests
- `CI` — CI environment flag (affects test retry behavior)

### Frontend Proxy (Development)

Vite dev server proxies `/api` and `/ws` to backend at `http://localhost:8080` (configured in `vite.config.ts`).

## Commit Message Convention

**Conventional Commits** format as noted in `CLAUDE.md`:
```
type(scope): description

Examples:
feat(frontend): complete revamp to Modern Soft-Light Glass theme
fix: null check for terminal chat stream payload
feat: Orchestra multi-agent collaboration platform
```

Types observed: `feat`, `fix`, `chore`, `docs`
Scopes: `frontend`, `backend` (optional)

## Code Standards Summary

- **Go:** `gofmt` and `goimports` for formatting
- **TypeScript:** ESLint (`@typescript-eslint`) + `eslint-plugin-vue` for linting
- **CSS:** Tailwind CSS utility classes with custom CSS in `src/assets/main.css`
- **No explicit Prettier config** detected in the repository

---

*Convention analysis: 2026-04-08*
