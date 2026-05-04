package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

// ChatBroadcaster is an interface for broadcasting chat messages.
type ChatBroadcaster interface {
	BroadcastToWorkspace(workspaceID string, event interface{})
}

// SessionConfig contains parameters for creating a session.
type SessionConfig struct {
	ID            string
	WorkspaceID   string
	WorkspaceDir  string
	MemberID      string
	MemberName    string
	TerminalType  string
	Member        *models.Member
}

// SessionLookup finds an active session for a workspace member.
// Defined here to avoid import cycles between a2a and handler packages.
type SessionLookup interface {
	SessionForWorkspaceMember(workspaceID, memberID string) *Session
	Acquire(ctx context.Context, config SessionConfig) (*Session, error)
}

// ToolHandler executes Orchestra tools on behalf of A2A sessions.
type ToolHandler struct {
	msgRepo      *repository.MessageRepository
	taskRepo     *repository.TaskRepo
	memberRepo   repository.MemberRepository
	convRepo     *repository.ConversationRepository
	chatHub      ChatBroadcaster
	browser      *filesystem.Browser
	validator    *filesystem.Validator
	pool         SessionLookup
	workspaceRepo  repository.WorkspaceRepository

	dispatchWg     sync.WaitGroup
	dispatchCtx    context.Context
	dispatchCancel context.CancelFunc
}

func NewToolHandler(
	msgRepo *repository.MessageRepository,
	taskRepo *repository.TaskRepo,
	memberRepo repository.MemberRepository,
	convRepo *repository.ConversationRepository,
	chatHub ChatBroadcaster,
	browser *filesystem.Browser,
	validator *filesystem.Validator,
) *ToolHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &ToolHandler{
		msgRepo:        msgRepo,
		taskRepo:       taskRepo,
		memberRepo:     memberRepo,
		convRepo:       convRepo,
		chatHub:        chatHub,
		browser:        browser,
		validator:      validator,
		dispatchCtx:    ctx,
		dispatchCancel: cancel,
	}
}

// SetPool sets the session pool for task dispatch.
func (h *ToolHandler) SetPool(pool SessionLookup) {
	h.pool = pool
}

// SetWorkspaceRepo sets the workspace repo for task dispatch.
func (h *ToolHandler) SetWorkspaceRepo(repo repository.WorkspaceRepository) {
	h.workspaceRepo = repo
}

// Shutdown gracefully shuts down the tool handler and waits for pending dispatch goroutines.
func (h *ToolHandler) Shutdown(ctx context.Context) error {
	h.dispatchCancel()
	done := make(chan struct{})
	go func() { h.dispatchWg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("tool handler shutdown timeout")
	}
}

// ExecuteTool executes a tool call and returns the result.
// The caller is responsible for sending the result back to the session.
func (h *ToolHandler) ExecuteTool(msg *ACPMessage, sess *Session) *ToolResult {
	toolUse, err := msg.ParseToolUseMessage()
	if err != nil {
		return &ToolResult{
			Type:      TypeToolResult,
			ToolUseID: "",
			IsError:   true,
			Content:   fmt.Sprintf("Failed to parse tool use: %v", err),
		}
	}

	ctx := context.Background()
	var result *ToolResult

	switch toolUse.Name {
	case ToolChatSend:
		result = h.handleChatSend(ctx, toolUse, sess)
	case ToolTaskCreate:
		result = h.handleTaskCreate(ctx, toolUse, sess)
	case ToolTaskStart:
		result = h.handleTaskStart(ctx, toolUse, sess)
	case ToolTaskComplete:
		result = h.handleTaskComplete(ctx, toolUse, sess)
	case ToolTaskFail:
		result = h.handleTaskFail(ctx, toolUse, sess)
	case ToolWorkloadList:
		result = h.handleWorkloadList(ctx, toolUse, sess)
	case ToolTaskList:
		result = h.handleTaskList(ctx, toolUse, sess)
	case ToolAgentStatus:
		result = h.handleAgentStatus(ctx, toolUse, sess)
	case ToolFileRead:
		result = h.handleFileRead(ctx, toolUse, sess)
	case ToolFileWrite:
		result = h.handleFileWrite(ctx, toolUse, sess)
	case ToolFileList:
		result = h.handleFileList(ctx, toolUse, sess)
	default:
		result = &ToolResult{
			Type:      TypeToolResult,
			ToolUseID: toolUse.ToolUseID,
			IsError:   true,
			Content:   fmt.Sprintf("Unknown tool: %s", toolUse.Name),
		}
	}

	return result
}

func (h *ToolHandler) handleChatSend(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input ChatSendInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	msg, err := h.msgRepo.Create(repository.MessageCreate{
		ConversationID: input.ConversationID,
		SenderID:       sess.MemberID,
		Content:        repository.MessageContent{Type: "text", Text: input.Text},
		IsAI:           true,
	})
	if err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to create message: %v", err)}
	}

	if h.chatHub != nil {
		h.chatHub.BroadcastToWorkspace(sess.WorkspaceID, map[string]interface{}{
			"type":           "new_message",
			"workspaceId":    sess.WorkspaceID,
			"conversationId": input.ConversationID,
			"messageId":      msg.ID,
			"senderId":       sess.MemberID,
			"senderName":     sess.MemberName,
			"content":        input.Text,
			"isAi":           true,
			"createdAt":      msg.CreatedAt,
		})
	}

	return &ToolResult{
		Type:      TypeToolResult,
		ToolUseID: toolUse.ToolUseID,
		Content:   fmt.Sprintf(`{"success":true,"messageId":"%s","sentAt":%d}`, msg.ID, msg.CreatedAt),
	}
}

func (h *ToolHandler) handleTaskCreate(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskCreateInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	now := time.Now().UnixMilli()
	task := &models.Task{
		ID:             "task_" + utils.GenerateID()[:10],
		WorkspaceID:    sess.WorkspaceID,
		ConversationID: input.ConversationID,
		SecretaryID:    sess.MemberID,
		Title:          input.Title,
		Description:    input.Description,
		Status:         models.TaskStatusPending,
		Priority:       input.Priority,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if input.AssigneeID != "" {
		task.AssigneeID = input.AssigneeID
		task.Status = models.TaskStatusAssigned
		task.AssignedAt = now
	}

	if err := h.taskRepo.Create(ctx, task); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to create task: %v", err)}
	}

	// Dispatch task to assignee via their session
	if input.AssigneeID != "" && h.pool != nil {
		h.dispatchWg.Add(1)
		go h.dispatchTaskToAssignee(task)
	}

	return &ToolResult{
		Type:      TypeToolResult,
		ToolUseID: toolUse.ToolUseID,
		Content:   fmt.Sprintf(`{"success":true,"taskId":"%s","status":"%s"}`, task.ID, task.Status),
	}
}

// notifySecretary sends a task status update to the secretary who created the task.
func (h *ToolHandler) notifySecretary(task *models.Task, status string, detail string, agentSess *Session) {
	if h.pool == nil || task.SecretaryID == "" {
		return
	}

	secSess := h.pool.SessionForWorkspaceMember(task.WorkspaceID, task.SecretaryID)
	if secSess == nil {
		ctx, cancel := context.WithTimeout(h.dispatchCtx, 10*time.Second)
		defer cancel()

		member, err := h.memberRepo.GetByID(ctx, task.SecretaryID)
		if err != nil || member == nil {
			log.Printf("[a2a] notifySecretary: secretary %s not found", task.SecretaryID)
			return
		}
		if !member.ACPEnabled || member.ACPCommand == "" {
			return
		}
		workspaceDir := ""
		if h.workspaceRepo != nil {
			ws, err := h.workspaceRepo.GetByID(ctx, task.WorkspaceID)
			if err == nil && ws != nil {
				workspaceDir = ws.Path
			}
		}
		secSess, _ = h.pool.Acquire(ctx, SessionConfig{
			WorkspaceID:  task.WorkspaceID,
			WorkspaceDir: workspaceDir,
			MemberID:     member.ID,
			MemberName:   member.Name,
			TerminalType: member.TerminalType,
			Member:       member,
		})
	}
	if secSess == nil {
		return
	}

	prompt := fmt.Sprintf(
		"#conversationId{%s}#taskId{%s}[任务%s] 助手 %s 的任务「%s」已%s：%s",
		task.ConversationID, task.ID, status, agentSess.MemberName, task.Title, status, detail,
	)
	if err := secSess.SendUserMessage(prompt); err != nil {
		log.Printf("[a2a] Failed to notify secretary %s about task %s: %v", task.SecretaryID, task.ID, err)
	} else {
		log.Printf("[a2a] Notified secretary %s: task %s %s", task.SecretaryID, task.ID, status)
	}
}

// dispatchTaskToAssignee finds the assignee's session and sends a task notification.
func (h *ToolHandler) dispatchTaskToAssignee(task *models.Task) {
	defer h.dispatchWg.Done()

	// Try existing session first
	sess := h.pool.SessionForWorkspaceMember(task.WorkspaceID, task.AssigneeID)
	if sess == nil {
		// No existing session — need to acquire one via Acquire
		// Look up member config
		ctx, cancel := context.WithTimeout(h.dispatchCtx, 10*time.Second)
		defer cancel()

		member, err := h.memberRepo.GetByID(ctx, task.AssigneeID)
		if err != nil {
			log.Printf("[a2a] Task dispatch: failed to get assignee member %s: %v", task.AssigneeID, err)
			return
		}
		if member == nil {
			log.Printf("[a2a] Task dispatch: assignee member %s not found", task.AssigneeID)
			return
		}

		// Only dispatch if member has ACP enabled
		if !member.ACPEnabled || member.ACPCommand == "" {
			log.Printf("[a2a] Task dispatch: assignee %s has no ACP configured, skipping", member.Name)
			return
		}

		// Look up workspace for path info
		workspaceDir := ""
		if h.workspaceRepo != nil {
			workspace, err := h.workspaceRepo.GetByID(ctx, task.WorkspaceID)
			if err == nil && workspace != nil {
				workspaceDir = workspace.Path
			}
		}

		newSess, err := h.pool.Acquire(ctx, SessionConfig{
			WorkspaceID:  task.WorkspaceID,
			WorkspaceDir: workspaceDir,
			MemberID:     member.ID,
			MemberName:   member.Name,
			TerminalType: member.TerminalType,
			Member:       member,
		})
		if err != nil {
			log.Printf("[a2a] Task dispatch: failed to acquire session for assignee %s: %v", member.Name, err)
			return
		}
		if newSess == nil {
			log.Printf("[a2a] Task dispatch: no session created for assignee %s", member.Name)
			return
		}
		sess = newSess
	}

	// Send task notification to the assignee
	prompt := fmt.Sprintf(`#conversationId{%s}#taskId{%s}[秘书分配任务]: %s`,
		task.ConversationID,
		task.ID,
		task.Description,
	)
	if task.Title != "" {
		prompt = fmt.Sprintf(`#conversationId{%s}#taskId{%s}[秘书分配任务] %s: %s`,
			task.ConversationID,
			task.ID,
			task.Title,
			task.Description,
		)
	}

	log.Printf("[a2a] Dispatching task %s to %s", task.ID, sess.MemberName)
	if err := sess.SendUserMessage(prompt); err != nil {
		log.Printf("[a2a] Failed to dispatch task %s to %s: %v", task.ID, sess.MemberName, err)
	}
}

func (h *ToolHandler) handleTaskStart(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskStartInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "started_at": now}

	// Retry logic for concurrent updates (max 3 attempts)
	backoffs := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond}
	for i := 0; i < len(backoffs); i++ {
		if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusInProgress, updates); err != nil {
			if strings.Contains(err.Error(), "concurrent update") {
				if i < len(backoffs)-1 {
					time.Sleep(backoffs[i])
					continue
				}
				return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to start task after retries: %v", err)}
			}
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to start task: %v", err)}
		}
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"in_progress"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to start task after retries"}
}

func (h *ToolHandler) handleTaskComplete(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskCompleteInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	task, _ := h.taskRepo.GetByID(ctx, input.TaskID)

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "result_summary": input.ResultSummary}

	backoffs := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond}
	for i := 0; i < len(backoffs); i++ {
		if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusCompleted, updates); err != nil {
			if strings.Contains(err.Error(), "concurrent update") {
				if i < len(backoffs)-1 {
					time.Sleep(backoffs[i])
					continue
				}
				return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to complete task after retries: %v", err)}
			}
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to complete task: %v", err)}
		}
		if task != nil && task.SecretaryID != "" {
			h.notifySecretary(task, "completed", input.ResultSummary, sess)
		}
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"completed"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to complete task after retries"}
}

func (h *ToolHandler) handleTaskFail(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskFailInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	task, _ := h.taskRepo.GetByID(ctx, input.TaskID)

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "error_message": input.ErrorMessage}

	backoffs := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond}
	for i := 0; i < len(backoffs); i++ {
		if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusFailed, updates); err != nil {
			if strings.Contains(err.Error(), "concurrent update") {
				if i < len(backoffs)-1 {
					time.Sleep(backoffs[i])
					continue
				}
				return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to mark task as failed after retries: %v", err)}
			}
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to mark task as failed: %v", err)}
		}
		if task != nil && task.SecretaryID != "" {
			h.notifySecretary(task, "failed", input.ErrorMessage, sess)
		}
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"failed"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to mark task as failed after retries"}
}

func (h *ToolHandler) handleTaskList(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskListInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	var tasks []*models.Task
	var err error
	if input.SecretaryID != "" {
		tasks, err = h.taskRepo.ListBySecretary(ctx, input.SecretaryID)
	} else {
		tasks, err = h.taskRepo.ListByWorkspace(ctx, sess.WorkspaceID)
	}
	if err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to list tasks: %v", err)}
	}

	type taskSummary struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		Status        string `json:"status"`
		AssigneeID    string `json:"assigneeId,omitempty"`
		ResultSummary string `json:"resultSummary,omitempty"`
		ErrorMessage  string `json:"errorMessage,omitempty"`
		CreatedAt     int64  `json:"createdAt"`
	}
	summaries := make([]taskSummary, 0, len(tasks))
	for _, t := range tasks {
		summaries = append(summaries, taskSummary{
			ID: t.ID, Title: t.Title, Status: string(t.Status),
			AssigneeID: t.AssigneeID, ResultSummary: t.ResultSummary,
			ErrorMessage: t.ErrorMessage, CreatedAt: t.CreatedAt,
		})
	}
	data, _ := json.Marshal(summaries)
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: string(data)}
}

func (h *ToolHandler) handleWorkloadList(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input WorkloadListInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	stats, err := h.taskRepo.GetWorkloadStats(ctx, input.WorkspaceID)
	if err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to get workload stats: %v", err)}
	}

	data, _ := json.Marshal(stats)
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: string(data)}
}

func (h *ToolHandler) handleAgentStatus(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input AgentStatusInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	log.Printf("[a2a] Agent status update from %s: status=%s, message=%s", sess.MemberName, input.Status, input.Message)
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true}`}
}

func (h *ToolHandler) handleFileRead(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input FileReadInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	absPath := filepath.Join(sess.WorkspaceID, input.Path)
	if h.validator != nil {
		if err := h.validator.ValidatePath(absPath); err != nil {
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Path not allowed: %v", err)}
		}
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to read file: %v", err)}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: string(content)}
}

func (h *ToolHandler) handleFileWrite(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input FileWriteInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	absPath := filepath.Join(sess.WorkspaceID, input.Path)
	if h.validator != nil {
		if err := h.validator.ValidatePath(absPath); err != nil {
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Path not allowed: %v", err)}
		}
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to create directory: %v", err)}
	}

	if err := os.WriteFile(absPath, []byte(input.Content), 0644); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to write file: %v", err)}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true}`}
}

func (h *ToolHandler) handleFileList(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input FileListInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	absPath := filepath.Join(sess.WorkspaceID, input.Path)
	if h.validator != nil {
		if err := h.validator.ValidatePath(absPath); err != nil {
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Path not allowed: %v", err)}
		}
	}

	if h.browser != nil {
		entries, err := h.browser.ListDir(absPath, false)
		if err != nil {
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to list directory: %v", err)}
		}
		data, _ := json.Marshal(entries)
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: string(data)}
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to read directory: %v", err)}
	}

	var result []map[string]any
	for _, e := range entries {
		info, _ := e.Info()
		result = append(result, map[string]any{
			"name":    e.Name(),
			"isDir":   e.IsDir(),
			"size":    info.Size(),
			"modTime": info.ModTime().Unix(),
		})
	}

	data, _ := json.Marshal(result)
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: string(data)}
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Type      MessageType `json:"type"`
	ToolUseID string      `json:"tool_use_id"`
	Content   string      `json:"content"`
	IsError   bool        `json:"is_error"`
}

// ParseToolInput parses raw JSON input into a typed struct.
func ParseToolInput(input json.RawMessage, out interface{}) error {
	return json.Unmarshal(input, out)
}

// mustJSON marshals a value to JSON, panicking on error.
func mustJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
	}
	return data
}

// Tool name constants (same as ACP)
const (
	ToolChatSend     = "orchestra_chat_send"
	ToolTaskCreate   = "orchestra_task_create"
	ToolTaskStart    = "orchestra_task_start"
	ToolTaskComplete = "orchestra_task_complete"
	ToolTaskFail     = "orchestra_task_fail"
	ToolTaskList     = "orchestra_task_list"
	ToolWorkloadList = "orchestra_workload_list"
	ToolAgentStatus  = "orchestra_agent_status"
	ToolFileRead     = "orchestra_file_read"
	ToolFileWrite    = "orchestra_file_write"
	ToolFileList     = "orchestra_file_list"
)

// Input types for parsing
type ChatSendInput struct {
	ConversationID string `json:"conversationId"`
	Text           string `json:"text"`
}

type TaskCreateInput struct {
	ConversationID string `json:"conversationId"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	AssigneeID     string `json:"assigneeId"`
	Priority       int    `json:"priority"`
}

type TaskStartInput struct {
	TaskID  string `json:"taskId"`
	Message string `json:"message"`
}

type TaskCompleteInput struct {
	TaskID        string `json:"taskId"`
	ResultSummary string `json:"resultSummary"`
}

type TaskFailInput struct {
	TaskID       string `json:"taskId"`
	ErrorMessage string `json:"errorMessage"`
}

type TaskListInput struct {
	SecretaryID string `json:"secretaryId"`
}

type WorkloadListInput struct {
	WorkspaceID string `json:"workspaceId"`
}

type AgentStatusInput struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

type FileReadInput struct {
	Path string `json:"path"`
}

type FileWriteInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FileListInput struct {
	Path string `json:"path"`
}
