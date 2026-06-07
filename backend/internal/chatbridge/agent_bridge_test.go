package chatbridge

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
)

type fakeBridgeSession struct {
	workspaceID    string
	memberID       string
	memberName     string
	conversationID string
}

func (s fakeBridgeSession) LastChatTargetConversation() string { return s.conversationID }
func (s fakeBridgeSession) GetWorkspaceID() string             { return s.workspaceID }
func (s fakeBridgeSession) GetMemberID() string                { return s.memberID }
func (s fakeBridgeSession) GetMemberName() string              { return s.memberName }
func (s fakeBridgeSession) TrySendChatStream([]byte)           {}
func (s fakeBridgeSession) NextStreamSeq() uint64              { return 1 }
func (s fakeBridgeSession) StreamSpanID() string               { return "" }

func newBridgeRepos(t *testing.T) (*sql.DB, *repository.ConversationRepository, *repository.MessageRepository) {
	t.Helper()

	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	migrationDirs := []string{
		"../storage/migrations",
		"internal/storage/migrations",
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

	_, err = db.DB().Exec(`
		INSERT INTO workspaces (id, name, path, last_opened_at, created_at)
		VALUES ('ws-bridge', 'Bridge Workspace', '/tmp/orchestra-bridge', 1, 1)
	`)
	if err != nil {
		t.Fatalf("seed workspace: %v", err)
	}

	return db.DB(), repository.NewConversationRepository(db.DB()), repository.NewMessageRepository(db.DB())
}

func TestAgentBridgeAssistantMessagePersistsChatAndConversationPreview(t *testing.T) {
	_, convRepo, msgRepo := newBridgeRepos(t)
	conv, err := convRepo.Create("ws-bridge", repository.ConversationCreate{
		Type: repository.ConversationTypeDM,
		MemberIDs: []string{
			"owner-1",
			"assistant-1",
		},
		TargetID: "assistant-1",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	bridge := NewAgentBridge(msgRepo, convRepo, nil)
	content, err := json.Marshal(a2a.AssistantMessage{
		Type:    a2a.TypeAssistantMessage,
		Content: "agent finished the workspace check",
	})
	if err != nil {
		t.Fatalf("marshal assistant message: %v", err)
	}

	bridge.OnMessage(fakeBridgeSession{
		workspaceID:    "ws-bridge",
		memberID:       "assistant-1",
		memberName:     "Assistant",
		conversationID: conv.ID,
	}, &a2a.ACPMessage{
		Type:    a2a.TypeAssistantMessage,
		Content: content,
	})

	messages, err := msgRepo.ListByConversation(conv.ID, 20, "")
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(messages))
	}
	if !messages[0].IsAI || messages[0].SenderID != "assistant-1" || messages[0].Content.Text != "agent finished the workspace check" {
		t.Fatalf("unexpected message: %#v", messages[0])
	}

	updated, err := convRepo.GetByID(conv.ID)
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if updated.LastMessagePreview != "agent finished the workspace check" {
		t.Fatalf("last preview = %q", updated.LastMessagePreview)
	}
	if updated.LastMessageAt != messages[0].CreatedAt {
		t.Fatalf("last message at = %d, want %d", updated.LastMessageAt, messages[0].CreatedAt)
	}
}
