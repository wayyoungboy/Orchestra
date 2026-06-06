package outbox

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
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
	t.Cleanup(func() { db.Close() })

	migrationDirs := []string{
		"internal/storage/migrations",
		"../storage/migrations",
	}
	var migrateErr error
	for _, dir := range migrationDirs {
		cleanDir := filepath.Clean(dir)
		if _, err := os.Stat(cleanDir); err != nil {
			continue
		}
		migrateErr = db.Migrate(cleanDir)
		if migrateErr == nil {
			break
		}
	}
	if migrateErr != nil {
		t.Fatalf("migrate: %v", migrateErr)
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

	items, err := worker.ListWorkspace(context.Background(), "ws-1", ListFilter{
		Status: "failed",
		Limit:  20,
	})
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
