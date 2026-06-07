package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/ws"
)

func TestAssignTaskBroadcastsUpdatedAssignee(t *testing.T) {
	workspaceDir := t.TempDir()
	router, _ := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name": "Task Broadcast Workspace",
		"path": workspaceDir,
	})
	if workspaceResp.Code != http.StatusCreated {
		t.Fatalf("create workspace status = %d body=%s", workspaceResp.Code, workspaceResp.Body.String())
	}
	workspace := decodeJSON[map[string]any](t, workspaceResp)
	workspaceID := workspace["id"].(string)

	secretaryResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":     "Planner",
		"roleType": "secretary",
	})
	if secretaryResp.Code != http.StatusCreated {
		t.Fatalf("create secretary status = %d body=%s", secretaryResp.Code, secretaryResp.Body.String())
	}
	secretary := decodeJSON[map[string]any](t, secretaryResp)

	assistantResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":     "Builder",
		"roleType": "assistant",
	})
	if assistantResp.Code != http.StatusCreated {
		t.Fatalf("create assistant status = %d body=%s", assistantResp.Code, assistantResp.Body.String())
	}
	assistant := decodeJSON[map[string]any](t, assistantResp)
	assistantID := assistant["id"].(string)

	conversationsResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations", nil)
	if conversationsResp.Code != http.StatusOK {
		t.Fatalf("list conversations status = %d body=%s", conversationsResp.Code, conversationsResp.Body.String())
	}
	conversations := decodeJSON[map[string]any](t, conversationsResp)
	conversationID := conversations["defaultChannelId"].(string)

	createResp := requestJSON(t, router, http.MethodPost, "/api/internal/tasks/create", map[string]any{
		"workspaceId":    workspaceID,
		"conversationId": conversationID,
		"secretaryId":    secretary["id"].(string),
		"title":          "Assign me live",
	})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create task status = %d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeJSON[map[string]any](t, createResp)
	task := created["task"].(map[string]any)
	taskID := task["id"].(string)

	client := &ws.ChatClient{
		ID:          "test-task-broadcast-" + taskID,
		WorkspaceID: workspaceID,
		Send:        make(chan []byte, 4),
		Quit:        make(chan struct{}),
	}
	ws.GlobalChatHub.Register(client)
	t.Cleanup(func() {
		ws.GlobalChatHub.Unregister(client.ID, workspaceID)
	})

	assignResp := requestJSON(t, router, http.MethodPost, "/api/internal/tasks/assign", map[string]any{
		"taskId":     taskID,
		"assigneeId": assistantID,
	})
	if assignResp.Code != http.StatusOK {
		t.Fatalf("assign task status = %d body=%s", assignResp.Code, assignResp.Body.String())
	}

	select {
	case raw := <-client.Send:
		var event map[string]string
		if err := json.Unmarshal(raw, &event); err != nil {
			t.Fatalf("decode task_status event %q: %v", raw, err)
		}
		if event["type"] != "task_status" {
			t.Fatalf("event type = %q, want task_status: %#v", event["type"], event)
		}
		if event["taskId"] != taskID {
			t.Fatalf("event taskId = %q, want %q", event["taskId"], taskID)
		}
		if event["assigneeId"] != assistantID {
			t.Fatalf("event assigneeId = %q, want %q", event["assigneeId"], assistantID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for task_status event")
	}
}
