package api

import (
	"net/http"
	"testing"
)

func TestAssistantInternalChatCreatesOwnerCompletionNotification(t *testing.T) {
	workspaceDir := t.TempDir()
	router, _ := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name":             "Notification Workspace",
		"path":             workspaceDir,
		"ownerDisplayName": "Notification Owner",
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

	assistantResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            "Notification Assistant",
		"roleType":        "assistant",
		"terminalType":    "native",
		"terminalCommand": "/bin/cat",
		"acpEnabled":      true,
		"acpCommand":      "/bin/cat",
	})
	if assistantResp.Code != http.StatusCreated {
		t.Fatalf("create assistant status = %d body=%s", assistantResp.Code, assistantResp.Body.String())
	}
	assistant := decodeJSON[map[string]any](t, assistantResp)
	assistantID := assistant["id"].(string)

	dmResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/conversations/direct", map[string]any{
		"userId":   ownerID,
		"targetId": assistantID,
	})
	if dmResp.Code != http.StatusOK && dmResp.Code != http.StatusCreated {
		t.Fatalf("create dm status = %d body=%s", dmResp.Code, dmResp.Body.String())
	}
	dm := decodeJSON[map[string]any](t, dmResp)
	conversationID := dm["id"].(string)

	reportText := "Assistant finished notification loop"
	chatResp := requestJSON(t, router, http.MethodPost, "/api/internal/chat/send", map[string]any{
		"workspaceId":      workspaceID,
		"conversationId":   conversationID,
		"senderId":         assistantID,
		"senderName":       "Notification Assistant",
		"text":             reportText,
		"depth":            0,
		"visitedMemberIDs": []string{},
	})
	if chatResp.Code != http.StatusOK {
		t.Fatalf("internal chat send status = %d body=%s", chatResp.Code, chatResp.Body.String())
	}

	badgeResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/notifications/badge?userId="+ownerID, nil)
	if badgeResp.Code != http.StatusOK {
		t.Fatalf("badge status = %d body=%s", badgeResp.Code, badgeResp.Body.String())
	}
	badge := decodeJSON[map[string]any](t, badgeResp)
	if int(badge["unread"].(float64)) != 1 {
		t.Fatalf("unread notifications = %v, want 1", badge["unread"])
	}

	listResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/notifications?userId="+ownerID, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list notifications status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	notifications := decodeJSON[[]map[string]any](t, listResp)
	if len(notifications) != 1 {
		t.Fatalf("expected one notification, got %#v", notifications)
	}
	if notifications[0]["type"] != "agent_completion" ||
		notifications[0]["conversationId"] != conversationID ||
		notifications[0]["body"] != reportText {
		t.Fatalf("unexpected notification: %#v", notifications[0])
	}
}
