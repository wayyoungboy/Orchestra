package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/orchestra/backend/internal/models"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, task *models.Task) error {
	// Convert empty string to nil for nullable FK columns
	var assigneeID *string
	if task.AssigneeID != "" {
		assigneeID = &task.AssigneeID
	}

	query := `
		INSERT INTO tasks (
			id, workspace_id, conversation_id, secretary_id,
			title, description, status, assignee_id, priority,
			deadline_at, assigned_at, started_at, completed_at,
			result_summary, error_message, version, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.WorkspaceID, task.ConversationID, task.SecretaryID,
		task.Title, task.Description, task.Status, assigneeID, task.Priority,
		task.DeadlineAt, task.AssignedAt, task.StartedAt, task.CompletedAt,
		task.ResultSummary, task.ErrorMessage, task.Version, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, workspace_id, conversation_id, secretary_id,
			title, description, status, assignee_id, priority,
			deadline_at, assigned_at, started_at, completed_at,
			result_summary, error_message, version, created_at, updated_at
		FROM tasks WHERE id = ?
	`
	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.WorkspaceID, &task.ConversationID, &task.SecretaryID,
		&task.Title, &task.Description, &task.Status, &task.AssigneeID, &task.Priority,
		&task.DeadlineAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
		&task.ResultSummary, &task.ErrorMessage, &task.Version, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepo) ListByWorkspace(ctx context.Context, workspaceID string, status ...models.TaskStatus) ([]*models.Task, error) {
	query := `
		SELECT id, workspace_id, conversation_id, secretary_id,
			title, description, status, assignee_id, priority,
			deadline_at, assigned_at, started_at, completed_at,
			result_summary, error_message, version, created_at, updated_at
		FROM tasks WHERE workspace_id = ?
	`
	args := []interface{}{workspaceID}

	if len(status) > 0 {
		placeholders := make([]string, len(status))
		for i, s := range status {
			placeholders[i] = "?"
			args = append(args, string(s))
		}
		query += " AND status IN (" + strings.Join(placeholders, ",") + ")"
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.WorkspaceID, &task.ConversationID, &task.SecretaryID,
			&task.Title, &task.Description, &task.Status, &task.AssigneeID, &task.Priority,
			&task.DeadlineAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
			&task.ResultSummary, &task.ErrorMessage, &task.Version, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepo) ListByAssignee(ctx context.Context, assigneeID string) ([]*models.Task, error) {
	query := `
		SELECT id, workspace_id, conversation_id, secretary_id,
			title, description, status, assignee_id, priority,
			deadline_at, assigned_at, started_at, completed_at,
			result_summary, error_message, version, created_at, updated_at
		FROM tasks WHERE assignee_id = ?
		ORDER BY priority DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.WorkspaceID, &task.ConversationID, &task.SecretaryID,
			&task.Title, &task.Description, &task.Status, &task.AssigneeID, &task.Priority,
			&task.DeadlineAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
			&task.ResultSummary, &task.ErrorMessage, &task.Version, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepo) ListBySecretary(ctx context.Context, secretaryID string) ([]*models.Task, error) {
	query := `
		SELECT id, workspace_id, conversation_id, secretary_id,
			title, description, status, assignee_id, priority,
			deadline_at, assigned_at, started_at, completed_at,
			result_summary, error_message, version, created_at, updated_at
		FROM tasks WHERE secretary_id = ?
		ORDER BY priority DESC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, secretaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID, &task.WorkspaceID, &task.ConversationID, &task.SecretaryID,
			&task.Title, &task.Description, &task.Status, &task.AssigneeID, &task.Priority,
			&task.DeadlineAt, &task.AssignedAt, &task.StartedAt, &task.CompletedAt,
			&task.ResultSummary, &task.ErrorMessage, &task.Version, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id string, status models.TaskStatus, updates map[string]interface{}) error {
	// Get current task to check version and validate transition
	currentTask, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if currentTask == nil {
		return fmt.Errorf("task %s not found", id)
	}

	// Validate state transition
	if !models.IsValidTaskTransition(currentTask.Status, status) {
		return fmt.Errorf("invalid task transition from %s to %s", currentTask.Status, status)
	}

	// Build dynamic update query with optimistic locking
	setClauses := []string{"status = ?", "version = version + 1", "updated_at = ?"}
	args := []interface{}{string(status), updates["updated_at"]}

	// Add optional fields
	if v, ok := updates["assignee_id"]; ok {
		setClauses = append(setClauses, "assignee_id = ?")
		args = append(args, v)
	}
	if v, ok := updates["assigned_at"]; ok {
		setClauses = append(setClauses, "assigned_at = ?")
		args = append(args, v)
	}
	if v, ok := updates["started_at"]; ok {
		setClauses = append(setClauses, "started_at = ?")
		args = append(args, v)
	}
	if v, ok := updates["completed_at"]; ok {
		setClauses = append(setClauses, "completed_at = ?")
		args = append(args, v)
	}
	if v, ok := updates["result_summary"]; ok {
		setClauses = append(setClauses, "result_summary = ?")
		args = append(args, v)
	}
	if v, ok := updates["error_message"]; ok {
		setClauses = append(setClauses, "error_message = ?")
		args = append(args, v)
	}

	args = append(args, id, currentTask.Version)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ? AND version = ?", strings.Join(setClauses, ", "))
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("concurrent update: task %s was modified", id)
	}

	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	return err
}

// GetWorkloadStats returns workload statistics for a workspace
func (r *TaskRepo) GetWorkloadStats(ctx context.Context, workspaceID string) ([]map[string]interface{}, error) {
	query := `
		SELECT
			m.id as member_id,
			m.name as name,
			COALESCE(SUM(CASE WHEN t.status IN ('assigned', 'in_progress') THEN 1 ELSE 0 END), 0) as current_task_count,
			COALESCE(SUM(CASE WHEN t.status = 'pending' THEN 1 ELSE 0 END), 0) as pending_task_count,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END), 0) as completed_task_count
		FROM members m
		LEFT JOIN tasks t ON m.id = t.assignee_id AND t.workspace_id = ?
		WHERE m.workspace_id = ? AND m.role_type = 'assistant'
		GROUP BY m.id, m.name
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var memberID, name string
		var currentCount, pendingCount, completedCount int
		err := rows.Scan(&memberID, &name, &currentCount, &pendingCount, &completedCount)
		if err != nil {
			return nil, err
		}
		status := "idle"
		if currentCount > 0 {
			status = "working"
		}
		results = append(results, map[string]interface{}{
			"memberId":           memberID,
			"name":               name,
			"currentTaskCount":   currentCount,
			"pendingTaskCount":   pendingCount,
			"completedTaskCount": completedCount,
			"status":             status,
		})
	}
	return results, nil
}