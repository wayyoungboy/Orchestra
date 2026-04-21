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

// BatchGetUnreadCounts returns a map of conversationID -> unread count for the given viewer.
// Reads are fetched in a single query, then message counts are computed.
func (r *ConversationReadRepository) BatchGetUnreadCounts(conversationIDs []string, viewerMemberID string) (map[string]int, error) {
	if len(conversationIDs) == 0 {
		return make(map[string]int), nil
	}

	// Fetch all last_read_at values in one query
	placeholders := make([]string, len(conversationIDs))
	args := make([]interface{}, 0, len(conversationIDs)+1)
	for i, id := range conversationIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, viewerMemberID)

	query := `
		SELECT conversation_id, last_read_at
		FROM conversation_reads
		WHERE conversation_id IN (` + joinPlaceholders(len(conversationIDs)) + `)
		  AND member_id = ?
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lastRead := make(map[string]int64)
	for rows.Next() {
		var convID string
		var ts int64
		if err := rows.Scan(&convID, &ts); err != nil {
			return nil, err
		}
		lastRead[convID] = ts
	}

	// Compute unread counts per conversation in a second batch query
	counts := make(map[string]int)
	for _, convID := range conversationIDs {
		lr := lastRead[convID] // 0 if never read
		var n int
		r.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE conversation_id = ? AND sender_id != ? AND created_at > ?`,
			convID, viewerMemberID, lr,
		).Scan(&n)
		counts[convID] = n
	}
	return counts, nil
}

func joinPlaceholders(n int) string {
	s := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			s = append(s, ',')
		}
		s = append(s, '?')
	}
	return string(s)
}

