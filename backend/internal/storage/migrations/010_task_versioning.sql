-- Add optimistic locking version to tasks
ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
