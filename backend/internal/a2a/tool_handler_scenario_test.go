package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

// mockBroadcaster captures all broadcast events for assertion.
type mockBroadcaster struct {
	mu      sync.Mutex
	events  []broadcastEvent
}

type broadcastEvent struct {
	workspaceID string
	event       interface{}
}

func (b *mockBroadcaster) BroadcastToWorkspace(workspaceID string, event interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, broadcastEvent{workspaceID: workspaceID, event: event})
}

func (b *mockBroadcaster) getEvents() []broadcastEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]broadcastEvent, len(b.events))
	copy(result, b.events)
	return result
}

// scenarioEnv extends testEnv with scenario-specific fixtures.
type scenarioEnv struct {
	*testEnv
	broadcaster *mockBroadcaster
	toolHandler *ToolHandler
	pool        *scenarioPool
	asstFiles   *models.Member // Assistant specialized in file analysis
	asstCount   *models.Member // Assistant specialized in code counting
}

// scenarioPool extends trackingPool with synthetic response injection.
type scenarioPool struct {
	*trackingPool
	// syntheticResponses maps "memberID" -> list of responses to inject when that member receives a message
	syntheticResponses map[string][]syntheticResponse
	responseIndex      map[string]int // tracks which response to use next per member
}

type syntheticResponse struct {
	// What the assistant "says back" after receiving a task notification
	toolCalls []scenarioToolCall // tool calls the assistant would make
}

type scenarioToolCall struct {
	toolName string
	input    map[string]any
	// After this tool call, what does the secretary respond?
	toolResult string
	isError    bool
}

func newScenarioPool() *scenarioPool {
	return &scenarioPool{
		trackingPool:       newTrackingPool(),
		syntheticResponses: make(map[string][]syntheticResponse),
		responseIndex:      make(map[string]int),
	}
}

// setupScenario creates a workspace with a secretary and two assistants.
func setupScenario(t *testing.T) *scenarioEnv {
	t.Helper()
	env := setupTestEnv(t)

	// Add a second assistant (specialized in code counting)
	asstFiles := &models.Member{
		ID:          "asst_files",
		WorkspaceID: env.workspace.ID,
		Name:        "FileAnalyzer",
		RoleType:    models.RoleAssistant,
		ACPEnabled:  true,
		ACPCommand:  "echo files",
		Status:      "online",
		CreatedAt:   time.Now(),
	}
	if err := env.memberRepo.Create(context.Background(), asstFiles); err != nil {
		t.Fatalf("create file analyzer assistant: %v", err)
	}

	asstCount := &models.Member{
		ID:          "asst_count",
		WorkspaceID: env.workspace.ID,
		Name:        "CodeCounter",
		RoleType:    models.RoleAssistant,
		ACPEnabled:  true,
		ACPCommand:  "echo count",
		Status:      "online",
		CreatedAt:   time.Now(),
	}
	if err := env.memberRepo.Create(context.Background(), asstCount); err != nil {
		t.Fatalf("create code counter assistant: %v", err)
	}

	broadcaster := &mockBroadcaster{}
	toolHandler := NewToolHandler(env.msgRepo, env.taskRepo, env.memberRepo, env.convRepo, broadcaster, env.browser, env.validator)
	toolHandler.SetWorkspaceRepo(env.wsRepo)

	pool := newScenarioPool()
	toolHandler.SetPool(pool)

	return &scenarioEnv{
		testEnv:       env,
		broadcaster:   broadcaster,
		toolHandler:   toolHandler,
		pool:          pool,
		asstFiles:     asstFiles,
		asstCount:     asstCount,
	}
}

// injectSecretaryResponse configures the secretary's behavior when it receives a user message.
// The secretary will dispatch the specified tool calls in sequence.
func injectSecretaryResponse(pool *scenarioPool, memberID string, responses []syntheticResponse) {
	pool.syntheticResponses[memberID] = responses
	pool.responseIndex[memberID] = 0
}

// simulateSecretaryReceivingMessage simulates the secretary receiving a user message
// and dispatching tool calls as defined in the synthetic response.
func simulateSecretaryReceivingMessage(ctx context.Context, env *scenarioEnv, secretarySession *Session, msg string) []toolExecutionRecord {
	var records []toolExecutionRecord

	responses := env.pool.syntheticResponses[env.secretary.ID]
	idx := env.pool.responseIndex[env.secretary.ID]
	if idx >= len(responses) {
		return records
	}
	resp := responses[idx]
	env.pool.responseIndex[env.secretary.ID] = idx + 1

	// Execute each tool call the secretary would make
	for _, tc := range resp.toolCalls {
		toolUse := &ToolUseMessage{
			Type:      TypeToolUse,
			Name:      tc.toolName,
			ToolUseID: fmt.Sprintf("tool_%d", len(records)+1),
		}
		inputBytes, _ := json.Marshal(tc.input)
		toolUse.Input = json.RawMessage(inputBytes)

		// Handle the tool call
		var result *ToolResult
		switch tc.toolName {
		case ToolTaskCreate:
			result = env.toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
		case ToolChatSend:
			result = env.toolHandler.handleChatSend(ctx, toolUse, secretarySession)
		case ToolWorkloadList:
			result = env.toolHandler.handleWorkloadList(ctx, toolUse, secretarySession)
		}

		records = append(records, toolExecutionRecord{
			toolName: tc.toolName,
			input:    tc.input,
			result:   result,
		})

		// Wait for async dispatch to complete
		if tc.toolName == ToolTaskCreate {
			time.Sleep(300 * time.Millisecond)
		}
	}

	return records
}

type toolExecutionRecord struct {
	toolName string
	input    map[string]any
	result   *ToolResult
}

// simulateAssistantTaskFlow simulates an assistant receiving a task notification,
// starting the task, and completing it with a result.
func simulateAssistantTaskFlow(ctx context.Context, env *scenarioEnv, taskID string, assistantSession *Session, resultSummary string) error {
	// Step 1: Assistant starts the task
	startToolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskStart,
		ToolUseID: "tool_start_" + taskID,
		Input:     json.RawMessage(fmt.Sprintf(`{"taskId":"%s","message":"Starting work"}`, taskID)),
	}
	result := env.toolHandler.handleTaskStart(ctx, startToolUse, assistantSession)
	if result.IsError {
		return fmt.Errorf("failed to start task: %s", result.Content)
	}

	// Step 2: Assistant completes the task
	completeToolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskComplete,
		ToolUseID: "tool_complete_" + taskID,
		Input:     json.RawMessage(fmt.Sprintf(`{"taskId":"%s","resultSummary":"%s"}`, taskID, resultSummary)),
	}
	result = env.toolHandler.handleTaskComplete(ctx, completeToolUse, assistantSession)
	if result.IsError {
		return fmt.Errorf("failed to complete task: %s", result.Content)
	}

	return nil
}

// extractTaskIDFromDispatchMessage parses the taskId from a dispatch message.
func extractTaskIDFromDispatchMessage(msg string) string {
	// Format: #conversationId{...}#taskId{task_xxx}[秘书分配任务] ...
	start := strings.Index(msg, "#taskId{")
	if start == -1 {
		return ""
	}
	start += len("#taskId{")
	end := strings.Index(msg[start:], "}")
	if end == -1 {
		return ""
	}
	return msg[start : start+end]
}

func TestScenario_CodebaseSummaryWorkflow(t *testing.T) {
	env := setupScenario(t)
	ctx := context.Background()

	// Create secretary session
	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	// Create assistant sessions (pre-register them)
	env.pool.AddSessionWithCapture(env.workspace.ID, env.assistant.ID, env.assistant.Name)
	env.pool.AddSessionWithCapture(env.workspace.ID, "asst_files", "FileAnalyzer")
	env.pool.AddSessionWithCapture(env.workspace.ID, "asst_count", "CodeCounter")

	// ============================================================
	// Phase 1: User asks secretary to summarize the codebase
	// ============================================================

	// Secretary's behavior: analyze workload, then dispatch 2 tasks
	injectSecretaryResponse(env.pool, env.secretary.ID, []syntheticResponse{
		{
			toolCalls: []scenarioToolCall{
				{
					// Check workload first
					toolName: ToolWorkloadList,
					input:    map[string]any{"workspaceId": env.workspace.ID},
				},
				{
					// Task 1: Analyze project structure
					toolName: ToolTaskCreate,
					input: map[string]any{
						"conversationId": env.conversation.ID,
						"title":          "分析项目结构",
						"description":    "列出项目的主要目录结构和关键文件，统计文件数量",
						"assigneeId":     "asst_files",
						"priority":       1,
					},
				},
				{
					// Task 2: Count code lines
					toolName: ToolTaskCreate,
					input: map[string]any{
						"conversationId": env.conversation.ID,
						"title":          "统计代码行数",
						"description":    "统计所有Go和TypeScript文件的代码行数",
						"assigneeId":     "asst_count",
						"priority":       1,
					},
				},
			},
		},
	})

	// Secretary receives the user's request
	records := simulateSecretaryReceivingMessage(ctx, env, secretarySession, "请总结一下这个项目有多少代码")

	// Verify Phase 1: Secretary made the expected tool calls
	if len(records) != 3 {
		t.Fatalf("expected 3 tool calls (workload + 2 tasks), got %d", len(records))
	}

	if records[0].toolName != ToolWorkloadList {
		t.Errorf("first tool call should be workload list, got %s", records[0].toolName)
	}

	createdTaskIDs := []string{}
	for _, r := range records {
		if r.toolName == ToolTaskCreate {
			if r.result.IsError {
				t.Errorf("task creation failed: %s", r.result.Content)
			}
			// Extract taskId from result
			var resultData map[string]any
			if err := json.Unmarshal([]byte(r.result.Content), &resultData); err == nil {
				if tid, ok := resultData["taskId"].(string); ok {
					createdTaskIDs = append(createdTaskIDs, tid)
				}
			}
		}
	}

	if len(createdTaskIDs) != 2 {
		t.Fatalf("expected 2 tasks created, got %d", len(createdTaskIDs))
	}

	// ============================================================
	// Phase 2: Verify tasks were dispatched to assistants
	// ============================================================

	// Filter messages sent to FileAnalyzer (asst_files)
	var fileAnalyzerDispatches []string
	var codeCounterDispatches []string

	// The tracking pool captures messages via the capturingRunner
	// We need to check which assistant received which message
	// Since all messages go through the same capture mechanism, check the session mapping
	// Actually, the trackingPool stores sessions by key, and the capturingRunner captures all messages

	// Get all sent messages and classify by content
	allMessages := env.pool.getSentMessages()

	for _, msg := range allMessages {
		if strings.Contains(msg, "分析项目结构") || strings.Contains(msg, "主要目录结构") {
			fileAnalyzerDispatches = append(fileAnalyzerDispatches, msg)
		}
		if strings.Contains(msg, "统计代码行数") || strings.Contains(msg, "代码行数") {
			codeCounterDispatches = append(codeCounterDispatches, msg)
		}
	}

	if len(fileAnalyzerDispatches) == 0 {
		t.Error("expected FileAnalyzer to receive task dispatch message")
	}
	if len(codeCounterDispatches) == 0 {
		t.Error("expected CodeCounter to receive task dispatch message")
	}

	// ============================================================
	// Phase 3: Verify tasks are in the database
	// ============================================================

	for _, taskID := range createdTaskIDs {
		task, err := env.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			t.Errorf("task %s not found in database: %v", taskID, err)
			continue
		}
		if task.Status != models.TaskStatusAssigned && task.Status != models.TaskStatusPending {
			t.Errorf("task %s should be pending/assigned, got status=%s", taskID, task.Status)
		}
	}

	// ============================================================
	// Phase 4: Assistants execute their tasks
	// ============================================================

	// Simulate FileAnalyzer receiving and executing its task
	fileAnalyzerSession := env.pool.SessionForWorkspaceMember(env.workspace.ID, "asst_files")
	if fileAnalyzerSession == nil {
		t.Fatal("FileAnalyzer session not found")
	}

	for _, taskID := range createdTaskIDs {
		task, err := env.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			continue
		}
		if task.AssigneeID == "asst_files" {
			err := simulateAssistantTaskFlow(ctx, env, taskID, fileAnalyzerSession,
				"项目包含32个Go文件, 18个TypeScript文件。主要目录: backend/cmd/, backend/internal/, frontend/src/")
			if err != nil {
				t.Errorf("FileAnalyzer task flow failed: %v", err)
			}
		}
		if task.AssigneeID == "asst_count" {
			codeCounterSession := env.pool.SessionForWorkspaceMember(env.workspace.ID, "asst_count")
			if codeCounterSession == nil {
				t.Fatal("CodeCounter session not found")
			}
			err := simulateAssistantTaskFlow(ctx, env, taskID, codeCounterSession,
				"总计: Go文件4,521行, TypeScript文件6,234行, 总共10,755行代码")
			if err != nil {
				t.Errorf("CodeCounter task flow failed: %v", err)
			}
		}
	}

	// ============================================================
	// Phase 5: Verify tasks completed in database
	// ============================================================

	for _, taskID := range createdTaskIDs {
		task, err := env.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			t.Errorf("task %s not found after completion: %v", taskID, err)
			continue
		}
		if task.Status != models.TaskStatusCompleted {
			t.Errorf("task %s should be completed, got status=%s", taskID, task.Status)
		}
		if task.ResultSummary == "" {
			t.Errorf("task %s should have a result summary", taskID)
		}
		if task.StartedAt == 0 {
			t.Errorf("task %s should have a started_at timestamp", taskID)
		}
		if task.CompletedAt == 0 {
			t.Errorf("task %s should have a completed_at timestamp", taskID)
		}
	}

	// ============================================================
	// Phase 6: Secretary sends final summary
	// ============================================================

	// Inject secretary's final summary behavior
	injectSecretaryResponse(env.pool, env.secretary.ID, []syntheticResponse{
		{
			toolCalls: []scenarioToolCall{
				{
					toolName: ToolChatSend,
					input: map[string]any{
						"conversationId": env.conversation.ID,
						"text":           "代码库总结：项目共10,755行代码（Go 4,521行 + TypeScript 6,234行），包含50个源文件。主要结构：backend/cmd/（入口）、backend/internal/（核心逻辑）、frontend/src/（前端界面）。",
					},
				},
			},
		},
	})

	// Secretary sends the summary
	summaryRecords := simulateSecretaryReceivingMessage(ctx, env, secretarySession, "总结任务结果")

	if len(summaryRecords) != 1 {
		t.Fatalf("expected 1 tool call (chat_send) for summary, got %d", len(summaryRecords))
	}
	if summaryRecords[0].toolName != ToolChatSend {
		t.Errorf("expected chat_send for summary, got %s", summaryRecords[0].toolName)
	}
	if summaryRecords[0].result.IsError {
		t.Errorf("summary chat_send failed: %s", summaryRecords[0].result.Content)
	}

	// ============================================================
	// Phase 7: Verify the summary message was broadcast
	// ============================================================

	broadcastEvents := env.broadcaster.getEvents()
	summaryFound := false
	for _, evt := range broadcastEvents {
		evtMap, ok := evt.event.(map[string]interface{})
		if !ok {
			continue
		}
		if msgType, _ := evtMap["type"].(string); msgType == "new_message" {
			if content, _ := evtMap["content"].(string); strings.Contains(content, "代码库总结") {
				summaryFound = true
				break
			}
		}
	}

	if !summaryFound {
		t.Error("expected summary message to be broadcast to workspace")
	}

	// ============================================================
	// Phase 8: Verify all messages are persisted in DB
	// ============================================================

	messages, err := env.msgRepo.ListByConversation(env.conversation.ID, 100, "")
	if err != nil {
		t.Fatalf("failed to list messages: %v", err)
	}

	// Should have: user request + summary message from secretary
	var secretaryMessages []repository.Message
	for _, msg := range messages {
		if msg.SenderID == env.secretary.ID {
			secretaryMessages = append(secretaryMessages, msg)
		}
	}

	if len(secretaryMessages) == 0 {
		t.Error("expected at least one message from secretary in the conversation")
	}
}

// TestScenario_TaskFailureRecovery tests the scenario where an assistant fails a task.
func TestScenario_TaskFailureRecovery(t *testing.T) {
	env := setupScenario(t)
	ctx := context.Background()

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	env.pool.AddSessionWithCapture(env.workspace.ID, env.assistant.ID, env.assistant.Name)

	// Secretary creates a task
	toolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskCreate,
		ToolUseID: "tool_fail_test",
		Input: json.RawMessage(fmt.Sprintf(`{
			"conversationId": "%s",
			"title": "Impossible Task",
			"description": "Do something that will fail",
			"assigneeId": "%s",
			"priority": 3
		}`, env.conversation.ID, env.assistant.ID)),
	}

	result := env.toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
	if result.IsError {
		t.Fatalf("task creation failed: %s", result.Content)
	}

	time.Sleep(300 * time.Millisecond)

	// Extract task ID from result
	var resultData map[string]any
	if err := json.Unmarshal([]byte(result.Content), &resultData); err != nil {
		t.Fatalf("failed to parse task creation result: %v", err)
	}
	taskID, _ := resultData["taskId"].(string)
	if taskID == "" {
		t.Fatal("no task ID in creation result")
	}

	// Assistant attempts the task but fails
	assistantSession := env.pool.SessionForWorkspaceMember(env.workspace.ID, env.assistant.ID)
	if assistantSession == nil {
		t.Fatal("assistant session not found")
	}

	failToolUse := &ToolUseMessage{
		Type:      TypeToolUse,
		Name:      ToolTaskFail,
		ToolUseID: "tool_fail_" + taskID,
		Input:     json.RawMessage(fmt.Sprintf(`{"taskId":"%s","errorMessage":"无法完成：缺少必要依赖"}`, taskID)),
	}

	failResult := env.toolHandler.handleTaskFail(ctx, failToolUse, assistantSession)
	if failResult.IsError {
		t.Errorf("task fail reporting failed: %s", failResult.Content)
	}

	// Verify task is marked as failed in DB
	task, err := env.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		t.Fatalf("task not found: %v", err)
	}

	if task.Status != models.TaskStatusFailed {
		t.Errorf("task should be failed, got status=%s", task.Status)
	}
	if task.ErrorMessage == "" {
		t.Error("failed task should have an error message")
	}
}

// TestScenario_MultipleAssistantsSequentialTasks tests dispatching multiple tasks to different
// assistants and verifying the complete lifecycle of each.
func TestScenario_MultipleAssistantsSequentialTasks(t *testing.T) {
	env := setupScenario(t)
	ctx := context.Background()

	secretarySession := &Session{
		ID:          "sec_session",
		WorkspaceID: env.workspace.ID,
		MemberID:    env.secretary.ID,
		MemberName:  env.secretary.Name,
	}

	env.pool.AddSessionWithCapture(env.workspace.ID, "asst_files", "FileAnalyzer")
	env.pool.AddSessionWithCapture(env.workspace.ID, "asst_count", "CodeCounter")

	// Create 3 tasks sequentially
	taskConfigs := []struct {
		title      string
		desc       string
		assigneeID string
		priority   int
	}{
		{"Task A", "First task for FileAnalyzer", "asst_files", 1},
		{"Task B", "Second task for CodeCounter", "asst_count", 2},
		{"Task C", "Third task for FileAnalyzer", "asst_files", 3},
	}

	var taskIDs []string
	for _, cfg := range taskConfigs {
		toolUse := &ToolUseMessage{
			Type:      TypeToolUse,
			Name:      ToolTaskCreate,
			ToolUseID: fmt.Sprintf("tool_seq_%s", cfg.title),
			Input: json.RawMessage(fmt.Sprintf(`{
				"conversationId": "%s",
				"title": "%s",
				"description": "%s",
				"assigneeId": "%s",
				"priority": %d
			}`, env.conversation.ID, cfg.title, cfg.desc, cfg.assigneeID, cfg.priority)),
		}

		result := env.toolHandler.handleTaskCreate(ctx, toolUse, secretarySession)
		if result.IsError {
			t.Fatalf("task creation failed for %s: %s", cfg.title, result.Content)
		}

		time.Sleep(300 * time.Millisecond)

		var resultData map[string]any
		_ = json.Unmarshal([]byte(result.Content), &resultData)
		tid, _ := resultData["taskId"].(string)
		taskIDs = append(taskIDs, tid)
	}

	// Verify all tasks exist and are assigned
	for i, taskID := range taskIDs {
		task, err := env.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			t.Fatalf("task %d not found: %v", i, err)
		}
		expectedAssignee := taskConfigs[i].assigneeID
		if task.AssigneeID != expectedAssignee {
			t.Errorf("task %d assignee: expected %s, got %s", i, expectedAssignee, task.AssigneeID)
		}
		if task.Status != models.TaskStatusAssigned {
			t.Errorf("task %d status: expected assigned, got %s", i, task.Status)
		}
	}

	// Verify dispatch messages were sent
	messages := env.pool.getSentMessages()
	fileAnalyzerMsgCount := 0
	codeCounterMsgCount := 0
	for _, msg := range messages {
		if strings.Contains(msg, "Task A") || strings.Contains(msg, "Task C") {
			fileAnalyzerMsgCount++
		}
		if strings.Contains(msg, "Task B") {
			codeCounterMsgCount++
		}
	}

	if fileAnalyzerMsgCount != 2 {
		t.Errorf("expected 2 dispatch messages to FileAnalyzer, got %d", fileAnalyzerMsgCount)
	}
	if codeCounterMsgCount != 1 {
		t.Errorf("expected 1 dispatch message to CodeCounter, got %d", codeCounterMsgCount)
	}

	// Now complete all tasks and verify
	fileAnalyzerSession := env.pool.SessionForWorkspaceMember(env.workspace.ID, "asst_files")
	codeCounterSession := env.pool.SessionForWorkspaceMember(env.workspace.ID, "asst_count")

	assistantResults := map[string]string{
		taskIDs[0]: "FileAnalyzer completed Task A",
		taskIDs[1]: "CodeCounter completed Task B",
		taskIDs[2]: "FileAnalyzer completed Task C",
	}

	for i, taskID := range taskIDs {
		assigneeID := taskConfigs[i].assigneeID
		var sess *Session
		if assigneeID == "asst_files" {
			sess = fileAnalyzerSession
		} else {
			sess = codeCounterSession
		}

		err := simulateAssistantTaskFlow(ctx, env, taskID, sess, assistantResults[taskID])
		if err != nil {
			t.Errorf("task %s execution failed: %v", taskID, err)
		}
	}

	// Final verification: all tasks completed
	for _, taskID := range taskIDs {
		task, err := env.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			t.Errorf("task %s not found after completion: %v", taskID, err)
			continue
		}
		if task.Status != models.TaskStatusCompleted {
			t.Errorf("task %s should be completed, got %s", taskID, task.Status)
		}
		if task.ResultSummary == "" {
			t.Errorf("task %s missing result summary", taskID)
		}
	}

	// Verify workload shows all tasks completed
	stats, err := env.taskRepo.GetWorkloadStats(ctx, env.workspace.ID)
	if err != nil {
		t.Fatalf("failed to get workload stats: %v", err)
	}

	for _, memberStat := range stats {
		memberID, _ := memberStat["memberId"].(string)
		if memberID == "asst_files" || memberID == "asst_count" {
			completedCount, _ := memberStat["completedTaskCount"].(int)
			if completedCount == 0 {
				t.Errorf("member %s should have completed tasks", memberID)
			}
		}
	}
}
