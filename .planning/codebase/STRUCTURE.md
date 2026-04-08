# Codebase Structure

**Analysis Date:** 2026-04-08

## Directory Layout

```
Orchestra/
├── backend/                      # Go backend service
│   ├── cmd/server/               # Application entry point
│   │   └── main.go               # Server startup, DI wiring, signal handling
│   ├── configs/                  # Configuration files
│   │   └── config.yaml           # YAML configuration
│   ├── internal/                 # Internal (non-exported) packages
│   │   ├── a2a/                  # A2A (Agent-to-Agent) protocol layer
│   │   ├── api/                  # HTTP API layer
│   │   ├── chatbridge/           # Chat-to-terminal bridging
│   │   ├── config/               # Configuration loading
│   │   ├── filesystem/           # Server-side file browsing
│   │   ├── models/               # Domain data structures
│   │   ├── security/             # Auth, JWT, crypto, whitelisting
│   │   ├── storage/              # Database layer
│   │   └── ws/                   # WebSocket gateway
│   ├── pkg/utils/                # Shared utilities (exported)
│   ├── docs/                     # Swagger auto-generated docs
│   └── data/                     # SQLite database (gitignored)
├── frontend/                     # Vue 3 SPA
│   ├── src/
│   │   ├── app/                  # Application bootstrap
│   │   ├── assets/               # Global CSS/styles
│   │   ├── features/             # Feature modules (domain-driven)
│   │   ├── i18n/                 # Internationalization
│   │   ├── shared/               # Shared utilities and components
│   │   └── stores/               # Global stores (cross-feature)
│   ├── tests/                    # Playwright E2E tests
│   └── public/                   # Static assets
└── docs/                         # Project documentation
    └── superpowers/
        ├── specs/                # Design specifications
        └── plans/                # Implementation plans
```

## Directory Purposes

### `backend/cmd/server/`
- **Purpose:** Application entry point — single `main.go`
- **Contains:** Server bootstrap, dependency injection, signal handling, default user creation
- **Key file:** `backend/cmd/server/main.go`

### `backend/internal/a2a/`
- **Purpose:** A2A (Agent-to-Agent) protocol implementation
- **Contains:**
  - `pool.go` — Session pool manager with idle timeout cleanup
  - `session.go` — Individual agent session with SSE subscription, message conversion
  - `registry.go` — Agent URL registry
  - `tool_handler.go` — Tool execution routing
  - `messages.go` — ACP message types and conversion
- **Key pattern:** Pool manages lifecycle; Session handles HTTP/SSE to a single agent

### `backend/internal/api/`
- **Purpose:** HTTP API routing, middleware, and handlers
- **Contains:**
  - `router.go` — Gin engine setup, all route definitions, DI of handlers/repos
  - `handlers/` — One file per domain (auth, workspace, member, conversation, terminal, task, attachment, api_key, health)
  - `middleware/` — Auth (JWT + legacy), CORS, logging
- **Key file:** `backend/internal/api/router.go` — single source of truth for all routes

### `backend/internal/chatbridge/`
- **Purpose:** Bridge between chat conversations and terminal sessions
- **Contains:** `agent_bridge.go` — connects chat messages to A2A agent sessions
- **Note:** Previous files (`bridge.go`, `filter_noise.go`, `screen.go`, `strip_tty.go`) have been removed; functionality consolidated

### `backend/internal/config/`
- **Purpose:** Configuration loading and defaults
- **Contains:**
  - `config.go` — Config struct definitions (Server, Terminal, Security, Storage, Auth)
  - `loader.go` — YAML file loading with env var override support
  - `loader_test.go` — Unit tests

### `backend/internal/filesystem/`
- **Purpose:** Server-side file system browsing with path validation
- **Contains:**
  - `browser.go` — Directory listing with `FileInfo` output
  - `validator.go` — Path security (prevents traversal outside allowed paths)
- **Key pattern:** Validator -> Browser pipeline; paths validated against whitelist before any `os.ReadDir`

### `backend/internal/models/`
- **Purpose:** Domain data structures
- **Contains:**
  - `workspace.go` — Workspace, WorkspaceCreate, WorkspaceUpdate
  - `member.go` — Member, MemberCreate, Presence, PresenceUpdate, MemberRole enum
  - `user.go` — User, UserLogin, UserCreate, AuthToken
  - `task.go` — Task, TaskStatus enum, TaskCreate, TaskStatusUpdate, AgentWorkload
  - `api_key.go` — APIKey, APIKeyProvider enum
  - `attachment.go` — Attachment model
  - `agent_status.go` — Agent status tracking
- **Pattern:** Model struct + Request/Response structs in same file

### `backend/internal/security/`
- **Purpose:** Security primitives
- **Contains:**
  - `jwt.go` — JWTConfig: GenerateToken, ValidateToken (HS256)
  - `password.go` — Password hashing
  - `crypto.go` — KeyEncryptor for API key encryption
  - `whitelist.go` — AllowedCommands and AllowedPaths enforcement
  - `*_test.go` — Unit tests

### `backend/internal/storage/`
- **Purpose:** Database connection and migrations
- **Contains:**
  - `database.go` — `Database` wrapper: NewDatabase (WAL + foreign keys), Migrate, Close
  - `migrations/` — Numbered SQL migration files (001_init.sql through 008_a2a_support.sql)
  - `repository/` — Data access layer

### `backend/internal/storage/repository/`
- **Purpose:** Repository pattern implementations
- **Contains:**
  - `interface.go` — Repository interface definitions (WorkspaceRepository, MemberRepository, TaskRepository, APIKeyRepository)
  - `workspace.go` — Workspace CRUD
  - `member.go` — Member CRUD
  - `task.go` — Task CRUD with status filtering
  - `api_key.go` — API key management
  - `conversation.go` — Conversation CRUD
  - `message.go` — Message CRUD
  - `conversation_read.go` — Read tracking
  - `user.go` — User CRUD
  - `attachment.go` — Attachment storage

### `backend/internal/ws/`
- **Purpose:** WebSocket gateway and handlers
- **Contains:**
  - `gateway.go` — Gateway struct: HandleTerminal, HandleChat, origin checking
  - `a2a_terminal.go` — A2ATerminalHandler: readLoop, writeLoop, ACP-to-WS conversion
  - `chat.go` — ChatHandler, ChatHub (pub/sub broadcast), ChatClient, ChatEvent types
  - `gateway_test.go` — Unit tests

### `backend/pkg/utils/`
- **Purpose:** Exported utility functions
- **Contains:** ID generation (GenerateID, GenerateULID-style), helpers

### `frontend/src/app/`
- **Purpose:** Application bootstrap and routing
- **Contains:**
  - `main.ts` — Vue app creation, Pinia, i18n, Router setup
  - `App.vue` — Root component with router-view, ContextMenuHost, ToastStack
  - `router.ts` — Vue Router with auth guard

### `frontend/src/features/auth/`
- **Purpose:** Authentication flow
- **Contains:**
  - `authStore.ts` — Pinia store: login, logout, register, config fetching, token management
  - `LoginPage.vue` — Login form component

### `frontend/src/features/chat/`
- **Purpose:** Chat/conversation management
- **Contains:**
  - `chatStore.ts` — Pinia store: conversation loading, message polling, sending
  - `ChatInterface.vue` — Main chat layout
  - `types.ts` — TypeScript types for chat
  - `components/` — ChatHeader, ChatInput, ChatSidebar, MessagesList, MembersSidebar, CreateConversationModal, InviteMenu

### `frontend/src/features/members/`
- **Purpose:** Member management
- **Contains:**
  - `memberStore.ts` — Pinia store for member CRUD
  - `MembersPage.vue` — Member list page
  - `AddMemberModal.vue`, `EditMemberModal.vue` — Member forms
  - `MemberRow.vue` — Single member display
  - `index.ts` — Barrel export

### `frontend/src/features/terminal/`
- **Purpose:** Terminal session management
- **Contains:**
  - `terminalStore.ts` — Pinia store for terminal sessions
  - `TerminalPane.vue` — xterm.js terminal component
  - `TerminalWorkspace.vue` — Multi-terminal workspace layout
  - `terminalChatBridge.ts` — Bridge between chat and terminal
  - `terminalMemberStore.ts` — Member-specific terminal state

### `frontend/src/features/workspace/`
- **Purpose:** Workspace selection and management
- **Contains:**
  - `workspaceStore.ts` — Pinia store for workspaces
  - `projectStore.ts` — Project/workspace binding
  - `WorkspaceSelection.vue` — Workspace picker
  - `WorkspaceMain.vue` — Main workspace layout with nav
  - `PathBrowser.vue` — Server-side path browser component

### `frontend/src/features/settings/`
- **Purpose:** User settings
- **Contains:**
  - `Settings.vue` — Settings page layout
  - `settingsStore.ts` — Settings state (locale, theme, etc.)
  - `ApiKeysSection.vue` — API key management UI
  - `apiKeyStore.ts` — API key store

### `frontend/src/features/tasks/`
- **Purpose:** Task management UI
- **Contains:**
  - `TasksPage.vue` — Task list page
  - `TaskCard.vue` — Individual task display
  - `taskStore.ts` — Task state management

### `frontend/src/features/notifications/`
- **Purpose:** Notification system
- **Contains:**
  - `notificationStore.ts` — Notification state

### `frontend/src/shared/`
- **Purpose:** Shared cross-feature utilities

**`shared/api/`** — API layer
- `client.ts` — Axios instance with auth interceptor and error handling
- `errors.ts` — Error type definitions
- `index.ts` — Barrel export
- `member.ts` — Member-specific API functions
- `terminal.ts` — Terminal-specific API functions
- `workspace.ts` — Workspace-specific API functions

**`shared/socket/`** — WebSocket clients
- `index.ts` — Barrel export
- `terminal.ts` — `TerminalSocket` class with reconnect logic
- `chat.ts` — Chat WebSocket client
- `types.ts` — Terminal message type definitions

**`shared/types/`** — Shared TypeScript types
- `index.ts` — Barrel export
- `chat.ts`, `terminal.ts`, `workspace.ts`, `member.ts`, `settings.ts`, `acp.ts`

**`shared/components/`** — Shared UI components
- `NotificationToast.vue`, `ToastStack.vue` — Toast notifications
- `SidebarNav.vue` — Navigation sidebar
- `WorkspaceSwitcher.vue` — Workspace selector

**`shared/composables/`** — Vue composables
- `useAppShortcuts.ts` — Keyboard shortcuts
- `useKeyboard.ts` — Keyboard utilities

**`shared/context-menu/`** — Context menu system
- `ContextMenuHost.vue` — Global context menu renderer
- `useContextMenu.ts` — Composable for triggering context menus

**`shared/utils/`** — Utility functions
- `markdown.ts` — Markdown processing
- `stripAnsiForChat.ts` — ANSI code stripping

**Other shared files:**
- `shared/messageQueue.ts` — Message queue utility
- `shared/notifyError.ts` — Centralized error notification
- `shared/tabSync.ts` — Cross-tab synchronization

### `frontend/src/stores/`
- **Purpose:** Global (cross-feature) Pinia stores
- **Contains:**
  - `diagnosticsStore.ts` — Diagnostic/dev tools state
  - `toastStore.ts` — Toast notification state

### `frontend/src/i18n/`
- **Purpose:** Internationalization
- **Contains:**
  - `index.ts` — Vue I18n setup
  - `locales/en.json`, `locales/zh.json` — Translation files

## Key File Locations

**Entry Points:**
- `backend/cmd/server/main.go`: Backend server bootstrap
- `frontend/src/app/main.ts`: Frontend Vue app bootstrap
- `frontend/src/app/router.ts`: Vue Router with auth guard

**Configuration:**
- `backend/configs/config.yaml`: Server configuration
- `frontend/package.json`: Frontend dependencies and scripts

**Core Logic:**
- `backend/internal/api/router.go`: All route definitions
- `backend/internal/storage/repository/interface.go`: Repository interfaces
- `backend/internal/ws/gateway.go`: WebSocket connection routing

**Testing:**
- `backend/internal/ws/gateway_test.go`: WebSocket gateway tests
- `backend/internal/security/crypto_test.go`: Crypto tests
- `backend/internal/config/loader_test.go`: Config loader tests
- `backend/internal/api/middleware/cors_test.go`: CORS middleware tests
- `frontend/tests/`: Playwright E2E tests

## Naming Conventions

**Backend (Go):**
- Files: `snake_case.go` (e.g., `chat_handler.go`, `api_key.go`)
- Packages: lowercase single word (e.g., `handlers`, `middleware`, `a2a`)
- Structs: PascalCase (e.g., `AuthHandler`, `ChatHub`)
- Interfaces: PascalCase with -er suffix when applicable (e.g., `WorkspaceRepository`)
- Constructors: `New*()` pattern (e.g., `NewAuthHandler`, `NewPool`)

**Frontend (TypeScript/Vue):**
- Components: PascalCase.vue (e.g., `ChatInterface.vue`, `TerminalPane.vue`)
- Stores: camelCaseStore.ts (e.g., `authStore.ts`, `chatStore.ts`)
- Types: camelCase.ts (e.g., `chat.ts`, `terminal.ts`)
- API modules: camelCase.ts (e.g., `client.ts`, `member.ts`)
- Composables: usePascalCase.ts (e.g., `useKeyboard.ts`)

**Directories:**
- Backend: lowercase (e.g., `internal/api/handlers/`)
- Frontend: camelCase for features (e.g., `features/chat/`), lowercase for shared (e.g., `shared/api/`)

## Module Boundaries

### Backend Internal Modules
```
cmd/server/main.go
  └── config          ← Configuration loading
  └── security        ← Auth, JWT, crypto
  └── storage         ← Database + migrations
  └── models          ← Domain structs
  └── repository      ← Data access (depends on models + storage)
  └── a2a             ← Agent sessions (depends on models + pkg/utils)
  └── filesystem      ← Path browsing (independent)
  └── ws              ← WebSocket (depends on a2a, models)
  └── api/handlers    ← HTTP handlers (depends on repos, a2a, ws, fs, security)
  └── api/middleware  ← Middleware (depends on security)
  └── api/router.go   ← Wiring everything together
```

**Dependency rules:**
- `models` has no internal dependencies
- `storage` depends only on `models`
- `repository` depends on `models` + `database/sql`
- `security` is independent
- `a2a` depends on `models` + `pkg/utils`
- `ws` depends on `a2a`
- `filesystem` is independent
- `api/handlers` is the top-level consumer: depends on repos, a2a, ws, fs, security
- `api/router.go` wires everything

### Frontend Feature Modules
```
app/main.ts
  └── Pinia (stores)
  └── Vue Router
  └── i18n

features/auth/     ← Authentication (authStore, LoginPage)
features/chat/     ← Chat interface (chatStore, ChatInterface, components)
features/members/  ← Member management (memberStore, MembersPage, modals)
features/terminal/ ← Terminal (terminalStore, TerminalPane, bridge)
features/workspace/← Workspace (workspaceStore, PathBrowser)
features/settings/ ← Settings (settingsStore, ApiKeysSection)
features/tasks/    ← Tasks (taskStore, TasksPage)
features/notifications/ ← Notifications

shared/            ← Cross-feature utilities (api, socket, types, components)
stores/            ← Global stores (toastStore, diagnosticsStore)
```

**Dependency rules:**
- `features/*` can import from `shared/` and `stores/`
- `features/*` can import from other `features/` (e.g., chat imports auth)
- `shared/` must not import from `features/`
- `stores/` is for cross-cutting state only

## Where to Add New Code

**New API endpoint:**
1. Add route in `backend/internal/api/router.go`
2. Add handler method in appropriate file under `backend/internal/api/handlers/`
3. Add repository methods in `backend/internal/storage/repository/` if needed
4. Add model struct in `backend/internal/models/` if needed

**New WebSocket endpoint:**
1. Add route in `backend/internal/api/router.go` with `wsAuth` middleware
2. Add handler method in `backend/internal/ws/gateway.go`
3. Create handler file in `backend/internal/ws/` for complex logic

**New feature (frontend):**
1. Create `frontend/src/features/{name}/` directory
2. Add `{name}Store.ts` for Pinia state
3. Add `MainPage.vue` or main component
4. Add route in `frontend/src/app/router.ts`
5. Use `shared/api/client.ts` for HTTP calls
6. Use `shared/socket/` for WebSocket if real-time

**New shared type:**
- Add to `frontend/src/shared/types/{domain}.ts`
- Export from `frontend/src/shared/types/index.ts`

**New shared API function:**
- Add to `frontend/src/shared/api/{domain}.ts`

**New migration:**
- Add numbered SQL file to `backend/internal/storage/migrations/` (e.g., `009_new_feature.sql`)

## Special Directories

**`backend/data/`**
- Purpose: SQLite database file (`orchestra.db`)
- Generated: Yes (created on first run)
- Committed: No (gitignored)

**`backend/docs/`**
- Purpose: Swagger auto-generated documentation
- Generated: Yes (by swag CLI)
- Committed: Yes

**`backend/backend/`**
- Purpose: Compiled binary output (from `make build`)
- Generated: Yes
- Committed: No (gitignored)

**`backend/bin/`**
- Purpose: Additional binary artifacts
- Generated: Yes
- Committed: No (gitignored)

**`backend/uploads/`**
- Purpose: File attachment storage
- Generated: Yes (on upload)
- Committed: No (gitignored)

---

*Structure analysis: 2026-04-08*
