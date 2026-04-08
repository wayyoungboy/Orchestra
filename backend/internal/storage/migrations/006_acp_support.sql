-- 006_acp_support.sql
-- Add ACP (Agent Communication Protocol) support columns to members table

ALTER TABLE members ADD COLUMN acp_enabled INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN acp_command TEXT;
ALTER TABLE members ADD COLUMN acp_args TEXT;

-- Update existing members to use PTY mode (acp_enabled = 0) by default
UPDATE members SET acp_enabled = 0 WHERE acp_enabled IS NULL;