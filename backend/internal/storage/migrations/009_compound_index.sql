-- Compound index for workload stats query performance
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_workspace_status
ON tasks(assignee_id, workspace_id, status);
