package handlers

import (
	"encoding/json"
	"testing"

	"github.com/orchestra/backend/internal/models"
)

func TestTaskStatusEventPayloadEscapesTaskTitle(t *testing.T) {
	task := &models.Task{
		ID:          "task-1",
		WorkspaceID: "workspace-1",
		AssigneeID:  "assistant-1",
		Title:       "quote \" and slash \\ and newline\n",
	}

	payload, err := taskStatusEventPayload(task, models.TaskStatusCompleted)
	if err != nil {
		t.Fatalf("taskStatusEventPayload returned error: %v", err)
	}

	var event map[string]string
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("payload is not valid JSON: %v\npayload=%s", err, payload)
	}

	if event["type"] != "task_status" {
		t.Fatalf("type = %q, want task_status", event["type"])
	}
	if event["workspaceId"] != task.WorkspaceID {
		t.Fatalf("workspaceId = %q, want %q", event["workspaceId"], task.WorkspaceID)
	}
	if event["taskId"] != task.ID {
		t.Fatalf("taskId = %q, want %q", event["taskId"], task.ID)
	}
	if event["status"] != string(models.TaskStatusCompleted) {
		t.Fatalf("status = %q, want %q", event["status"], models.TaskStatusCompleted)
	}
	if event["assigneeId"] != task.AssigneeID {
		t.Fatalf("assigneeId = %q, want %q", event["assigneeId"], task.AssigneeID)
	}
	if event["title"] != task.Title {
		t.Fatalf("title = %q, want %q", event["title"], task.Title)
	}
}
