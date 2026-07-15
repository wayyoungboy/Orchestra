package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage"
)

func TestListByWorkspaceSupportsUnassignedTasks(t *testing.T) {
	db, err := storage.NewDatabase(filepath.Join(t.TempDir(), "tasks.db"))
	if err != nil {
		t.Fatalf("NewDatabase() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	ctx := context.Background()
	now := time.Now()
	workspaceRepo := NewWorkspaceRepository(db.DB())
	if err := workspaceRepo.Create(ctx, &models.Workspace{ID: "workspace", Name: "Workspace", Path: t.TempDir(), LastOpenedAt: now, CreatedAt: now}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	memberRepo := NewMemberRepository(db.DB())
	if err := memberRepo.Create(ctx, &models.Member{ID: "secretary", WorkspaceID: "workspace", Name: "Secretary", RoleType: models.RoleSecretary, Status: "online", CreatedAt: now}); err != nil {
		t.Fatalf("create secretary: %v", err)
	}
	convRepo := NewConversationRepository(db.DB())
	conversation, err := convRepo.Create("workspace", ConversationCreate{Type: ConversationTypeChannel, MemberIDs: []string{"secretary"}, Name: "general"})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	taskRepo := NewTaskRepository(db.DB())
	task := &models.Task{
		ID:             "task-unassigned",
		WorkspaceID:    "workspace",
		ConversationID: conversation.ID,
		SecretaryID:    "secretary",
		Title:          "Unassigned task",
		Status:         models.TaskStatusPending,
		Version:        1,
		CreatedAt:      now.UnixMilli(),
		UpdatedAt:      now.UnixMilli(),
	}
	if err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	tasks, err := taskRepo.ListByWorkspace(ctx, "workspace")
	if err != nil {
		t.Fatalf("ListByWorkspace() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != task.ID || tasks[0].AssigneeID != "" {
		t.Fatalf("ListByWorkspace() = %#v, want unassigned task", tasks)
	}
}
