package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

type TaskHandler struct {
	taskRepo *repository.TaskRepo
	memberRepo repository.MemberRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepo, memberRepo repository.MemberRepository) *TaskHandler {
	return &TaskHandler{
		taskRepo:   taskRepo,
		memberRepo: memberRepo,
	}
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

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"taskId":        req.TaskID,
		"status":        models.TaskStatusInProgress,
		"conversationId": task.ConversationID,
		"secretaryId":   task.SecretaryID,
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

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"taskId":       req.TaskID,
		"status":       models.TaskStatusCompleted,
		"conversationId": task.ConversationID,
		"secretaryId":  task.SecretaryID,
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

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"taskId":        req.TaskID,
		"status":        models.TaskStatusFailed,
		"conversationId": task.ConversationID,
		"secretaryId":   task.SecretaryID,
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

	task, err := h.taskRepo.GetByID(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"taskId":       taskID,
		"status":       models.TaskStatusCancelled,
		"conversationId": task.ConversationID,
		"secretaryId":  task.SecretaryID,
	})
}

// ListTasks returns tasks for a workspace
func (h *TaskHandler) ListTasks(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
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

	task, err := h.taskRepo.GetByID(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"task": task,
	})
}

// GetMyTasks returns tasks assigned to a specific member
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	memberID := c.Query("memberId")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "memberId required"})
		return
	}

	tasks, err := h.taskRepo.ListByAssignee(c.Request.Context(), memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"tasks": tasks,
	})
}

// ListWorkloads returns workload statistics for all assistants in a workspace
func (h *TaskHandler) ListWorkloads(c *gin.Context) {
	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspaceId required"})
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