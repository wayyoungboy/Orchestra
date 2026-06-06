# Orchestra MVP Loop Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the MVP collaboration loop measurable and reliable enough to guide the next autonomous implementation phase.

**Architecture:** Treat `chat -> dispatch -> agent session -> chat/task` as the acceptance path. Add focused backend tests, visible outbox diagnostics, frontend state handling, and a README/roadmap update so future work follows the MVP loop instead of broad parity chasing.

**Tech Stack:** Go 1.21+, Gin, SQLite repositories, Vue 3, Pinia, TypeScript, Playwright/Vitest where already available.

---

## Progress Update - 2026-06-07

Completed hardening passes:

- Added backend and frontend diagnostics around the MVP dispatch loop.
- Reused saved member CLI/ACP configuration when creating member terminal sessions.
- Exposed a visible member-card action for configured `assistant` and `secretary` members to start or reuse their backend agent session.
- Added member-card session probing so existing backend sessions are visible when the Members page loads.
- Added a workspace Agent Sessions navigation tab that lists active backend sessions, owning members, and read-only terminal stream output.

The next product gap is interactive terminal inspection. The member-card session action and Agent Sessions tab make agent startup, ownership, and output visible before a complete terminal input UI is rebuilt.

Validation note: backend tests and frontend production builds are passing. Vitest currently hangs in this local `/Volumes` + pnpm environment before reporting even a minimal smoke test; sampling shows Node spending time in ESM/package resolution filesystem calls. Keep Vitest tests as behavioral guardrails, but use build/e2e/manual verification until the runner environment is fixed.

---

## File Structure

The first hardening pass is intentionally small and avoids changing the session engine unless a test proves a concrete bug.

- Modify `backend/internal/api/handlers/conversation.go`: keep message-send behavior as the core route and add small helper seams only if needed for tests.
- Create `backend/internal/api/handlers/conversation_mvp_loop_test.go`: backend acceptance tests for DM creation, mention dispatch routing inputs, unread sync, and last-message preview.
- Modify `backend/internal/outbox/worker.go`: add a repository-style listing method for workspace-scoped outbox diagnostics.
- Create `backend/internal/outbox/worker_diagnostics_test.go`: tests for status-filtered outbox listing.
- Modify `backend/internal/api/router.go`: register a workspace-scoped outbox diagnostics endpoint.
- Create `backend/internal/api/handlers/outbox.go`: HTTP handler for visible outbox delivery state.
- Modify `frontend/src/features/chat/chatStore.ts`: add state for dispatch/outbox delivery failures returned by the diagnostics endpoint.
- Create `frontend/src/features/chat/dispatchDiagnostics.ts`: focused API helper for outbox diagnostics.
- Modify `frontend/src/features/chat/components/MessagesList.vue`: show a compact delivery warning for failed/dead dispatches associated with the active conversation.
- Modify `README.md` and `README_CN.md`: make the MVP product loop the stated roadmap anchor.

---

## Task 1: Backend MVP Chat Acceptance Tests

**Files:**
- Create: `backend/internal/api/handlers/conversation_mvp_loop_test.go`
- Read: `backend/internal/api/handlers/conversation.go`
- Read: `backend/internal/storage/repository/conversation.go`
- Read: `backend/internal/storage/repository/message.go`

- [ ] **Step 1: Write failing tests for the conversation side of the loop**

Create `backend/internal/api/handlers/conversation_mvp_loop_test.go` with tests for:

```go
package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
)

func newMVPConversationRepos(t *testing.T) (*sql.DB, *repository.ConversationRepository, *repository.MessageRepository, *repository.ConversationReadRepository) {
	t.Helper()
	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate("../storage/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db.DB(), repository.NewConversationRepository(db.DB()), repository.NewMessageRepository(db.DB()), repository.NewConversationReadRepository(db.DB())
}

func TestMVPLoopDMIsStableForSameMembers(t *testing.T) {
	_, convRepo, _, _ := newMVPConversationRepos(t)

	first, created, err := convRepo.GetOrCreateDM("ws-1", "owner-1", "assistant-1")
	if err != nil {
		t.Fatalf("first dm: %v", err)
	}
	if !created {
		t.Fatal("expected first call to create DM")
	}

	second, created, err := convRepo.GetOrCreateDM("ws-1", "assistant-1", "owner-1")
	if err != nil {
		t.Fatalf("second dm: %v", err)
	}
	if created {
		t.Fatal("expected second call to reuse DM")
	}
	if first.ID != second.ID {
		t.Fatalf("expected same DM id, got %s and %s", first.ID, second.ID)
	}
}

func TestMVPLoopLastMessagePreviewUpdatesOnMessageCreate(t *testing.T) {
	_, convRepo, msgRepo, _ := newMVPConversationRepos(t)

	conv, err := convRepo.Create("ws-1", repository.ConversationCreate{
		Type: repository.ConversationTypeChannel,
		Name: "general",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	msg, err := msgRepo.Create(repository.MessageCreate{
		ConversationID: conv.ID,
		SenderID:       "owner-1",
		Content:        repository.MessageContent{Type: "text", Text: "please inspect the workspace loop"},
		IsAI:           false,
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	if err := convRepo.UpdateLastMessage(conv.ID, "please inspect the workspace loop", msg.CreatedAt); err != nil {
		t.Fatalf("update last message: %v", err)
	}

	list, err := convRepo.ListByWorkspace("ws-1")
	if err != nil {
		t.Fatalf("list conversations: %v", err)
	}
	if got := list[0].LastMessagePreview; got != "please inspect the workspace loop" {
		t.Fatalf("last message preview = %q", got)
	}
	if got := list[0].LastMessageAt; got != msg.CreatedAt {
		t.Fatalf("last message at = %d, want %d", got, msg.CreatedAt)
	}
}

func TestMVPLoopUnreadCursorClearsUnreadCount(t *testing.T) {
	_, convRepo, msgRepo, readRepo := newMVPConversationRepos(t)

	conv, err := convRepo.Create("ws-1", repository.ConversationCreate{
		Type: repository.ConversationTypeChannel,
		Name: "general",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	msg, err := msgRepo.Create(repository.MessageCreate{
		ConversationID: conv.ID,
		SenderID:       "assistant-1",
		Content:        repository.MessageContent{Type: "text", Text: "done"},
		IsAI:           true,
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	before, err := readRepo.BatchGetUnreadCounts([]string{conv.ID}, "owner-1")
	if err != nil {
		t.Fatalf("before unread: %v", err)
	}
	if before[conv.ID] != 1 {
		t.Fatalf("unread before mark-read = %d", before[conv.ID])
	}

	if err := readRepo.Upsert(conv.ID, "owner-1", msg.CreatedAt); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	after, err := readRepo.BatchGetUnreadCounts([]string{conv.ID}, "owner-1")
	if err != nil {
		t.Fatalf("after unread: %v", err)
	}
	if after[conv.ID] != 0 {
		t.Fatalf("unread after mark-read = %d", after[conv.ID])
	}

	_ = context.Background()
}
```

- [ ] **Step 2: Run the tests and confirm the current behavior**

Run:

```bash
cd backend
go test ./internal/api/handlers -run 'TestMVPLoop' -count=1
```

Expected: tests fail if migration path or repository behavior does not support the MVP assumptions; otherwise they pass and become guardrails.

- [ ] **Step 3: Fix only proven failures**

If migration lookup fails, replace the migration call in the test helper with:

```go
if err := db.Migrate("internal/storage/migrations"); err != nil {
	t.Fatalf("migrate: %v", err)
}
```

If unread counts include self-authored messages incorrectly, update `ConversationReadRepository.BatchGetUnreadCounts` so it counts messages newer than the cursor and not authored by the viewer.

- [ ] **Step 4: Verify**

Run:

```bash
cd backend
go test ./internal/api/handlers ./internal/storage/repository -count=1
```

Expected: all listed packages pass.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/handlers/conversation_mvp_loop_test.go backend/internal/storage/repository/conversation_read.go
git commit -m "test: add mvp chat loop acceptance coverage"
```

---

## Task 2: Workspace Outbox Diagnostics API

**Files:**
- Modify: `backend/internal/outbox/worker.go`
- Create: `backend/internal/outbox/worker_diagnostics_test.go`
- Create: `backend/internal/api/handlers/outbox.go`
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: Add failing outbox listing tests**

Create `backend/internal/outbox/worker_diagnostics_test.go`:

```go
package outbox

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/storage"
)

func newOutboxDiagnosticsDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate("internal/storage/migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db.DB()
}

func insertOutboxDiagnosticItem(t *testing.T, db *sql.DB, id, workspaceID, conversationID, status string) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, last_error, created_at, updated_at, workspace_id, target_member_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, conversationID, "owner-1", "hello", status, 2, "boom", now, now, workspaceID, "assistant-1")
	if err != nil {
		t.Fatalf("insert outbox item: %v", err)
	}
}

func TestListWorkspaceReturnsFilteredOutboxItems(t *testing.T) {
	db := newOutboxDiagnosticsDB(t)
	worker := New(db, DefaultConfig(), func(context.Context, *Item) error { return nil })
	insertOutboxDiagnosticItem(t, db, "outbox-1", "ws-1", "conv-1", "failed")
	insertOutboxDiagnosticItem(t, db, "outbox-2", "ws-1", "conv-2", "dead")
	insertOutboxDiagnosticItem(t, db, "outbox-3", "ws-2", "conv-1", "failed")

	items, err := worker.ListWorkspace(context.Background(), "ws-1", ListFilter{Status: "failed", Limit: 20})
	if err != nil {
		t.Fatalf("list workspace: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 failed item for ws-1, got %d", len(items))
	}
	if items[0].ID != "outbox-1" || items[0].LastError != "boom" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd backend
go test ./internal/outbox -run TestListWorkspaceReturnsFilteredOutboxItems -count=1
```

Expected: FAIL because `ListFilter` and `ListWorkspace` are not defined.

- [ ] **Step 3: Implement diagnostics listing**

Add to `backend/internal/outbox/worker.go`:

```go
type ListFilter struct {
	Status         string
	ConversationID string
	Limit          int
}

func (w *Worker) ListWorkspace(ctx context.Context, workspaceID string, filter ListFilter) ([]*Item, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	query := `
		SELECT id, conversation_id, sender_id, content, status, attempt_count,
		       COALESCE(last_error, ''), created_at, updated_at,
		       COALESCE(workspace_id, ''), COALESCE(target_member_id, '')
		FROM outbox
		WHERE workspace_id = ?
	`
	args := []interface{}{workspaceID}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.ConversationID != "" {
		query += " AND conversation_id = ?"
		args = append(args, filter.ConversationID)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, filter.Limit)

	rows, err := w.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list workspace outbox: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var it Item
		var createdAt, updatedAt int64
		if err := rows.Scan(&it.ID, &it.ConversationID, &it.SenderID, &it.Content, &it.Status, &it.AttemptCount, &it.LastError, &createdAt, &updatedAt, &it.WorkspaceID, &it.TargetMemberID); err != nil {
			return nil, fmt.Errorf("scan workspace outbox item: %w", err)
		}
		it.CreatedAt = time.Unix(createdAt, 0)
		it.UpdatedAt = time.Unix(updatedAt, 0)
		items = append(items, &it)
	}
	return items, rows.Err()
}
```

- [ ] **Step 4: Add HTTP handler**

Create `backend/internal/api/handlers/outbox.go`:

```go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/outbox"
)

type OutboxHandler struct {
	worker *outbox.Worker
}

func NewOutboxHandler(worker *outbox.Worker) *OutboxHandler {
	return &OutboxHandler{worker: worker}
}

func (h *OutboxHandler) ListWorkspace(c *gin.Context) {
	if h.worker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox worker unavailable"})
		return
	}
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	items, err := h.worker.ListWorkspace(c.Request.Context(), workspaceID, outbox.ListFilter{
		Status:         c.Query("status"),
		ConversationID: c.Query("conversationId"),
		Limit:          limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
```

- [ ] **Step 5: Register route**

In `backend/internal/api/router.go`, add `OutboxHandler *handlers.OutboxHandler` to `Dependencies`, initialize it with:

```go
outboxHandler := handlers.NewOutboxHandler(outboxWorker)
```

and return it in `Dependencies`.

Register in a new helper or near notification routes:

```go
api.GET("/workspaces/:id/outbox", deps.OutboxHandler.ListWorkspace)
```

- [ ] **Step 6: Verify**

Run:

```bash
cd backend
go test ./internal/outbox ./internal/api -count=1
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/outbox/worker.go backend/internal/outbox/worker_diagnostics_test.go backend/internal/api/handlers/outbox.go backend/internal/api/router.go
git commit -m "feat: expose workspace outbox diagnostics"
```

---

## Task 3: Frontend Dispatch Diagnostics State

**Files:**
- Create: `frontend/src/features/chat/dispatchDiagnostics.ts`
- Modify: `frontend/src/features/chat/chatStore.ts`
- Modify: `frontend/src/features/chat/components/MessagesList.vue`

- [ ] **Step 1: Add diagnostics API helper**

Create `frontend/src/features/chat/dispatchDiagnostics.ts`:

```ts
import client from '@/shared/api/client'

export interface OutboxDiagnosticItem {
  id: string
  conversation_id: string
  sender_id: string
  content: string
  status: 'pending' | 'sending' | 'sent' | 'failed' | 'dead'
  attempt_count: number
  last_error: string
  workspace_id: string
  target_member_id: string
}

export async function fetchConversationDispatchDiagnostics(workspaceId: string, conversationId: string) {
  const response = await client.get(`/workspaces/${workspaceId}/outbox`, {
    params: {
      conversationId,
      limit: 20
    },
    skipErrorToast: true
  })
  return (response.data?.items || []) as OutboxDiagnosticItem[]
}
```

- [ ] **Step 2: Store failed/dead dispatch diagnostics**

In `frontend/src/features/chat/chatStore.ts`, import the helper:

```ts
import { fetchConversationDispatchDiagnostics, type OutboxDiagnosticItem } from './dispatchDiagnostics'
```

Add state:

```ts
const dispatchDiagnostics = ref<Record<string, OutboxDiagnosticItem[]>>({})
```

Add action:

```ts
async function loadDispatchDiagnostics(convId: string) {
  if (!workspaceId.value) return
  try {
    const items = await fetchConversationDispatchDiagnostics(workspaceId.value, convId)
    dispatchDiagnostics.value[convId] = items.filter(item => item.status === 'failed' || item.status === 'dead')
  } catch {
    dispatchDiagnostics.value[convId] = []
  }
}
```

Call it after loading messages for an active conversation:

```ts
await loadDispatchDiagnostics(id)
```

Return it from the store:

```ts
dispatchDiagnostics,
loadDispatchDiagnostics,
```

- [ ] **Step 3: Show compact delivery warning**

In `frontend/src/features/chat/components/MessagesList.vue`, read the store diagnostics for the current conversation and show a small warning row above the input-adjacent message list when there are failed/dead items:

```vue
<div v-if="failedDispatches.length" class="dispatch-warning">
  <span>Delivery issue</span>
  <span>{{ failedDispatches.length }} agent dispatch{{ failedDispatches.length === 1 ? '' : 'es' }} need attention.</span>
</div>
```

Use computed state:

```ts
const failedDispatches = computed(() => {
  if (!props.conversationId) return []
  return chatStore.dispatchDiagnostics[props.conversationId] || []
})
```

Add restrained CSS:

```css
.dispatch-warning {
  margin: 8px 12px;
  padding: 8px 10px;
  border: 1px solid #f59e0b;
  background: #fffbeb;
  color: #92400e;
  border-radius: 6px;
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
```

- [ ] **Step 4: Verify frontend build**

Run:

```bash
cd frontend
pnpm build
```

Expected: `vue-tsc` and Vite build pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/features/chat/dispatchDiagnostics.ts frontend/src/features/chat/chatStore.ts frontend/src/features/chat/components/MessagesList.vue
git commit -m "feat: show chat dispatch diagnostics"
```

---

## Task 4: Roadmap Anchor Documentation

**Files:**
- Modify: `README.md`
- Modify: `README_CN.md`
- Modify: `docs/superpowers/plans/2026-05-04-next-phase-plan.md`

- [ ] **Step 1: Update README positioning**

In `README.md`, add this after the opening paragraph:

```markdown
## Product Direction

Orchestra's near-term roadmap is anchored on one MVP loop:

```text
Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state
```

Reference desktop behavior and Golutra specifications are useful guides, but they are not treated as full parity backlogs. Work that improves this loop has priority over broad feature porting.
```

- [ ] **Step 2: Update Chinese README positioning**

In `README_CN.md`, add the Chinese equivalent after the opening paragraph:

```markdown
## 产品方向

Orchestra 近期路线围绕一个 MVP 闭环：

```text
Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state
```

参考桌面端行为与 Golutra 规格用于约束产品语义，但不作为全量 parity backlog。优先级以是否能强化这个闭环为准，而不是追逐所有可迁移功能。
```

- [ ] **Step 3: Reclassify the old next-phase plan**

At the top of `docs/superpowers/plans/2026-05-04-next-phase-plan.md`, add:

```markdown
> **Status update 2026-06-07:** This plan is retained as reference material. New actionable work should be derived from `docs/superpowers/specs/2026-06-07-orchestra-mvp-product-loop-design.md` and should prioritize the MVP loop over broad parity coverage.
```

- [ ] **Step 4: Verify docs contain the MVP anchor**

Run:

```bash
rg -n "Workspace -> Members -> Chat mention/DM -> Dispatch -> Agent session -> Output -> Chat/Task state" README.md README_CN.md docs/superpowers/plans/2026-05-04-next-phase-plan.md
```

Expected: matches in both READMEs and related planning docs.

- [ ] **Step 5: Commit**

```bash
git add README.md README_CN.md docs/superpowers/plans/2026-05-04-next-phase-plan.md
git commit -m "docs: anchor roadmap on mvp product loop"
```

---

## Task 5: End-to-End Smoke Verification

**Files:**
- No source files expected unless smoke verification reveals a concrete bug.

- [ ] **Step 1: Run backend package tests**

Run:

```bash
cd backend
go test ./...
```

Expected: all backend tests pass. If tmux integration tests require local tmux and fail due environment, rerun the failing package with verbose output and document the environmental blocker before changing code.

- [ ] **Step 2: Run frontend build**

Run:

```bash
cd frontend
pnpm build
```

Expected: TypeScript and Vite build pass.

- [ ] **Step 3: Run e2e if both dev processes can be started**

Start backend:

```bash
cd backend
make run
```

In another terminal, start frontend:

```bash
cd frontend
pnpm dev
```

Then run:

```bash
cd frontend
ORCHESTRA_API_URL=http://127.0.0.1:8080 pnpm test:e2e
```

Expected: e2e suite passes or reports only pre-existing unrelated failures. Capture failures with test names and screenshots before making fixes.

- [ ] **Step 4: Commit verification-only fixes if needed**

If a real bug is found and fixed, commit only that fix:

```bash
git add <changed-files>
git commit -m "fix: stabilize mvp loop smoke path"
```

If no code changes are needed, do not create a commit.

---

## Self-Review

Spec coverage:

- Workspace and member setup are not reimplemented in this first plan because existing routes and UI already cover them; this plan focuses on the measurable message loop.
- Chat, DM, unread, last-message preview, outbox visibility, frontend delivery state, and roadmap anchoring are covered.
- Full agent CLI correctness is intentionally left to the next plan after these guardrails exist.

Placeholder scan:

- No `TBD`, `TODO`, or intentionally vague implementation steps remain.

Type consistency:

- Backend outbox diagnostic types use the existing `outbox.Item` shape.
- Frontend diagnostic keys match the current JSON tags on `outbox.Item`.
