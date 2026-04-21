package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/orchestra/backend/pkg/utils"
)

type ConversationType string

const (
	ConversationTypeChannel ConversationType = "channel"
	ConversationTypeDM      ConversationType = "dm"
)

type Conversation struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspaceId"`
	Type        ConversationType `json:"type"`
	Name        string          `json:"name,omitempty"`
	TargetID    string          `json:"targetId,omitempty"`
	MemberIDs   []string        `json:"memberIds"`
	Pinned      bool            `json:"pinned"`
	Muted       bool            `json:"muted"`
	CreatedAt   int64           `json:"createdAt"`
	UpdatedAt   int64           `json:"updatedAt"`
}

type ConversationCreate struct {
	Type       ConversationType `json:"type"`
	MemberIDs  []string         `json:"memberIds"`
	Name       string           `json:"name,omitempty"`
	TargetID   string           `json:"targetId,omitempty"`
}

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) ListByWorkspace(workspaceID string) ([]Conversation, error) {
	query := `SELECT id, workspace_id, type, name, target_id, member_ids, pinned, muted, created_at, updated_at
	          FROM conversations WHERE workspace_id = ? ORDER BY pinned DESC, updated_at DESC`

	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)
	for rows.Next() {
		var conv Conversation
		var memberIDsJSON string
		var name, targetID sql.NullString

		err := rows.Scan(&conv.ID, &conv.WorkspaceID, &conv.Type, &name, &targetID, &memberIDsJSON,
			&conv.Pinned, &conv.Muted, &conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, err
		}

		conv.Name = name.String
		conv.TargetID = targetID.String
		if err := json.Unmarshal([]byte(memberIDsJSON), &conv.MemberIDs); err != nil {
			conv.MemberIDs = []string{}
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (r *ConversationRepository) GetByID(id string) (*Conversation, error) {
	query := `SELECT id, workspace_id, type, name, target_id, member_ids, pinned, muted, created_at, updated_at
	          FROM conversations WHERE id = ?`

	var conv Conversation
	var memberIDsJSON string
	var name, targetID sql.NullString

	err := r.db.QueryRow(query, id).Scan(&conv.ID, &conv.WorkspaceID, &conv.Type, &name, &targetID,
		&memberIDsJSON, &conv.Pinned, &conv.Muted, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, err
	}

	conv.Name = name.String
	conv.TargetID = targetID.String
	if err := json.Unmarshal([]byte(memberIDsJSON), &conv.MemberIDs); err != nil {
		conv.MemberIDs = []string{}
	}

	return &conv, nil
}

func (r *ConversationRepository) Create(workspaceID string, data ConversationCreate) (*Conversation, error) {
	now := time.Now().UnixMilli()
	id := "conv_" + utils.GenerateID()[:10]

	memberIDsJSON, err := json.Marshal(data.MemberIDs)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO conversations (id, workspace_id, type, name, target_id, member_ids, pinned, muted, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?, ?)`

	_, err = r.db.Exec(query, id, workspaceID, data.Type, data.Name, data.TargetID, memberIDsJSON, now, now)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *ConversationRepository) Update(id string, data map[string]interface{}) error {
	now := time.Now().UnixMilli()

	if name, ok := data["name"].(string); ok {
		_, err := r.db.Exec(`UPDATE conversations SET name = ?, updated_at = ? WHERE id = ?`, name, now, id)
		return err
	}
	if pinned, ok := data["pinned"].(bool); ok {
		_, err := r.db.Exec(`UPDATE conversations SET pinned = ?, updated_at = ? WHERE id = ?`, pinned, now, id)
		return err
	}
	if muted, ok := data["muted"].(bool); ok {
		_, err := r.db.Exec(`UPDATE conversations SET muted = ?, updated_at = ? WHERE id = ?`, muted, now, id)
		return err
	}

	return nil
}

func (r *ConversationRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM conversations WHERE id = ?`, id)
	return err
}

// SetMemberIDs replaces channel member list (JSON array).
func (r *ConversationRepository) SetMemberIDs(id string, memberIDs []string) error {
	now := time.Now().UnixMilli()
	memberIDsJSON, err := json.Marshal(memberIDs)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`UPDATE conversations SET member_ids = ?, updated_at = ? WHERE id = ?`, string(memberIDsJSON), now, id)
	return err
}

func (r *ConversationRepository) GetOrCreateDefaultChannel(workspaceID string, memberIDs []string) (*Conversation, error) {
	// Try to find existing default channel
	query := `SELECT id, workspace_id, type, name, target_id, member_ids, pinned, muted, created_at, updated_at
	          FROM conversations WHERE workspace_id = ? AND type = 'channel' AND name = 'general'`

	var conv Conversation
	var memberIDsJSON string
	var name, targetID sql.NullString

	err := r.db.QueryRow(query, workspaceID).Scan(&conv.ID, &conv.WorkspaceID, &conv.Type, &name, &targetID,
		&memberIDsJSON, &conv.Pinned, &conv.Muted, &conv.CreatedAt, &conv.UpdatedAt)

	if err == nil {
		conv.Name = name.String
		conv.TargetID = targetID.String
		if err := json.Unmarshal([]byte(memberIDsJSON), &conv.MemberIDs); err != nil {
			conv.MemberIDs = []string{}
		}
		return &conv, nil
	}

	if err == sql.ErrNoRows {
		// Create default channel
		return r.Create(workspaceID, ConversationCreate{
			Type:      ConversationTypeChannel,
			MemberIDs: memberIDs,
			Name:      "general",
		})
	}

	return nil, err
}