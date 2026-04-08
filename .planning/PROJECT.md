# Orchestra — Multi-Agent Collaboration Platform

## What This Is

A web-based multi-agent collaboration system that supports multiple AI agents (Claude Code, Gemini CLI, etc.) running in parallel on a server, with orchestration capabilities. The UI is a web client (Vue 3) — CLIs and PTYs run on the machine hosting Orchestra. Aligned with a reference Tauri desktop app used internally for parity.

## Core Value

Coordinate multiple AI agents to collaborate on shared work through workspace-scoped conversations, tasks, and terminal sessions.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] **AUTH-01**: User can sign in with username/password (JWT + legacy token)
- [ ] **AUTH-02**: Auth can be disabled for local-only use
- [ ] **WS-01**: Real-time chat message broadcasting via WebSocket
- [ ] **WS-02**: Real-time terminal session streaming via WebSocket
- [ ] **WRK-01**: Browse, create, switch, and delete workspaces
- [ ] **WRK-02**: Path browser for server-side workspace directories
- [ ] **MBR-01**: Full member CRUD with roles (owner/admin/secretary/assistant/member)
- [ ] **MBR-02**: Member presence broadcasting (typing, viewing, idle)
- [ ] **A2A-01**: A2A (Agent-to-Agent) protocol for agent communication
- [ ] **A2A-02**: AgentBridge auto-bridges A2A output to chat messages
- [ ] **A2A-03**: ToolHandler executes 9 tools (chat_send, task CRUD, workload, status, file ops)
- [ ] **A2A-04**: Agent status broadcasting via WebSocket
- [ ] **CONV-01**: Conversations (channels + DMs) with pinned/muted, member management
- [ ] **CONV-02**: Message CRUD with read tracking and AI-flagged messages
- [ ] **TASK-01**: Secretary-to-assistant task management (create/start/complete/fail)
- [ ] **TASK-02**: Task list UI with status filtering
- [ ] **ATTACH-01**: File upload/download with workspace-scoped directories
- [ ] **KEY-01**: API key management with encrypted storage and test endpoints
- [ ] **UI-01**: Modern Soft-Light Glass themed UI
- [ ] **UI-02**: Responsive design for mobile and desktop

### Out of Scope

- Real mobile app — web-first, responsive design sufficient
- Third-party OAuth — JWT + legacy token sufficient for current use case
- Video/voice calls — out of scope for text-based AI agent collaboration
- Public sharing/external collaboration — internal tool only

## Context

- **Product charter**: Align with reference Tauri desktop app in behavior/contracts, port features via HTTP/WebSocket and server paths
- **Recent major migration**: Migrated from PTY-based terminal to A2A (Agent-to-Agent) protocol — old terminal layer removed
- **Architecture**: Go backend serving Vue SPA, WebSocket for real-time, SQLite for storage, repository pattern for data access
- **Known gaps**: Frontend language selector and theme toggle are static placeholders; first-run UX lacks auto-workspace creation; three message paths exist (HTTP, internal API, AgentBridge)

## Constraints

- **Tech stack**: Go + Gin + gorilla/websocket + SQLite / Vue 3 + TypeScript + Pinia + Tailwind — already established, no changes
- **Compatibility**: Must remain recognizable to users of the reference Tauri desktop app
- **Performance**: Single SQLite instance — no horizontal scaling expected

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| A2A over PTY | PTY doesn't support structured agent communication | ✓ Good — enables tool calls and structured output |
| Vue SPA over SSR | Real-time WebSocket UI needs client-side state | ✓ Good — fits real-time collaboration model |
| Repository pattern | Clean separation between data access and business logic | ✓ Good — easy to swap storage backends later |
| WebSocket over polling for chat | 10s polling created unacceptable latency | ✓ Good — just implemented |
| Single SQLite | Simple deployment, no distributed sync needed | ⚠️ Revisit — if multi-server needed later |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-08 after codebase mapping and analysis*
