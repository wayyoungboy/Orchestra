package handlers

import (
	"database/sql"
	"os"
	"path/filepath"
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
	t.Cleanup(func() { db.Close() })

	migrationDirs := []string{
		"internal/storage/migrations",
		"../../storage/migrations",
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

	seedMVPWorkspace(t, db.DB())

	return db.DB(),
		repository.NewConversationRepository(db.DB()),
		repository.NewMessageRepository(db.DB()),
		repository.NewConversationReadRepository(db.DB())
}

func seedMVPWorkspace(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO workspaces (id, name, path, last_opened_at, created_at)
		VALUES ('ws-1', 'MVP Workspace', '/tmp/orchestra-mvp', 1, 1)
	`)
	if err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO members (id, workspace_id, name, role_type, created_at)
		VALUES
			('owner-1', 'ws-1', 'Owner', 'owner', 1),
			('assistant-1', 'ws-1', 'Assistant', 'assistant', 1)
	`)
	if err != nil {
		t.Fatalf("seed members: %v", err)
	}
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
		Content: repository.MessageContent{
			Type: "text",
			Text: "please inspect the workspace loop",
		},
		IsAI: false,
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
		Content: repository.MessageContent{
			Type: "text",
			Text: "done",
		},
		IsAI: true,
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
}
