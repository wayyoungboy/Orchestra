package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/pkg/utils"
)

// Notification represents a notification for a user.
type Notification struct {
	ID             string `json:"id"`
	WorkspaceID    string `json:"workspaceId"`
	UserID         string `json:"userId"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	ConversationID string `json:"conversationId,omitempty"`
	MessageID      string `json:"messageId,omitempty"`
	IsRead         bool   `json:"isRead"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// NotificationRepository manages notifications in the database.
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a new notification.
func (r *NotificationRepository) Create(ctx context.Context, n *Notification) error {
	if n.ID == "" {
		n.ID = utils.GenerateID()
	}
	now := time.Now().UnixMilli()
	if n.CreatedAt == 0 {
		n.CreatedAt = now
	}
	if n.UpdatedAt == 0 {
		n.UpdatedAt = now
	}
	if n.Type == "" {
		n.Type = "message"
	}

	isRead := 0
	if n.IsRead {
		isRead = 1
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (id, workspace_id, user_id, type, title, body, conversation_id, message_id, is_read, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, n.ID, n.WorkspaceID, n.UserID, n.Type, n.Title, n.Body, sql.NullString{String: n.ConversationID, Valid: n.ConversationID != ""}, sql.NullString{String: n.MessageID, Valid: n.MessageID != ""}, isRead, n.CreatedAt, n.UpdatedAt)

	return err
}

// ListByUser lists notifications for a user, ordered by created_at desc.
func (r *NotificationRepository) ListByUser(ctx context.Context, workspaceID, userID string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, user_id, type, title, body, conversation_id, message_id, is_read, created_at, updated_at
		FROM notifications
		WHERE workspace_id = ? AND user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, workspaceID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var convID, msgID sql.NullString
		var isRead int
		if err := rows.Scan(&n.ID, &n.WorkspaceID, &n.UserID, &n.Type, &n.Title, &n.Body, &convID, &msgID, &isRead, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		n.ConversationID = convID.String
		n.MessageID = msgID.String
		n.IsRead = isRead == 1
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []Notification{}
	}

	return notifications, nil
}

// MarkRead marks a single notification as read.
func (r *NotificationRepository) MarkRead(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = 1, updated_at = ? WHERE id = ?
	`, now, id)
	return err
}

// MarkAllRead marks all notifications as read for a user in a workspace.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, workspaceID, userID string) error {
	now := time.Now().UnixMilli()
	_, err := r.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = 1, updated_at = ? WHERE workspace_id = ? AND user_id = ? AND is_read = 0
	`, now, workspaceID, userID)
	return err
}

// BadgeCounts returns total and unread notification counts for a user.
func (r *NotificationRepository) BadgeCounts(ctx context.Context, workspaceID, userID string) (total int, unread int, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), SUM(CASE WHEN is_read = 0 THEN 1 ELSE 0 END)
		FROM notifications
		WHERE workspace_id = ? AND user_id = ?
	`, workspaceID, userID).Scan(&total, &unread)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return total, unread, err
}

// Delete removes a notification by ID.
func (r *NotificationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notifications WHERE id = ?`, id)
	return err
}
