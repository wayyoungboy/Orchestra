-- Per-member read cursor for conversations (reference-desktop-style unread)
CREATE TABLE IF NOT EXISTS conversation_reads (
    conversation_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    last_read_at INTEGER NOT NULL,
    PRIMARY KEY (conversation_id, member_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_conversation_reads_member ON conversation_reads(member_id);
