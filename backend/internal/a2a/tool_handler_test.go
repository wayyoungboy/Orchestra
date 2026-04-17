package a2a

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
)

// testRoot returns the backend/ directory root for resolving relative paths in tests.
func testRoot() string {
	wd, _ := os.Getwd()
	root := wd
	for i := 0; i < 2; i++ {
		root = filepath.Dir(root)
	}
	return root
}

// setupTestDB creates a file-based SQLite database with all migrations applied.
func setupTestDB(t *testing.T) *storage.Database {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	migrationsPath := filepath.Join(testRoot(), "internal", "storage", "migrations")
	if err := db.Migrate(migrationsPath); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return db
}

// testEnv sets up a single database with all repos and test data.
type testEnv struct {
	db         *storage.Database
	msgRepo    *repository.MessageRepository
	taskRepo   *repository.TaskRepo
	memberRepo repository.MemberRepository
	convRepo   *repository.ConversationRepository
	wsRepo     repository.WorkspaceRepository
	browser    *filesystem.Browser
	validator  *filesystem.Validator
	workspace  *models.Workspace
	secretary  *models.Member
	assistant  *models.Member
	conversation *repository.Conversation
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	db := setupTestDB(t)

	msgRepo := repository.NewMessageRepository(db.DB())
	taskRepo := repository.NewTaskRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	wsRepo := repository.NewWorkspaceRepository(db.DB())
	validator := filesystem.NewValidator([]string{"."})
	browser := filesystem.NewBrowser(validator)

	ctx := context.Background()

	ws := &models.Workspace{
		ID:   "test_ws",
		Name: "Test Workspace",
		Path: ".",
	}
	if err := wsRepo.Create(ctx, ws); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	secretary := &models.Member{
		ID:          "sec_001",
		WorkspaceID: ws.ID,
		Name:        "Secretary",
		RoleType:    models.RoleSecretary,
		Status:      "online",
		CreatedAt:   time.Now(),
	}
	if err := memberRepo.Create(ctx, secretary); err != nil {
		t.Fatalf("create secretary: %v", err)
	}

	assistant := &models.Member{
		ID:          "asst_001",
		WorkspaceID: ws.ID,
		Name:        "Assistant",
		RoleType:    models.RoleAssistant,
		ACPEnabled:  true,
		ACPCommand:  "echo assistant",
		Status:      "online",
		CreatedAt:   time.Now(),
	}
	if err := memberRepo.Create(ctx, assistant); err != nil {
		t.Fatalf("create assistant: %v", err)
	}

	conv, err := convRepo.Create(ws.ID, repository.ConversationCreate{
		Type:      repository.ConversationTypeChannel,
		MemberIDs: []string{"sec_001", "asst_001"},
		Name:      "Test Conversation",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	return &testEnv{
		db:           db,
		msgRepo:      msgRepo,
		taskRepo:     taskRepo,
		memberRepo:   memberRepo,
		convRepo:     convRepo,
		wsRepo:       wsRepo,
		browser:      browser,
		validator:    validator,
		workspace:    ws,
		secretary:    secretary,
		assistant:    assistant,
		conversation: conv,
	}
}

// trackingPool is a test double for Pool that captures SendUserMessage calls.
type trackingPool struct {
	mu           sync.Mutex
	sentMessages []string
	sessions     map[string]*Session // "workspaceID:memberID" -> session
	acquireCalls []SessionConfig
}

func newTrackingPool() *trackingPool {
	return &trackingPool{
		sessions: make(map[string]*Session),
	}
}

func (p *trackingPool) SessionForWorkspaceMember(workspaceID, memberID string) *Session {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := workspaceID + ":" + memberID
	return p.sessions[key]
}

func (p *trackingPool) AddSessionWithCapture(workspaceID, memberID, memberName string) {
	sess := &Session{
		ID:          "sess_" + memberID,
		WorkspaceID: workspaceID,
		MemberID:    memberID,
		MemberName:  memberName,
		DoneChan:    make(chan struct{}),
		OutputChan:  make(chan *ACPMessage, 16),
		ErrorChan:   make(chan error, 16),
	}
	sess.localRunner = &capturingRunner{
		capture: func(content string) {
			p.mu.Lock()
			p.sentMessages = append(p.sentMessages, content)
			p.mu.Unlock()
		},
	}
	key := workspaceID + ":" + memberID
	p.sessions[key] = sess
}

func (p *trackingPool) Acquire(ctx context.Context, config SessionConfig) (*Session, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.acquireCalls = append(p.acquireCalls, config)

	sess := &Session{
		ID:          "acquired_" + config.MemberID,
		WorkspaceID: config.WorkspaceID,
		MemberID:    config.MemberID,
		MemberName:  config.MemberName,
		DoneChan:    make(chan struct{}),
		OutputChan:  make(chan *ACPMessage, 16),
		ErrorChan:   make(chan error, 16),
	}
	sess.localRunner = &capturingRunner{
		capture: func(content string) {
			p.mu.Lock()
			p.sentMessages = append(p.sentMessages, content)
			p.mu.Unlock()
		},
	}
	return sess, nil
}

func (p *trackingPool) getSentMessages() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]string, len(p.sentMessages))
	copy(result, p.sentMessages)
	return result
}

// capturingRunner is a minimal LocalRunner stand-in for tests.
type capturingRunner struct {
	capture func(content string)
}

func (c *capturingRunner) SendUserMessage(text string) error {
	c.capture(text)
	return nil
}

func (c *capturingRunner) SendToolResult(toolUseID, content string, isError bool) error {
	return nil
}

func (c *capturingRunner) Stop() {}

func TestHandleTaskCreate_DispatchesToExistingSession(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	toolHandler := NewToolHandler(env.msgRepo, env.taskRepo, env.memberRepo, env.convRepo, nil, env.browser, env.validator)
	toolHandler.SetWorkspaceRepo(env.wsRepo)

	pool := newTrackingPool()
	pool.AddSessionWithCapture(env.workspace.ID, env.assistant.ID, env.assistant.Name)
	toolHandler.SetPool(pool)

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	toolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskCreate,
		ToolUseID: "tool_001",
		Input: json.RawMessage(`{
			"conversationId": "CONV_ID",
			"title": "Implement login",
			"description": "Add OAuth2 login support",
			"assigneeId": "asst_001",
			"priority": 2
		}`),
	}
	// Patch conversation ID
	inputMap := make(map[string]any)
	_ = json.Unmarshal(toolUse.Input, &inputMap)
	inputMap["conversationId"] = env.conversation.ID
	toolUse.Input, _ = json.Marshal(inputMap)

	result := toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
	if result.IsError {
		t.Fatalf("handleTaskCreate error: %s", result.Content)
	}

	time.Sleep(300 * time.Millisecond)

	messages := pool.getSentMessages()
	if len(messages) == 0 {
		t.Fatal("expected task dispatch message, got none")
	}

	msg := messages[0]
	if !containsAll(msg, "#conversationId{"+env.conversation.ID+"}", "#taskId{", "[秘书分配任务]", "Implement login") {
		t.Errorf("dispatch message format incorrect, got: %s", msg)
	}
	if !containsAll(msg, "Add OAuth2 login support") {
		t.Errorf("dispatch message missing description, got: %s", msg)
	}
}

func TestHandleTaskCreate_AcquiresSessionWhenMissing(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	toolHandler := NewToolHandler(env.msgRepo, env.taskRepo, env.memberRepo, env.convRepo, nil, env.browser, env.validator)
	toolHandler.SetWorkspaceRepo(env.wsRepo)

	pool := newTrackingPool()
	// No existing session for assistant — should trigger Acquire
	toolHandler.SetPool(pool)

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	toolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskCreate,
		ToolUseID: "tool_002",
		Input: json.RawMessage(`{
			"conversationId": "CONV_ID",
			"title": "Fix bug",
			"description": "Fix null pointer in auth",
			"assigneeId": "asst_001",
			"priority": 1
		}`),
	}
	inputMap := make(map[string]any)
	_ = json.Unmarshal(toolUse.Input, &inputMap)
	inputMap["conversationId"] = env.conversation.ID
	toolUse.Input, _ = json.Marshal(inputMap)

	result := toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
	if result.IsError {
		t.Fatalf("handleTaskCreate error: %s", result.Content)
	}

	time.Sleep(500 * time.Millisecond)

	messages := pool.getSentMessages()
	if len(messages) == 0 {
		t.Fatal("expected task dispatch message after Acquire, got none")
	}
	if !containsAll(messages[0], "#conversationId{"+env.conversation.ID+"}", "[秘书分配任务]", "Fix bug") {
		t.Errorf("dispatch message incorrect, got: %s", messages[0])
	}

	if len(pool.acquireCalls) == 0 {
		t.Fatal("expected Acquire to be called when no session exists")
	}
	if pool.acquireCalls[0].MemberID != "asst_001" {
		t.Errorf("expected Acquire for 'asst_001', got '%s'", pool.acquireCalls[0].MemberID)
	}
}

func TestHandleTaskCreate_SkipsDispatchWithoutAssignee(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	toolHandler := NewToolHandler(env.msgRepo, env.taskRepo, env.memberRepo, env.convRepo, nil, env.browser, env.validator)
	toolHandler.SetWorkspaceRepo(env.wsRepo)

	pool := newTrackingPool()
	toolHandler.SetPool(pool)

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	toolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskCreate,
		ToolUseID: "tool_003",
		Input: json.RawMessage(`{
			"conversationId": "CONV_ID",
			"title": "No Assignee Task",
			"description": "A task without an assignee",
			"priority": 0
		}`),
	}
	inputMap := make(map[string]any)
	_ = json.Unmarshal(toolUse.Input, &inputMap)
	inputMap["conversationId"] = env.conversation.ID
	toolUse.Input, _ = json.Marshal(inputMap)

	result := toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
	if result.IsError {
		t.Fatalf("handleTaskCreate error: %s", result.Content)
	}

	messages := pool.getSentMessages()
	if len(messages) > 0 {
		t.Errorf("expected no dispatch for task without assignee, got: %v", messages)
	}
}

func TestHandleTaskCreate_SkipsDispatchWhenACPDIsabled(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	assistantNoACP := &models.Member{
		ID:          "asst_no_acp",
		WorkspaceID: env.workspace.ID,
		Name:        "NoACP",
		RoleType:    models.RoleAssistant,
		ACPEnabled:  false,
		ACPCommand:  "",
		Status:      "online",
		CreatedAt:   time.Now(),
	}
	if err := env.memberRepo.Create(ctx, assistantNoACP); err != nil {
		t.Fatalf("create assistant: %v", err)
	}

	toolHandler := NewToolHandler(env.msgRepo, env.taskRepo, env.memberRepo, env.convRepo, nil, env.browser, env.validator)
	toolHandler.SetWorkspaceRepo(env.wsRepo)

	pool := newTrackingPool()
	toolHandler.SetPool(pool)

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	toolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskCreate,
		ToolUseID: "tool_004",
		Input: json.RawMessage(`{
			"conversationId": "CONV_ID",
			"title": "ACP Disabled Task",
			"description": "Should not dispatch to ACP-disabled assistant",
			"assigneeId": "asst_no_acp",
			"priority": 1
		}`),
	}
	inputMap := make(map[string]any)
	_ = json.Unmarshal(toolUse.Input, &inputMap)
	inputMap["conversationId"] = env.conversation.ID
	toolUse.Input, _ = json.Marshal(inputMap)

	result := toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
	if result.IsError {
		t.Fatalf("handleTaskCreate error: %s", result.Content)
	}

	time.Sleep(300 * time.Millisecond)

	messages := pool.getSentMessages()
	if len(messages) > 0 {
		t.Errorf("expected no dispatch for ACP-disabled assistant, got: %v", messages)
	}
}

// Helper: check if a string contains all substrings.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !containsStr(s, sub) {
			return false
		}
	}
	return true
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
