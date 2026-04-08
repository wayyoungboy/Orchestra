-- 008_a2a_support.sql
-- Add A2A (Agent-to-Agent Protocol) support

-- A2A Agent registry table
CREATE TABLE IF NOT EXISTS a2a_agents (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL REFERENCES workspaces(id),
    member_id TEXT NOT NULL REFERENCES members(id),
    agent_url TEXT NOT NULL,
    auth_type TEXT DEFAULT 'none',
    auth_token TEXT,
    status TEXT DEFAULT 'registered',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- A2A Task tracking table
CREATE TABLE IF NOT EXISTS a2a_tasks (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    a2a_task_id TEXT NOT NULL UNIQUE,
    a2a_agent_url TEXT NOT NULL,
    conversation_id TEXT,
    status TEXT DEFAULT 'submitted',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    completed_at INTEGER,
    error_message TEXT
);

-- Index for fast lookup by workspace+member
CREATE INDEX IF NOT EXISTS idx_a2a_agents_workspace_member ON a2a_agents(workspace_id, member_id);
CREATE INDEX IF NOT EXISTS idx_a2a_tasks_workspace_member ON a2a_tasks(workspace_id, member_id);
CREATE INDEX IF NOT EXISTS idx_a2a_tasks_a2a_task_id ON a2a_tasks(a2a_task_id);

-- Extend members table with A2A configuration
ALTER TABLE members ADD COLUMN a2a_enabled INTEGER DEFAULT 0;
ALTER TABLE members ADD COLUMN a2a_agent_url TEXT;
ALTER TABLE members ADD COLUMN a2a_auth_type TEXT DEFAULT 'none';
ALTER TABLE members ADD COLUMN a2a_auth_token TEXT;
