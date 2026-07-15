package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
)

type workspaceScopeFixture struct {
	conversationHandler *ConversationHandler
	attachmentHandler   *AttachmentHandler
	taskHandler         *TaskHandler
	convRepo            *repository.ConversationRepository
	msgRepo             *repository.MessageRepository
	taskRepo            *repository.TaskRepo
	workspaceAID        string
	workspaceBID        string
	conversationBID     string
	taskBID             string
}

func newWorkspaceScopeFixture(t *testing.T) *workspaceScopeFixture {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	db, err := storage.NewDatabase(filepath.Join(t.TempDir(), "orchestra.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(filepath.Join(filepath.Dir(sourceFile), "..", "..", "storage", "migrations")); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	ctx := context.Background()
	workspaceRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	msgRepo := repository.NewMessageRepository(db.DB())
	readRepo := repository.NewConversationReadRepository(db.DB())
	attachmentRepo := repository.NewAttachmentRepository(db.DB())
	taskRepo := repository.NewTaskRepository(db.DB())

	workspaceA := &models.Workspace{ID: "workspace-a", Name: "A", Path: t.TempDir(), LastOpenedAt: time.Now(), CreatedAt: time.Now()}
	workspaceB := &models.Workspace{ID: "workspace-b", Name: "B", Path: t.TempDir(), LastOpenedAt: time.Now(), CreatedAt: time.Now()}
	for _, workspace := range []*models.Workspace{workspaceA, workspaceB} {
		if err := workspaceRepo.Create(ctx, workspace); err != nil {
			t.Fatalf("create workspace %s: %v", workspace.ID, err)
		}
	}
	ownerA := &models.Member{ID: "owner-a", WorkspaceID: workspaceA.ID, Name: "Owner A", RoleType: models.RoleOwner, Status: "online", CreatedAt: time.Now()}
	secretaryB := &models.Member{ID: "secretary-b", WorkspaceID: workspaceB.ID, Name: "Secretary B", RoleType: models.RoleSecretary, Status: "online", CreatedAt: time.Now()}
	assistantB := &models.Member{ID: "assistant-b", WorkspaceID: workspaceB.ID, Name: "Assistant B", RoleType: models.RoleAssistant, Status: "online", CreatedAt: time.Now()}
	for _, member := range []*models.Member{ownerA, secretaryB, assistantB} {
		if err := memberRepo.Create(ctx, member); err != nil {
			t.Fatalf("create member %s: %v", member.ID, err)
		}
	}
	conversationB, err := convRepo.Create(workspaceB.ID, repository.ConversationCreate{
		Type:      repository.ConversationTypeChannel,
		MemberIDs: []string{secretaryB.ID, assistantB.ID},
		Name:      "private-b",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}
	if _, err := msgRepo.Create(repository.MessageCreate{ConversationID: conversationB.ID, SenderID: secretaryB.ID, Content: repository.MessageContent{Type: "text", Text: "secret"}}); err != nil {
		t.Fatalf("create message: %v", err)
	}
	task := models.NewTask(models.TaskCreate{
		WorkspaceID: workspaceB.ID, ConversationID: conversationB.ID, SecretaryID: secretaryB.ID,
		AssigneeID: assistantB.ID, Title: "Private task",
	})
	if err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	return &workspaceScopeFixture{
		conversationHandler: NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, workspaceRepo, nil, nil, "", false, ""),
		attachmentHandler:   NewAttachmentHandler(msgRepo, convRepo, attachmentRepo, memberRepo, t.TempDir()),
		taskHandler:         NewTaskHandler(taskRepo, memberRepo, workspaceRepo, convRepo, nil),
		convRepo:            convRepo,
		msgRepo:             msgRepo,
		taskRepo:            taskRepo,
		workspaceAID:        workspaceA.ID,
		workspaceBID:        workspaceB.ID,
		conversationBID:     conversationB.ID,
		taskBID:             task.ID,
	}
}

func scopedTestContext(method, workspaceID, conversationID string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, "/", nil)
	ctx.Params = gin.Params{{Key: "id", Value: workspaceID}}
	if conversationID != "" {
		ctx.Params = append(ctx.Params, gin.Param{Key: "convId", Value: conversationID})
	}
	return ctx, recorder
}

func TestConversationHandlersRejectCrossWorkspaceConversation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fixture := newWorkspaceScopeFixture(t)

	tests := []struct {
		name string
		call func(*gin.Context)
	}{
		{"get conversation", fixture.conversationHandler.GetConversation},
		{"get messages", fixture.conversationHandler.GetMessages},
		{"update settings", fixture.conversationHandler.UpdateSettings},
		{"clear messages", fixture.conversationHandler.ClearMessages},
		{"delete conversation", fixture.conversationHandler.Delete},
		{"mark read", fixture.conversationHandler.MarkConversationRead},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := scopedTestContext(http.MethodGet, fixture.workspaceAID, fixture.conversationBID)
			tt.call(ctx)
			if recorder.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
			}
		})
	}
	if _, err := fixture.convRepo.GetByID(fixture.conversationBID); err != nil {
		t.Fatalf("cross-workspace delete changed the target conversation: %v", err)
	}
	messages, err := fixture.msgRepo.ListByConversation(fixture.conversationBID, 10, "")
	if err != nil || len(messages) != 1 {
		t.Fatalf("cross-workspace clear changed target messages: count=%d err=%v", len(messages), err)
	}
}

func TestAttachmentAndTaskHandlersRejectCrossWorkspaceResources(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fixture := newWorkspaceScopeFixture(t)

	ctx, recorder := scopedTestContext(http.MethodGet, fixture.workspaceAID, fixture.conversationBID)
	ctx.Request.URL.RawQuery = "conversationId=" + fixture.conversationBID
	fixture.attachmentHandler.ListAttachments(ctx)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("attachment list status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	ctx, recorder = scopedTestContext(http.MethodGet, fixture.workspaceAID, "")
	ctx.Params = append(ctx.Params, gin.Param{Key: "taskId", Value: fixture.taskBID})
	fixture.taskHandler.GetTask(ctx)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("task get status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	ctx, recorder = scopedTestContext(http.MethodPost, fixture.workspaceAID, "")
	ctx.Params = append(ctx.Params, gin.Param{Key: "taskId", Value: fixture.taskBID})
	fixture.taskHandler.CancelTask(ctx)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("task cancel status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	task, err := fixture.taskRepo.GetByID(context.Background(), fixture.taskBID)
	if err != nil || task.Status != models.TaskStatusAssigned {
		t.Fatalf("cross-workspace cancel changed task: status=%q err=%v", task.Status, err)
	}
}

func TestAttachmentFilenameMatchesIDExactly(t *testing.T) {
	if attachmentFilenameMatchesID("attachment-123.txt", "attachment-12") {
		t.Fatal("a prefix must not match a different attachment ID")
	}
	if !attachmentFilenameMatchesID("attachment-123.txt", "attachment-123") {
		t.Fatal("the full attachment ID should match")
	}
}
