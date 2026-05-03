-- Add routing columns to outbox for workspace and member targeting
ALTER TABLE outbox ADD COLUMN workspace_id TEXT NOT NULL DEFAULT '';
ALTER TABLE outbox ADD COLUMN target_member_id TEXT NOT NULL DEFAULT '';
