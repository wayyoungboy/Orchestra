package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/internal/models"
)

// AttachmentRepository handles attachment database operations
type AttachmentRepository struct {
	db *sql.DB
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create inserts a new attachment record
func (r *AttachmentRepository) Create(ctx context.Context, a *models.Attachment) error {
	query := `INSERT INTO attachments (id, file_name, file_path, file_size, mime_type, message_id, workspace_id, uploaded_by, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.FileName, a.FilePath, a.FileSize, a.MimeType, a.MessageID, a.WorkspaceID, a.UploadedBy, a.CreatedAt.UnixMilli())
	return err
}

// GetByID retrieves an attachment by ID
func (r *AttachmentRepository) GetByID(ctx context.Context, id string) (*models.Attachment, error) {
	query := `SELECT id, file_name, file_path, file_size, mime_type, message_id, workspace_id, uploaded_by, created_at
	          FROM attachments WHERE id = ?`

	var a models.Attachment
	var messageID sql.NullString
	var createdAt int64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &messageID, &a.WorkspaceID, &a.UploadedBy, &createdAt)
	if err != nil {
		return nil, err
	}

	a.MessageID = messageID.String
	a.CreatedAt = time.UnixMilli(createdAt)
	return &a, nil
}

// ListByWorkspace lists all attachments in a workspace
func (r *AttachmentRepository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Attachment, error) {
	query := `SELECT id, file_name, file_path, file_size, mime_type, message_id, workspace_id, uploaded_by, created_at
	          FROM attachments WHERE workspace_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var a models.Attachment
		var messageID sql.NullString
		var createdAt int64

		if err := rows.Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &messageID, &a.WorkspaceID, &a.UploadedBy, &createdAt); err != nil {
			return nil, err
		}
		a.MessageID = messageID.String
		a.CreatedAt = time.UnixMilli(createdAt)
		attachments = append(attachments, &a)
	}
	return attachments, nil
}

// ListByConversation lists attachments associated with messages in a conversation
func (r *AttachmentRepository) ListByConversation(ctx context.Context, conversationID string) ([]*models.Attachment, error) {
	query := `SELECT a.id, a.file_name, a.file_path, a.file_size, a.mime_type, a.message_id, a.workspace_id, a.uploaded_by, a.created_at
	          FROM attachments a
	          JOIN messages m ON a.message_id = m.id
	          WHERE m.conversation_id = ?
	          ORDER BY a.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var a models.Attachment
		var messageID sql.NullString
		var createdAt int64

		if err := rows.Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &messageID, &a.WorkspaceID, &a.UploadedBy, &createdAt); err != nil {
			return nil, err
		}
		a.MessageID = messageID.String
		a.CreatedAt = time.UnixMilli(createdAt)
		attachments = append(attachments, &a)
	}
	return attachments, nil
}

// Delete removes an attachment record
func (r *AttachmentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = ?`, id)
	return err
}

// GetByMessageID retrieves attachments by message ID
func (r *AttachmentRepository) GetByMessageID(ctx context.Context, messageID string) ([]*models.Attachment, error) {
	query := `SELECT id, file_name, file_path, file_size, mime_type, message_id, workspace_id, uploaded_by, created_at
	          FROM attachments WHERE message_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var a models.Attachment
		var msgID sql.NullString
		var createdAt int64

		if err := rows.Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &msgID, &a.WorkspaceID, &a.UploadedBy, &createdAt); err != nil {
			return nil, err
		}
		a.MessageID = msgID.String
		a.CreatedAt = time.UnixMilli(createdAt)
		attachments = append(attachments, &a)
	}
	return attachments, nil
}