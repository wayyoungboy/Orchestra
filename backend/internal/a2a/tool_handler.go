package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

// ChatBroadcaster is an interface for broadcasting chat messages.
type ChatBroadcaster interface {
	BroadcastToWorkspace(workspaceID string, event interface{})
}

// ToolHandler executes Orchestra tools on behalf of A2A sessions.
type ToolHandler struct {
	msgRepo    *repository.MessageRepository
	taskRepo   *repository.TaskRepo
	memberRepo repository.MemberRepository
	convRepo   *repository.ConversationRepository
	chatHub    ChatBroadcaster
	browser   *filesystem.Browser
	validator *filesystem.Validator
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
	return &ToolHandler{
		msgRepo:    msgRepo,
		taskRepo:   taskRepo,
		memberRepo: memberRepo,
		convRepo:   convRepo,
		chatHub:    chatHub,
		browser:    browser,
		validator:  validator,
	}
}

// ExecuteTool executes a tool call and sends the result back to the session.
func (h *ToolHandler) ExecuteTool(msg *ACPMessage, sess *Session) {
	toolUse, err := msg.ParseToolUseMessage()
	if err != nil {
		sess.OutputChan <- &ACPMessage{
			Type: TypeToolResult,
			Content: mustJSON(map[string]any{
				"type":      "tool_result",
				"tool_use_id": "",
				"is_error":  true,
				"content":   fmt.Sprintf("Failed to parse tool use: %v", err),
			}),
		}
		return
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

	sess.OutputChan <- &ACPMessage{
		Type:    TypeToolResult,
		Content: mustJSON(map[string]any{
			"type":      string(result.Type),
			"tool_use_id": result.ToolUseID,
			"is_error":  result.IsError,
			"content":   result.Content,
		}),
	}
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
		ID:             "task_" + uuid.New().String()[:8],
		WorkspaceID:    sess.WorkspaceID,
		ConversationID: input.ConversationID,
		SecretaryID:    sess.MemberID,
		Title:          input.Title,
		Description:    input.Description,
		Status:         models.TaskStatusPending,
		AssigneeID:     input.AssigneeID,
		Priority:       input.Priority,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if input.AssigneeID != "" {
		task.Status = models.TaskStatusAssigned
		task.AssignedAt = now
	}

	if err := h.taskRepo.Create(ctx, task); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to create task: %v", err)}
	}

	return &ToolResult{
		Type:      TypeToolResult,
		ToolUseID: toolUse.ToolUseID,
		Content:   fmt.Sprintf(`{"success":true,"taskId":"%s","status":"%s"}`, task.ID, task.Status),
	}
}

func (h *ToolHandler) handleTaskStart(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskStartInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "started_at": now}
	if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusInProgress, updates); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to start task: %v", err)}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"in_progress"}`}
}

func (h *ToolHandler) handleTaskComplete(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskCompleteInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "result_summary": input.ResultSummary}
	if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusCompleted, updates); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to complete task: %v", err)}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"completed"}`}
}

func (h *ToolHandler) handleTaskFail(ctx context.Context, toolUse *ToolUseMessage, sess *Session) *ToolResult {
	var input TaskFailInput
	if err := ParseToolInput(toolUse.Input, &input); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Invalid input: %v", err)}
	}

	now := time.Now().UnixMilli()
	updates := map[string]interface{}{"updated_at": now, "completed_at": now, "error_message": input.ErrorMessage}
	if err := h.taskRepo.UpdateStatus(ctx, input.TaskID, models.TaskStatusFailed, updates); err != nil {
		return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, IsError: true, Content: fmt.Sprintf("Failed to mark task as failed: %v", err)}
	}

	return &ToolResult{Type: TypeToolResult, ToolUseID: toolUse.ToolUseID, Content: `{"success":true,"status":"failed"}`}
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
