# Orchestra MVP Product Loop Design

**Date**: 2026-06-07  
**Status**: Draft for user review  
**Scope**: Define Orchestra's near-term product direction and constrain implementation priorities around a usable Web agent workspace loop.

---

## 1. Decision

Orchestra should not continue as a broad "reference desktop app parity" backlog. It should become a focused Web agent workspace product whose behavior is informed by the reference desktop model, but whose implementation priorities are driven by one usable collaboration loop.

The core loop is:

```text
Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state
```

Reference desktop behavior remains a constraint for product semantics. It is not a requirement to port every desktop command, diagnostic feature, terminal detail, native window behavior, or plugin surface.

---

## 2. Product Positioning

Orchestra is a server-backed browser product for coordinating AI agents that run on the Orchestra host machine.

The browser is the control and collaboration surface. CLI processes, PTYs, tmux sessions, server-side workspace paths, persistence, and dispatch reliability live on the backend.

The deliberate extension beyond the reference desktop app is server-side workspace and working-directory switching. Users can create workspaces, bind each workspace to a server path, and run agents inside that path.

Everything else should align with the reference app only where it supports the main loop.

---

## 3. MVP User Story

A user opens Orchestra, selects or creates a workspace, binds it to a server-side project path, adds assistant members backed by local CLI providers, and chats with them.

In a channel or DM, the user can mention an agent or ask the secretary to coordinate work. Orchestra reliably records the message, resolves the target member, dispatches the prompt to the right agent session, streams or summarizes the agent output, and updates chat and task state.

The terminal is available for transparency and debugging, but the primary product surface is chat plus tasks.

---

## 4. MVP Scope

The MVP includes:

| Area | Requirement |
| --- | --- |
| Workspace | Create, switch, and bind workspaces to server-side paths. |
| Members | Support `owner`, `admin`, `secretary`, `assistant`, and `member`; assistants can bind CLI/provider settings. |
| Chat | Support channels, DMs, message history, unread counts, and `@mention` targeting. |
| Dispatch | Persist user messages, enqueue dispatch work, resolve targets, retry delivery, and avoid concurrent input corruption. |
| Agent session | Run CLI-backed agent sessions in the workspace path, preserve sessions with tmux where practical, and expose terminal output for inspection. |
| Output return | Convert agent output into chat-visible updates and task state changes when applicable. |
| Tasks | Keep task lifecycle small: create, start, complete, fail, assign, and show status in a Kanban-style view. |
| Notifications | Provide basic unread and agent-completion notification behavior. |

---

## 5. Explicit Non-Goals

The MVP does not include:

| Non-goal | Reason |
| --- | --- |
| Full desktop parity | It turns the project into a large porting backlog instead of a usable Web product. |
| Native desktop behavior | Windows, tray, Tauri IPC, local dialogs, and native plugin loading do not map cleanly to the browser product. |
| Complete Golutra spec coverage | The spec is useful reference material, not the MVP acceptance checklist. |
| Advanced terminal diagnostics | Snapshot triples, exhaustive terminal auditing, and deep emulator state can wait until the user loop is stable. |
| Complex multi-user security | Start with single-user or trusted-team assumptions; reserve stronger isolation for later. |
| SDK provider mode as first-class MVP | CLI/tmux sessions are the current product foundation. SDK providers should come after that loop is stable. |
| Plugin marketplace | Skills/plugin management can remain a CLI or minimal settings capability until core collaboration is reliable. |

---

## 6. Architecture Direction

The current Go + Gin backend and Vue + Pinia frontend remain appropriate.

Backend responsibilities:

- Own workspace path validation and server-side file/path browsing.
- Own member, conversation, task, and settings persistence.
- Own tmux/PTY/CLI process lifecycle.
- Own dispatch queuing, outbox retry, target resolution, and flow control.
- Broadcast chat, terminal, unread, and task events through WebSocket.

Frontend responsibilities:

- Provide a workspace-centered app shell.
- Make chat and task state the primary collaboration surface.
- Expose terminal sessions as inspectable member/session views.
- Keep settings focused on provider credentials, member configuration, theme/language, and diagnostics needed for support.

The implementation should favor shared contracts between chat, dispatch, terminal, and task modules over feature-specific shortcuts.

---

## 7. Data Flow

Primary message flow:

```text
User sends message
  -> API stores message
  -> API emits chat event
  -> Dispatch layer parses mentions / DM target / secretary routing
  -> Outbox records dispatch job
  -> Worker acquires target agent session
  -> Dispatch queue serializes input per agent
  -> Agent session writes to CLI/tmux/PTY
  -> Output processor emits terminal stream and semantic chat/task updates
  -> Frontend updates chat, task board, member status, and unread state
```

This flow is the product acceptance path. Work that does not improve this path should be treated as secondary unless it fixes reliability, correctness, or usability of the path.

---

## 8. Acceptance Criteria

The MVP direction is successful when these scenarios work reliably:

1. A user creates a workspace, binds it to a valid server path, and reopens it later.
2. A user adds an assistant member with a CLI provider and sees its session status.
3. A user sends a channel message with `@assistant`, and only that assistant receives the dispatch.
4. A user sends a DM to an assistant, and the agent response returns to the same DM.
5. Rapid consecutive messages do not corrupt the agent session; they are queued or merged predictably.
6. If dispatch fails transiently, the outbox retries or marks the job failed/dead with visible state.
7. Agent output appears in chat in a human-readable form and remains inspectable in terminal.
8. A secretary can create or update a task, and task state is visible in the task board.
9. Unread counts and completion notifications update without a page refresh.

---

## 9. Roadmap Priorities

Near-term priority order:

1. Stabilize the chat -> dispatch -> agent -> chat/task loop.
2. Tighten member/provider configuration so assistant setup is hard to misconfigure.
3. Improve terminal observability only where it helps debug or trust the loop.
4. Improve task integration with secretary and assistant workflows.
5. Add diagnostics that explain failures in the core loop.
6. Revisit SDK provider mode after CLI-backed collaboration feels stable.

Existing gap-analysis documents should be reclassified as reference material. Actionable plans should be derived from the MVP loop and acceptance criteria above.

---

## 10. Documentation Follow-Up

After this spec is accepted, update the project README or roadmap to distinguish:

- **Product goal**: Web agent workspace with a reliable collaboration loop.
- **Reference constraint**: Desktop app parity guides semantics where relevant.
- **Non-goal**: Full command-for-command or feature-for-feature parity.

This prevents future planning from drifting back into broad parity chasing.
