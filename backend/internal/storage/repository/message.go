package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	MessageStatusSent   MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead    MessageStatus = "read"
)

type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Message struct {
	ID             string        `json:"id"`
	ConversationID string        `json:"conversationId"`
	SenderID       string        `json:"senderId"`
	Content        MessageContent `json:"content"`
	IsAI           bool          `json:"isAi"`
	Attachment     string        `json:"attachment,omitempty"`
	Status         MessageStatus `json:"status"`
	CreatedAt      int64         `json:"createdAt"`
}

type MessageCreate struct {
	ConversationID string        `json:"conversationId"`
	SenderID       string        `json:"senderId"`
	Content        MessageContent `json:"content"`
	IsAI           bool          `json:"isAi"`
}

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) ListByConversation(conversationID string, limit int, beforeID string) ([]Message, error) {
	var query string
	var args []interface{}

	if beforeID != "" {
		query = `SELECT id, conversation_id, sender_id, content, is_ai, attachment, status, created_at
		         FROM messages WHERE conversation_id = ? AND id < ? ORDER BY created_at DESC LIMIT ?`
		args = []interface{}{conversationID, beforeID, limit}
	} else {
		query = `SELECT id, conversation_id, sender_id, content, is_ai, attachment, status, created_at
		         FROM messages WHERE conversation_id = ? ORDER BY created_at DESC LIMIT ?`
		args = []interface{}{conversationID, limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var contentJSON string
		var attachment sql.NullString

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &contentJSON,
			&msg.IsAI, &attachment, &msg.Status, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}

		msg.Attachment = attachment.String
		if err := json.Unmarshal([]byte(contentJSON), &msg.Content); err != nil {
			msg.Content = MessageContent{Type: "text", Text: contentJSON}
		}

		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) GetByID(id string) (*Message, error) {
	query := `SELECT id, conversation_id, sender_id, content, is_ai, attachment, status, created_at
	          FROM messages WHERE id = ?`

	var msg Message
	var contentJSON string
	var attachment sql.NullString

	err := r.db.QueryRow(query, id).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &contentJSON,
		&msg.IsAI, &attachment, &msg.Status, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	msg.Attachment = attachment.String
	if err := json.Unmarshal([]byte(contentJSON), &msg.Content); err != nil {
		msg.Content = MessageContent{Type: "text", Text: contentJSON}
	}

	return &msg, nil
}

func (r *MessageRepository) Create(data MessageCreate) (*Message, error) {
	now := time.Now().UnixMilli()
	id := "msg_" + uuid.New().String()[:8]

	contentJSON, err := json.Marshal(data.Content)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO messages (id, conversation_id, sender_id, content, is_ai, attachment, status, created_at)
	          VALUES (?, ?, ?, ?, ?, '', 'sent', ?)`

	_, err = r.db.Exec(query, id, data.ConversationID, data.SenderID, contentJSON, data.IsAI, now)
	if err != nil {
		return nil, err
	}

	// Update conversation's updated_at
	r.db.Exec(`UPDATE conversations SET updated_at = ? WHERE id = ?`, now, data.ConversationID)

	return r.GetByID(id)
}

func (r *MessageRepository) DeleteByConversation(conversationID string) error {
	_, err := r.db.Exec(`DELETE FROM messages WHERE conversation_id = ?`, conversationID)
	return err
}

func (r *MessageRepository) UpdateStatus(id string, status MessageStatus) error {
	_, err := r.db.Exec(`UPDATE messages SET status = ? WHERE id = ?`, status, id)
	return err
}

// CountUnreadForViewer returns messages from others strictly after lastReadAt (ms).
func (r *MessageRepository) CountUnreadForViewer(conversationID, viewerMemberID string, lastReadAt int64) (int, error) {
	const q = `SELECT COUNT(*) FROM messages WHERE conversation_id = ? AND created_at > ? AND sender_id != ?`
	var n int
	err := r.db.QueryRow(q, conversationID, lastReadAt, viewerMemberID).Scan(&n)
	return n, err
}

// LatestMessageTime returns the max created_at in the conversation, or 0 if none.
func (r *MessageRepository) LatestMessageTime(conversationID string) (int64, error) {
	var t sql.NullInt64
	err := r.db.QueryRow(
		`SELECT MAX(created_at) FROM messages WHERE conversation_id = ?`,
		conversationID,
	).Scan(&t)
	if err != nil {
		return 0, err
	}
	if !t.Valid {
		return 0, nil
	}
	return t.Int64, nil
}