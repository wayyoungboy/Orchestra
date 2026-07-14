package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
)

type TaskHandler struct {
	taskRepo   *repository.TaskRepo
	memberRepo repository.MemberRepository
	wsRepo     repository.WorkspaceRepository
	convRepo   *repository.ConversationRepository
	chatHub    *ws.ChatHub
}

func NewTaskHandler(taskRepo *repository.TaskRepo, memberRepo repository.MemberRepository, wsRepo repository.WorkspaceRepository, convRepo *repository.ConversationRepository, chatHub *ws.ChatHub) *TaskHandler {
	return &TaskHandler{
		taskRepo:   taskRepo,
		memberRepo: memberRepo,
		wsRepo:     wsRepo,
		convRepo:   convRepo,
		chatHub:    chatHub,
	}
}

func (h *TaskHandler) broadcastTaskStatus(task *models.Task, newStatus models.TaskStatus) {
	if h.chatHub == nil {
		return
	}
	payload, err := json.Marshal(map[string]string{
		"type":        "task_status",
		"workspaceId": task.WorkspaceID,
		"taskId":      task.ID,
		"status":      string(newStatus),
		"assigneeId":  task.AssigneeID,
		"title":       task.Title,
	})
	if err == nil {
		h.chatHub.BroadcastRawToWorkspace(task.WorkspaceID, payload)
	}
}

func (h *TaskHandler) requireWorkspace(c *gin.Context, workspaceID string) bool {
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return false
	}
	if _, err := h.wsRepo.GetByID(c.Request.Context(), workspaceID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return false
	}
	return true
}

func (h *TaskHandler) memberInWorkspace(c *gin.Context, workspaceID, memberID string, role models.MemberRole) bool {
	member, err := h.memberRepo.GetByID(c.Request.Context(), memberID)
	if err != nil || member == nil || member.WorkspaceID != workspaceID || (role != "" && member.RoleType != role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member does not have the required workspace role"})
		return false
	}
	return true
}

func (h *TaskHandler) taskForWorkspace(c *gin.Context, taskID string) (*models.Task, bool) {
	workspaceID := c.Param("id")
	if !h.requireWorkspace(c, workspaceID) {
		return nil, false
	}
	task, err := h.taskRepo.GetByID(c.Request.Context(), taskID)
	if err != nil || task == nil || task.WorkspaceID != workspaceID {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return nil, false
	}
	return task, true
}

// TaskCreateRequest is the request body for creating a task
type TaskCreateRequest struct {
	WorkspaceID    string `json:"workspaceId" binding:"required"`
	ConversationID string `json:"conversationId" binding:"required"`
	SecretaryID    string `json:"secretaryId" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description,omitempty"`
	AssigneeID     string `json:"assigneeId,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	DeadlineAt     int64  `json:"deadlineAt,omitempty"`
}

// CreateTask creates a new task (internal API for AI to call)
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	req.SecretaryID = strings.TrimSpace(req.SecretaryID)
	req.AssigneeID = strings.TrimSpace(req.AssigneeID)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if !h.requireWorkspace(c, req.WorkspaceID) {
		return
	}
	conv, err := h.convRepo.GetByID(req.ConversationID)
	if err != nil || conv == nil || conv.WorkspaceID != req.WorkspaceID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation does not belong to workspace"})
		return
	}
	if !h.memberInWorkspace(c, req.WorkspaceID, req.SecretaryID, models.RoleSecretary) {
		return
	}
	if req.AssigneeID != "" && !h.memberInWorkspace(c, req.WorkspaceID, req.AssigneeID, models.RoleAssistant) {
		return
	}

	task := models.NewTask(models.TaskCreate{
		WorkspaceID:    req.WorkspaceID,
		ConversationID: req.ConversationID,
		SecretaryID:    req.SecretaryID,
		Title:          req.Title,
		Description:    req.Description,
		AssigneeID:     req.AssigneeID,
		Priority:       req.Priority,
		DeadlineAt:     req.DeadlineAt,
	})

	if err := h.taskRepo.Create(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.broadcastTaskStatus(task, task.Status)

	c.JSON(http.StatusCreated, gin.H{
		"ok":     true,
		"taskId": task.ID,
		"task":   task,
	})
}

// TaskStartRequest is the request body for starting a task
type TaskStartRequest struct {
	TaskID  string `json:"taskId" binding:"required"`
	Message string `json:"message,omitempty"`
}

// StartTask marks a task as in progress
func (h *TaskHandler) StartTask(c *gin.Context) {
	var req TaskStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
		"started_at": time.Now().UnixMilli(),
	}

	if err := h.taskRepo.UpdateStatus(c.Request.Context(), req.TaskID, models.TaskStatusInProgress, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.broadcastTaskStatus(task, models.TaskStatusInProgress)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"taskId":         req.TaskID,
		"status":         models.TaskStatusInProgress,
		"conversationId": task.ConversationID,
		"secretaryId":    task.SecretaryID,
	})
}

// TaskCompleteRequest is the request body for completing a task
type TaskCompleteRequest struct {
	TaskID        string `json:"taskId" binding:"required"`
	ResultSummary string `json:"resultSummary,omitempty"`
	Message       string `json:"message,omitempty"`
}

// CompleteTask marks a task as completed
func (h *TaskHandler) CompleteTask(c *gin.Context) {
	var req TaskCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	updates := map[string]interface{}{
		"updated_at":     time.Now().UnixMilli(),
		"completed_at":   time.Now().UnixMilli(),
		"result_summary": req.ResultSummary,
	}

	if err := h.taskRepo.UpdateStatus(c.Request.Context(), req.TaskID, models.TaskStatusCompleted, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.broadcastTaskStatus(task, models.TaskStatusCompleted)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"taskId":         req.TaskID,
		"status":         models.TaskStatusCompleted,
		"conversationId": task.ConversationID,
		"secretaryId":    task.SecretaryID,
	})
}

// TaskFailRequest is the request body for reporting a failed task
type TaskFailRequest struct {
	TaskID       string `json:"taskId" binding:"required"`
	ErrorMessage string `json:"errorMessage" binding:"required"`
}

// FailTask marks a task as failed
func (h *TaskHandler) FailTask(c *gin.Context) {
	var req TaskFailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	updates := map[string]interface{}{
		"updated_at":    time.Now().UnixMilli(),
		"completed_at":  time.Now().UnixMilli(),
		"error_message": req.ErrorMessage,
	}

	if err := h.taskRepo.UpdateStatus(c.Request.Context(), req.TaskID, models.TaskStatusFailed, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.broadcastTaskStatus(task, models.TaskStatusFailed)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"taskId":         req.TaskID,
		"status":         models.TaskStatusFailed,
		"conversationId": task.ConversationID,
		"secretaryId":    task.SecretaryID,
	})
}

// TaskAssignRequest is the request body for assigning a task
type TaskAssignRequest struct {
	TaskID     string `json:"taskId" binding:"required"`
	AssigneeID string `json:"assigneeId,omitempty"`
}

// AssignTask marks a task as assigned
func (h *TaskHandler) AssignTask(c *gin.Context) {
	var req TaskAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), req.TaskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if !models.IsValidTaskTransition(task.Status, models.TaskStatusAssigned) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot assign task in current status"})
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}
	req.AssigneeID = strings.TrimSpace(req.AssigneeID)
	if req.AssigneeID != "" {
		if !h.memberInWorkspace(c, task.WorkspaceID, req.AssigneeID, models.RoleAssistant) {
			return
		}
		updates["assignee_id"] = req.AssigneeID
		updates["assigned_at"] = time.Now().UnixMilli()
	}

	if err := h.taskRepo.UpdateStatus(c.Request.Context(), req.TaskID, models.TaskStatusAssigned, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.broadcastTaskStatus(task, models.TaskStatusAssigned)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"taskId":         req.TaskID,
		"status":         models.TaskStatusAssigned,
		"conversationId": task.ConversationID,
		"secretaryId":    task.SecretaryID,
	})
}

// CancelTaskRequest is the request body for cancelling a task
type CancelTaskRequest struct {
	TaskID  string `json:"taskId" binding:"required"`
	Message string `json:"message,omitempty"`
}

// CancelTask marks a task as cancelled
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		var req struct {
			TaskID  string `json:"taskId" binding:"required"`
			Message string `json:"message,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "taskId required"})
			return
		}
		taskID = req.TaskID
	}

	var task *models.Task
	var err error
	if c.Param("id") != "" {
		task, _ = h.taskForWorkspace(c, taskID)
		if task == nil {
			return
		}
	} else {
		task, err = h.taskRepo.GetByID(c.Request.Context(), taskID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
	}

	if !models.IsValidTaskTransition(task.Status, models.TaskStatusCancelled) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot cancel task in current status"})
		return
	}

	updates := map[string]interface{}{
		"updated_at":    time.Now().UnixMilli(),
		"completed_at":  time.Now().UnixMilli(),
		"error_message": "",
	}

	if err := h.taskRepo.UpdateStatus(c.Request.Context(), taskID, models.TaskStatusCancelled, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.broadcastTaskStatus(task, models.TaskStatusCancelled)

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"taskId":         taskID,
		"status":         models.TaskStatusCancelled,
		"conversationId": task.ConversationID,
		"secretaryId":    task.SecretaryID,
	})
}

// ListTasks returns tasks for a workspace
func (h *TaskHandler) ListTasks(c *gin.Context) {
	workspaceID := c.Param("id")
	if !h.requireWorkspace(c, workspaceID) {
		return
	}

	statusFilter := c.QueryArray("status")
	var statuses []models.TaskStatus
	for _, s := range statusFilter {
		statuses = append(statuses, models.TaskStatus(s))
	}

	tasks, err := h.taskRepo.ListByWorkspace(c.Request.Context(), workspaceID, statuses...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"tasks": tasks,
	})
}

// GetTask returns a single task by ID
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task id required"})
		return
	}

	task, ok := h.taskForWorkspace(c, taskID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"task": task,
	})
}

// GetMyTasks returns tasks assigned to a specific member
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	workspaceID := c.Param("id")
	if !h.requireWorkspace(c, workspaceID) {
		return
	}
	memberID := c.Query("memberId")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId required"})
		return
	}
	if !h.memberInWorkspace(c, workspaceID, memberID, "") {
		return
	}

	tasks, err := h.taskRepo.ListByAssignee(c.Request.Context(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workspaceTasks := make([]*models.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.WorkspaceID == workspaceID {
			workspaceTasks = append(workspaceTasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"tasks": workspaceTasks,
	})
}

// ListWorkloads returns workload statistics for all assistants in a workspace
func (h *TaskHandler) ListWorkloads(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspaceId required"})
		return
	}
	if !h.requireWorkspace(c, workspaceID) {
		return
	}

	workloads, err := h.taskRepo.GetWorkloadStats(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"workloads": workloads,
	})
}
