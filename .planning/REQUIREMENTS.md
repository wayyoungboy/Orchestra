# Requirements: Orchestra — Multi-Agent Collaboration Platform

**Defined:** 2026-04-08
**Core Value:** Coordinate multiple AI agents to collaborate on shared work through workspace-scoped conversations, tasks, and terminal sessions

## v1 Requirements

### Authentication

- [x] **AUTH-01**: User can sign in with username/password (JWT + legacy token)
- [x] **AUTH-02**: Auth can be disabled for local-only use

### Workspace

- [x] **WRK-01**: User can browse, create, switch, and delete workspaces
- [x] **WRK-02**: Path browser for server-side workspace directories

### Members

- [x] **MBR-01**: Full member CRUD with roles (owner/admin/secretary/assistant/member)
- [ ] **MBR-02**: Member presence broadcasting via WebSocket
- [ ] **MBR-03**: Agent status broadcasting via WebSocket

### A2A Protocol

- [x] **A2A-01**: A2A (Agent-to-Agent) protocol for agent communication
- [ ] **A2A-02**: AgentBridge auto-bridges A2A output to chat messages
- [ ] **A2A-03**: ToolHandler executes 9 tools via WebSocket
- [ ] **A2A-04**: End-to-end A2A session stability tested

### Conversations

- [x] **CONV-01**: Conversations (channels + DMs) with pinned/muted, member management
- [x] **CONV-02**: Message CRUD with read tracking and AI-flagged messages
- [x] **CONV-03**: Real-time chat message broadcasting via WebSocket

### Terminal

- [x] **TERM-01**: Real-time terminal session streaming via WebSocket (A2A-based)

### Tasks

- [x] **TASK-01**: Secretary-to-assistant task management (create/start/complete/fail)
- [x] **TASK-02**: Task list UI with status filtering

### Attachments

- [x] **ATTACH-01**: File upload/download with workspace-scoped directories

### API Keys

- [x] **KEY-01**: API key management with encrypted storage and test endpoints

### UI

- [x] **UI-01**: Modern Soft-Light Glass themed UI
- [ ] **UI-02**: Language selector functional (i18n switching)
- [ ] **UI-03**: Theme toggle functional (light/dark)

## v2 Requirements

### Notifications

- **NOTF-01**: Push notifications for new messages
- **NOTF-02**: Desktop notifications for @mentions

### Advanced

- **ADV-01**: Conversation search across all workspaces
- **ADV-02**: Member activity audit log
- **ADV-03**: Workspace-level settings templates

## Out of Scope

| Feature | Reason |
|---------|--------|
| Third-party OAuth | JWT + legacy token sufficient for current internal use |
| Real mobile app | Web-first, responsive design sufficient |
| Video/voice calls | Out of scope for text-based AI agent collaboration |
| Public sharing | Internal tool only |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Complete |
| AUTH-02 | Phase 1 | Complete |
| WRK-01 | Phase 2 | Complete |
| WRK-02 | Phase 2 | Complete |
| MBR-01 | Phase 3 | Complete |
| MBR-02 | Phase 4 | In Progress |
| MBR-03 | Phase 4 | In Progress |
| A2A-01 | Phase 4 | Complete |
| A2A-02 | Phase 4 | In Progress |
| A2A-03 | Phase 4 | In Progress |
| A2A-04 | Phase 5 | Pending |
| CONV-01 | Phase 3 | Complete |
| CONV-02 | Phase 3 | Complete |
| CONV-03 | Phase 4 | Complete |
| TERM-01 | Phase 4 | Complete |
| TASK-01 | Phase 3 | Complete |
| TASK-02 | Phase 3 | Complete |
| ATTACH-01 | Phase 3 | Complete |
| KEY-01 | Phase 3 | Complete |
| UI-01 | Phase 2 | Complete |
| UI-02 | Phase 5 | Pending |
| UI-03 | Phase 5 | Pending |

**Coverage:**
- v1 requirements: 24 total
- Mapped to phases: 24
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-08*
*Last updated: 2026-04-08 after codebase analysis*
