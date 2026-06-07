package api

import (
	"net/http"
	"strings"
	"testing"
)

func TestMentionDispatchAcquireFailureCreatesOutboxDiagnostic(t *testing.T) {
	workspaceDir := t.TempDir()
	router, _ := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name":             "Dispatch Diagnostics Workspace",
		"path":             workspaceDir,
		"ownerDisplayName": "Dispatch Owner",
	})
	if workspaceResp.Code != http.StatusCreated {
		t.Fatalf("create workspace status = %d body=%s", workspaceResp.Code, workspaceResp.Body.String())
	}
	workspace := decodeJSON[map[string]any](t, workspaceResp)
	workspaceID := workspace["id"].(string)

	membersResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/members", nil)
	if membersResp.Code != http.StatusOK {
		t.Fatalf("list members status = %d body=%s", membersResp.Code, membersResp.Body.String())
	}
	members := decodeJSON[[]map[string]any](t, membersResp)
	var ownerID string
	for _, member := range members {
		if member["roleType"] == "owner" {
			ownerID = member["id"].(string)
			break
		}
	}
	if ownerID == "" {
		t.Fatalf("owner member not found: %#v", members)
	}

	assistantName := "Broken Dispatch Assistant"
	assistantResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            assistantName,
		"roleType":        "assistant",
		"terminalType":    "native",
		"terminalCommand": "/bin/definitely-missing-orchestra-agent",
		"acpEnabled":      true,
		"acpCommand":      "/bin/definitely-missing-orchestra-agent",
	})
	if assistantResp.Code != http.StatusCreated {
		t.Fatalf("create assistant status = %d body=%s", assistantResp.Code, assistantResp.Body.String())
	}
	assistant := decodeJSON[map[string]any](t, assistantResp)
	assistantID := assistant["id"].(string)

	conversationsResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations?userId="+ownerID, nil)
	if conversationsResp.Code != http.StatusOK {
		t.Fatalf("list conversations status = %d body=%s", conversationsResp.Code, conversationsResp.Body.String())
	}
	conversations := decodeJSON[map[string]any](t, conversationsResp)
	conversationID, _ := conversations["defaultChannelId"].(string)
	if conversationID == "" {
		t.Fatalf("default channel missing: %#v", conversations)
	}

	userText := "please surface this dispatch failure"
	messageResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/conversations/"+conversationID+"/messages", map[string]any{
		"text":       "@" + assistantName + " " + userText,
		"senderId":   ownerID,
		"senderName": "Dispatch Owner",
		"mentionIds": []string{assistantID},
	})
	if messageResp.Code != http.StatusCreated {
		t.Fatalf("send message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}

	outboxResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/outbox?conversationId="+conversationID, nil)
	if outboxResp.Code != http.StatusOK {
		t.Fatalf("list outbox status = %d body=%s", outboxResp.Code, outboxResp.Body.String())
	}
	outboxBody := decodeJSON[map[string]any](t, outboxResp)
	items, _ := outboxBody["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected one outbox diagnostic item, got %#v", outboxBody["items"])
	}
	item, _ := items[0].(map[string]any)
	if item["conversation_id"] != conversationID || item["target_member_id"] != assistantID {
		t.Fatalf("unexpected outbox diagnostic item: %#v", item)
	}
	content, _ := item["content"].(string)
	if !strings.Contains(content, userText) || !strings.Contains(content, "#conversationId{"+conversationID+"}") {
		t.Fatalf("outbox content did not preserve dispatch prompt: %q", content)
	}
}
