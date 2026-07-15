package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// SessionLookup finds an active session for a workspace member.
// Defined here to avoid import cycles between a2a and handler packages.
type SessionLookup interface {
	SessionForWorkspaceMember(workspaceID, memberID string) *Session
	Acquire(ctx context.Context, config SessionConfig) (*Session, error)
}

// ToolHandler executes Orchestra tools on behalf of A2A sessions.
type ToolHandler struct {
	msgRepo       *repository.MessageRepository
	taskRepo      *repository.TaskRepo
	memberRepo    repository.MemberRepository
	convRepo      *repository.ConversationRepository
	chatHub       ChatBroadcaster
	browser       *filesystem.Browser
	validator     *filesystem.Validator
	pool          SessionLookup
	workspaceRepo repository.WorkspaceRepository

	dispatchWg     sync.WaitGroup
	dispatchCtx    context.Context
	dispatchCancel context.CancelFunc
}

const maxAgentFileContentSize = 1 << 20 // 1 MiB

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
	if _, err := h.sessionMember(ctx, sess); err != nil {
		return toolError(toolUse, err)
	}
	if err := h.conversationInSessionWorkspace(input.ConversationID, sess); err != nil {
		return toolError(toolUse, err)
	}
	if strings.TrimSpace(input.Text) == "" {
		return toolError(toolUse, fmt.Errorf("message text is required"))
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
	member, err := h.sessionMember(ctx, sess)
	if err != nil {
		return toolError(toolUse, err)
	}
	if member.RoleType != models.RoleSecretary {
		return toolError(toolUse, fmt.Errorf("only a secretary can create tasks"))
	}
	if err := h.conversationInSessionWorkspace(input.ConversationID, sess); err != nil {
		return toolError(toolUse, err)
	}
	if strings.TrimSpace(input.Title) == "" {
		return toolError(toolUse, fmt.Errorf("task title is required"))
	}
	if input.AssigneeID != "" {
		assignee, err := h.memberRepo.GetByID(ctx, input.AssigneeID)
		if err != nil || assignee == nil || assignee.WorkspaceID != sess.WorkspaceID || assignee.RoleType != models.RoleAssistant {
			return toolError(toolUse, fmt.Errorf("assignee must be an assistant in this workspace"))
		}
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
	h.broadcastTaskStatus(task, task.Status)

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
		if member.WorkspaceID != task.WorkspaceID || member.RoleType != models.RoleAssistant {
			log.Printf("[a2a] Task dispatch: assignee %s is not an assistant in workspace %s", member.ID, task.WorkspaceID)
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
	task, err := h.taskAssignedToSession(ctx, input.TaskID, sess)
	if err != nil {
		return toolError(toolUse, err)
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
		h.broadcastTaskStatus(task, models.TaskStatusInProgress)
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"in_progress"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to start task after retries"}
}

func (h *ToolHandler) handleTaskComplete(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskCompleteInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}
	task, err := h.taskAssignedToSession(ctx, input.TaskID, sess)
	if err != nil {
		return toolError(toolUse, err)
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "result_summary": input.ResultSummary}

	// Retry logic for concurrent updates (max 3 attempts)
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
		h.broadcastTaskStatus(task, models.TaskStatusCompleted)
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"completed"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to complete task after retries"}
}

func (h *ToolHandler) handleTaskFail(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskFailInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}
	task, err := h.taskAssignedToSession(ctx, input.TaskID, sess)
	if err != nil {
		return toolError(toolUse, err)
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "error_message": input.ErrorMessage}

	// Retry logic for concurrent updates (max 3 attempts)
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
		h.broadcastTaskStatus(task, models.TaskStatusFailed)
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"failed"}`}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: "Failed to mark task as failed after retries"}
}

func (h *ToolHandler) handleWorkloadList(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input WorkloadListInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	stats, err := h.taskRepo.GetWorkloadStats(ctx, sess.WorkspaceID)
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
	if h.chatHub != nil {
		h.chatHub.BroadcastToWorkspace(sess.WorkspaceID, map[string]interface{}{
			"type":        "message_status",
			"workspaceId": sess.WorkspaceID,
			"senderId":    sess.MemberID,
			"status":      input.Status,
			"content":     input.Message,
		})
	}
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true}`}
}

func (h *ToolHandler) handleFileRead(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input FileReadInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	absPath, err := h.sessionWorkspacePath(ctx, sess, input.Path, false)
	if err != nil {
		return toolError(toolUse, err)
	}

	content, err := readAgentFile(absPath)
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

	absPath, err := h.sessionWorkspacePath(ctx, sess, input.Path, false)
	if err != nil {
		return toolError(toolUse, err)
	}
	if len(input.Content) > maxAgentFileContentSize {
		return toolError(toolUse, fmt.Errorf("file content exceeds the %d byte limit", maxAgentFileContentSize))
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

	absPath, err := h.sessionWorkspacePath(ctx, sess, input.Path, true)
	if err != nil {
		return toolError(toolUse, err)
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
		info, err := e.Info()
		if err != nil {
			return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to inspect directory entry: %v", err)}
		}
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

func toolError(toolUse *ToolUseMessage, err error) *ToolResult {
	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: err.Error()}
}

func (h *ToolHandler) sessionMember(ctx context.Context, sess *Session) (*models.Member, error) {
	if sess == nil || sess.WorkspaceID == "" || sess.MemberID == "" {
		return nil, fmt.Errorf("session is missing workspace or member identity")
	}
	if h.memberRepo == nil {
		return nil, fmt.Errorf("member repository is unavailable")
	}
	member, err := h.memberRepo.GetByID(ctx, sess.MemberID)
	if err != nil || member == nil || member.WorkspaceID != sess.WorkspaceID {
		return nil, fmt.Errorf("session member does not belong to this workspace")
	}
	return member, nil
}

func (h *ToolHandler) conversationInSessionWorkspace(conversationID string, sess *Session) error {
	if sess == nil || strings.TrimSpace(conversationID) == "" {
		return fmt.Errorf("conversation id is required")
	}
	conv, err := h.convRepo.GetByID(conversationID)
	if err != nil || conv == nil || conv.WorkspaceID != sess.WorkspaceID {
		return fmt.Errorf("conversation does not belong to this workspace")
	}
	return nil
}

func (h *ToolHandler) taskAssignedToSession(ctx context.Context, taskID string, sess *Session) (*models.Task, error) {
	if sess == nil || strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("task id is required")
	}
	task, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil || task.WorkspaceID != sess.WorkspaceID {
		return nil, fmt.Errorf("task does not belong to this workspace")
	}
	if task.AssigneeID != sess.MemberID {
		return nil, fmt.Errorf("task is not assigned to this agent")
	}
	return task, nil
}

func (h *ToolHandler) sessionWorkspacePath(ctx context.Context, sess *Session, requestedPath string, allowRoot bool) (string, error) {
	if sess == nil || sess.WorkspaceID == "" {
		return "", fmt.Errorf("session is missing workspace identity")
	}
	if h.workspaceRepo == nil {
		return "", fmt.Errorf("workspace repository is unavailable")
	}
	workspace, err := h.workspaceRepo.GetByID(ctx, sess.WorkspaceID)
	if err != nil || workspace == nil || strings.TrimSpace(workspace.Path) == "" {
		return "", fmt.Errorf("workspace path is unavailable")
	}

	root, err := filepath.Abs(workspace.Path)
	if err != nil {
		return "", fmt.Errorf("resolve workspace path: %w", err)
	}
	root = filepath.Clean(root)
	if requestedPath == "" {
		if !allowRoot {
			return "", fmt.Errorf("path is required")
		}
		requestedPath = "."
	}
	if filepath.IsAbs(requestedPath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	path := filepath.Join(root, requestedPath)
	workspaceValidator := filesystem.NewValidator([]string{root})
	if err := workspaceValidator.ValidatePath(path); err != nil {
		return "", fmt.Errorf("path is outside the workspace: %w", err)
	}
	if h.validator != nil {
		if err := h.validator.ValidatePath(path); err != nil {
			return "", fmt.Errorf("path is not allowed: %w", err)
		}
	}
	return path, nil
}

func readAgentFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}
	if info.Size() > maxAgentFileContentSize {
		return nil, fmt.Errorf("file exceeds the %d byte limit", maxAgentFileContentSize)
	}
	content, err := io.ReadAll(io.LimitReader(file, maxAgentFileContentSize+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxAgentFileContentSize {
		return nil, fmt.Errorf("file exceeds the %d byte limit", maxAgentFileContentSize)
	}
	return content, nil
}

func (h *ToolHandler) broadcastTaskStatus(task *models.Task, status models.TaskStatus) {
	if h.chatHub == nil {
		return
	}
	h.chatHub.BroadcastToWorkspace(task.WorkspaceID, map[string]interface{}{
		"type":        "task_status",
		"workspaceId": task.WorkspaceID,
		"taskId":      task.ID,
		"status":      string(status),
		"assigneeId":  task.AssigneeID,
		"title":       task.Title,
	})
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
