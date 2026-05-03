-- Add last_message_preview and last_message_at to conversations for list display.
ALTER TABLE conversations ADD COLUMN last_message_preview TEXT NOT NULL DEFAULT '';
ALTER TABLE conversations ADD COLUMN last_message_at INTEGER NOT NULL DEFAULT 0;
