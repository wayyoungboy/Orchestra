package repository

import (
	"database/sql"
)

type ConversationReadRepository struct {
	db *sql.DB
}

func NewConversationReadRepository(db *sql.DB) *ConversationReadRepository {
	return &ConversationReadRepository{db: db}
}

func (r *ConversationReadRepository) GetLastRead(conversationID, memberID string) (int64, error) {
	var t sql.NullInt64
	err := r.db.QueryRow(
		`SELECT last_read_at FROM conversation_reads WHERE conversation_id = ? AND member_id = ?`,
		conversationID, memberID,
	).Scan(&t)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !t.Valid {
		return 0, nil
	}
	return t.Int64, nil
}

func (r *ConversationReadRepository) Upsert(conversationID, memberID string, lastReadAt int64) error {
	_, err := r.db.Exec(`
		INSERT INTO conversation_reads (conversation_id, member_id, last_read_at)
		VALUES (?, ?, ?)
		ON CONFLICT(conversation_id, member_id) DO UPDATE SET last_read_at = excluded.last_read_at
	`, conversationID, memberID, lastReadAt)
	return err
}
