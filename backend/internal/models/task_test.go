package models

import "testing"

func TestNewTaskUsesUniqueIDs(t *testing.T) {
	ids := make(map[string]struct{}, 128)
	for i := 0; i < 128; i++ {
		task := NewTask(TaskCreate{WorkspaceID: "workspace", ConversationID: "conversation", SecretaryID: "secretary", Title: "Task"})
		if _, exists := ids[task.ID]; exists {
			t.Fatalf("NewTask() generated duplicate id %q", task.ID)
		}
		ids[task.ID] = struct{}{}
	}
}
