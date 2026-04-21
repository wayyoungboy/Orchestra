package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/orchestra/backend/pkg/utils"
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
	id := "msg_" + utils.GenerateID()[:10]

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

// SearchResult represents a search result item
type SearchResult struct {
	Message       Message `json:"message"`
	ConversationID string `json:"conversationId"`
	WorkspaceID   string `json:"workspaceId"`
	Snippet       string `json:"snippet"`
}

// SearchInWorkspace searches messages across all conversations in a workspace
func (r *MessageRepository) SearchInWorkspace(workspaceID, query string, limit int) ([]SearchResult, error) {
	// Search in message content using LIKE (simple approach)
	// For better performance, consider using SQLite FTS5 or external search engine
	searchQuery := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.is_ai, m.attachment, m.status, m.created_at,
		       c.workspace_id
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE c.workspace_id = ? AND m.content LIKE ?
		ORDER BY m.created_at DESC
		LIMIT ?
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(searchQuery, workspaceID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var msg Message
		var contentJSON string
		var attachment sql.NullString
		var wsID string

		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &contentJSON,
			&msg.IsAI, &attachment, &msg.Status, &msg.CreatedAt, &wsID)
		if err != nil {
			return nil, err
		}

		msg.Attachment = attachment.String
		if err := json.Unmarshal([]byte(contentJSON), &msg.Content); err != nil {
			msg.Content = MessageContent{Type: "text", Text: contentJSON}
		}

		// Generate snippet (extract surrounding text)
		snippet := generateSnippet(msg.Content.Text, query, 100)

		results = append(results, SearchResult{
			Message:        msg,
			ConversationID: msg.ConversationID,
			WorkspaceID:    wsID,
			Snippet:        snippet,
		})
	}

	return results, nil
}

// generateSnippet extracts a snippet around the matching text
func generateSnippet(text, query string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	// Find the position of the query in text (case-insensitive)
	lowerText := text
	lowerQuery := query
	idx := -1
	for i := 0; i <= len(lowerText)-len(lowerQuery); i++ {
		match := true
		for j := 0; j < len(lowerQuery); j++ {
			if lowerText[i+j] != lowerQuery[j] && !(lowerText[i+j] >= 'A' && lowerText[i+j] <= 'Z' && lowerText[i+j]+32 == lowerQuery[j]) {
				match = false
				break
			}
		}
		if match {
			idx = i
			break
		}
	}

	if idx == -1 {
		// Query not found, return first maxLen chars
		if len(text) > maxLen {
			return text[:maxLen] + "..."
		}
		return text
	}

	// Calculate start position to center the match
	start := idx - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(text) {
		end = len(text)
	}

	snippet := text[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}
// Delete deletes a single message by ID
func (r *MessageRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM messages WHERE id = ?`, id)
	return err
}
