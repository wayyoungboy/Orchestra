-- Add loop detection columns to messages
ALTER TABLE messages ADD COLUMN depth INTEGER DEFAULT 0;
ALTER TABLE messages ADD COLUMN parent_message_id TEXT;
CREATE INDEX IF NOT EXISTS idx_messages_parent ON messages(parent_message_id);
