package models

import "time"

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task represents a task assigned by a secretary to an assistant
type Task struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspaceId"`
	ConversationID string     `json:"conversationId"`
	SecretaryID    string     `json:"secretaryId"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	Status         TaskStatus `json:"status"`
	AssigneeID     string     `json:"assigneeId,omitempty"`
	Priority       int        `json:"priority"`
	DeadlineAt     int64      `json:"deadlineAt,omitempty"`
	AssignedAt     int64      `json:"assignedAt,omitempty"`
	StartedAt      int64      `json:"startedAt,omitempty"`
	CompletedAt    int64      `json:"completedAt,omitempty"`
	ResultSummary  string     `json:"resultSummary,omitempty"`
	ErrorMessage   string     `json:"errorMessage,omitempty"`
	Version        int        `json:"version"`
	CreatedAt      int64      `json:"createdAt"`
	UpdatedAt      int64      `json:"updatedAt"`
}

// TaskCreate is the input for creating a new task
type TaskCreate struct {
	WorkspaceID    string `json:"workspaceId" binding:"required"`
	ConversationID string `json:"conversationId" binding:"required"`
	SecretaryID    string `json:"secretaryId" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description,omitempty"`
	AssigneeID     string `json:"assigneeId,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	DeadlineAt     int64  `json:"deadlineAt,omitempty"`
}

// TaskStatusUpdate is the input for updating task status
type TaskStatusUpdate struct {
	TaskID        string     `json:"taskId" binding:"required"`
	Status        TaskStatus `json:"status" binding:"required"`
	Message       string     `json:"message,omitempty"`
	ResultSummary string     `json:"resultSummary,omitempty"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
}

// TaskStart is the input for starting a task
type TaskStart struct {
	TaskID  string `json:"taskId" binding:"required"`
	Message string `json:"message,omitempty"`
}

// TaskComplete is the input for completing a task
type TaskComplete struct {
	TaskID        string `json:"taskId" binding:"required"`
	ResultSummary string `json:"resultSummary,omitempty"`
	Message       string `json:"message,omitempty"`
}

// TaskFail is the input for reporting a failed task
type TaskFail struct {
	TaskID       string `json:"taskId" binding:"required"`
	ErrorMessage string `json:"errorMessage" binding:"required"`
}

// AgentWorkload represents an agent's current workload
type AgentWorkload struct {
	MemberID           string `json:"memberId"`
	Name               string `json:"name"`
	CurrentTaskCount   int    `json:"currentTaskCount"`
	PendingTaskCount   int    `json:"pendingTaskCount"`
	CompletedTaskCount int    `json:"completedTaskCount"`
	Status             string `json:"status"` // idle, working, offline
}

// IsValidTaskTransition checks if a task status transition is allowed
func IsValidTaskTransition(from, to TaskStatus) bool {
	transitions := map[TaskStatus][]TaskStatus{
		TaskStatusPending:    {TaskStatusAssigned, TaskStatusCancelled},
		TaskStatusAssigned:   {TaskStatusInProgress, TaskStatusCancelled},
		TaskStatusInProgress: {TaskStatusCompleted, TaskStatusFailed},
	}
	allowed, ok := transitions[from]
	if !ok {
		return false // terminal state
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// NewTask creates a new Task instance
func NewTask(create TaskCreate) *Task {
	now := time.Now().UnixMilli()
	task := &Task{
		ID:             generateTaskID(),
		WorkspaceID:    create.WorkspaceID,
		ConversationID: create.ConversationID,
		SecretaryID:    create.SecretaryID,
		Title:          create.Title,
		Description:    create.Description,
		Status:         TaskStatusPending,
		Priority:       create.Priority,
		DeadlineAt:     create.DeadlineAt,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if create.AssigneeID != "" {
		task.AssigneeID = create.AssigneeID
		task.Status = TaskStatusAssigned
		task.AssignedAt = now
	}
	return task
}

func generateTaskID() string {
	return "task_" + generateULID()
}

func generateULID() string {
	// Simple ULID-like ID generation
	now := time.Now().UnixMilli()
	return encodeTime(now) + randomString(10)
}

func encodeTime(ms int64) string {
	const charset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	var result [10]byte
	for i := 9; i >= 0; i-- {
		result[i] = charset[ms%32]
		ms /= 32
	}
	return string(result[:])
}

func randomString(n int) string {
	const charset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}