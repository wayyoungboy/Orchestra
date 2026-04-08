-- Add updated_at column to api_keys table
ALTER TABLE api_keys ADD COLUMN updated_at INTEGER NOT NULL DEFAULT 0;

-- Update existing records to have updated_at = created_at
UPDATE api_keys SET updated_at = created_at WHERE updated_at = 0;