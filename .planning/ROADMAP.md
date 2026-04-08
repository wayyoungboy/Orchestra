# Roadmap: Orchestra

## Phase Overview

| # | Phase | Goal | Requirements | Success Criteria |
|---|-------|------|--------------|------------------|
| 1 | Foundation | Auth + workspace basics (✅ done) | AUTH-01, AUTH-02, WRK-01, WRK-02 | User can login, create workspace, browse paths |
| 2 | Core Features | Conversations, members, tasks, attachments (✅ done) | MBR-01, CONV-01, CONV-02, TASK-01, TASK-02, ATTACH-01, KEY-01, UI-01 | Full CRUD for members/conversations, task UI works, file attachments functional |
| 3 | A2A + Real-Time | Agent protocol + WebSocket real-time | MBR-02, MBR-03, A2A-01, A2A-02, A2A-03, A2A-04, CONV-03, TERM-01 | A2A output appears in chat, tool calls execute, presence/status broadcast, terminal streams |
| 4 | Polish + E2E | UI polish, first-run UX, end-to-end tests | A2A-04, UI-02, UI-03 | Language switching works, theme toggle works, E2E tests pass, clean first-run experience |

## Phase Details

### Phase 1: Foundation (✅ Complete)
**Goal:** Users can authenticate and manage workspaces with path browsing

Requirements: AUTH-01, AUTH-02, WRK-01, WRK-02

Success criteria:
1. User can sign in with orchestra/orchestra credentials
2. User can create a new workspace by selecting a server-side path
3. Path browser shows directory tree with file type icons
4. User can switch between workspaces without page reload

### Phase 2: Core Features (✅ Complete)
**Goal:** Full collaboration features — members, conversations, tasks, attachments

Requirements: MBR-01, CONV-01, CONV-02, TASK-01, TASK-02, ATTACH-01, KEY-01, UI-01

Success criteria:
1. Users can add/edit/delete members with role assignment
2. Users can create channel and DM conversations, send/receive messages
3. Task page shows tasks by status, secretary can assign to assistants
4. File attachments can be uploaded and downloaded per workspace
5. API keys can be created, tested, and deleted
6. UI uses Modern Soft-Light Glass theme consistently

### Phase 3: A2A + Real-Time (🔧 In Progress)
**Goal:** A2A agents communicate through Orchestra with real-time broadcasting

Requirements: MBR-02, MBR-03, A2A-01, A2A-02, A2A-03, A2A-04, CONV-03, TERM-01

Success criteria:
1. A2A agent output appears as chat messages in the conversation (AgentBridge wired)
2. Agent tool calls execute (ToolHandler wired — chat_send, task CRUD, file ops)
3. Member presence (typing/viewing/idle) broadcasts via WebSocket
4. Agent status updates broadcast via WebSocket
5. Chat messages arrive in real-time (no polling)
6. Terminal sessions stream output via WebSocket

### Phase 4: Polish + E2E (⏳ Pending)
**Goal:** UI polish, first-run UX, end-to-end test coverage

Requirements: A2A-04, UI-02, UI-03

Success criteria:
1. Language selector switches UI between English and Chinese
2. Theme toggle switches between light and dark modes
3. First login creates default workspace automatically
4. E2E test suite covers login → workspace → chat → send message flow
5. No TODOs remaining in handler code

---
*Roadmap created: 2026-04-08*
*Last updated: 2026-04-08 after codebase analysis*
