package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/agent"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/ws"
)

func tmuxRuntimeAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func newRuntimeRouter(t *testing.T, workspaceDir string) (*gin.Engine, *Dependencies) {
	t.Helper()

	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("new database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	migrations := filepath.Clean("../storage/migrations")
	if err := db.Migrate(migrations); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := config.Default()
	cfg.Auth.Enabled = false
	cfg.Security.AllowedPaths = []string{workspaceDir}
	cfg.Security.AllowedCommands = []string{"/bin/bash", "/bin/cat"}

	registry := agent.NewRegistry()
	gateway := ws.NewGateway(nil, cfg.Security.AllowedOrigins)
	router, _, deps := SetupRouter(registry, gateway, db, cfg)
	t.Cleanup(func() { deps.Stop() })

	return router, deps
}

func requestJSON(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return out
}

func TestTerminalRuntimeAPIWorkspaceMemberSessionLifecycle(t *testing.T) {
	if !tmuxRuntimeAvailable() {
		t.Skip("tmux not installed")
	}

	workspaceDir := t.TempDir()
	router, deps := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name":             "Runtime Workspace",
		"path":             workspaceDir,
		"ownerDisplayName": "Runtime Owner",
	})
	if workspaceResp.Code != http.StatusCreated {
		t.Fatalf("create workspace status = %d body=%s", workspaceResp.Code, workspaceResp.Body.String())
	}
	workspace := decodeJSON[map[string]any](t, workspaceResp)
	workspaceID := workspace["id"].(string)

	memberResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            "Runtime Shell",
		"roleType":        "assistant",
		"terminalType":    "bash",
		"terminalCommand": "/bin/bash",
		"acpEnabled":      true,
		"acpCommand":      "/bin/bash",
	})
	if memberResp.Code != http.StatusCreated {
		t.Fatalf("create member status = %d body=%s", memberResp.Code, memberResp.Body.String())
	}
	member := decodeJSON[map[string]any](t, memberResp)
	memberID := member["id"].(string)

	sessionResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members/"+memberID+"/terminal-session", map[string]any{})
	if sessionResp.Code != http.StatusCreated {
		t.Fatalf("create terminal session status = %d body=%s", sessionResp.Code, sessionResp.Body.String())
	}
	session := decodeJSON[map[string]any](t, sessionResp)
	sessionID := session["sessionId"].(string)
	defer func() {
		if s := deps.Registry.GetByID(sessionID); s != nil {
			_ = s.Kill()
			deps.Registry.Unregister(sessionID)
		}
	}()

	listResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/terminal-sessions", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list sessions status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	list := decodeJSON[map[string][]map[string]string](t, listResp)
	if len(list["sessions"]) != 1 || list["sessions"][0]["sessionId"] != sessionID || list["sessions"][0]["memberId"] != memberID {
		t.Fatalf("unexpected session list: %#v", list)
	}

	marker := "ORCH_API_RUNTIME"
	if transport := deps.Registry.GetByID(sessionID).Transport(); transport != nil {
		if err := transport.SendRawInput("printf '" + marker + "\\n'"); err != nil {
			t.Fatalf("send raw command: %v", err)
		}
		if err := transport.SendRawInput("\r"); err != nil {
			t.Fatalf("send raw enter: %v", err)
		}
	}

	var snapshot map[string]any
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snapshotResp := requestJSON(t, router, http.MethodGet, "/api/terminals/"+sessionID+"/snapshot?lines=80", nil)
		if snapshotResp.Code != http.StatusOK {
			t.Fatalf("snapshot status = %d body=%s", snapshotResp.Code, snapshotResp.Body.String())
		}
		snapshot = decodeJSON[map[string]any](t, snapshotResp)
		if snapshot["sessionId"] != sessionID {
			t.Fatalf("snapshot sessionId = %v, want %s", snapshot["sessionId"], sessionID)
		}
		content, _ := snapshot["content"].(string)
		if strings.Contains(content, marker) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	content, _ := snapshot["content"].(string)
	if !strings.Contains(content, marker) {
		t.Fatalf("snapshot did not contain marker %q: %#v", marker, snapshot)
	}

	deleteResp := requestJSON(t, router, http.MethodDelete, "/api/terminals/"+sessionID, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	if got := deps.Registry.GetByID(sessionID); got != nil {
		t.Fatalf("expected session to be unregistered after delete, got %#v", got)
	}
}

func TestMentionedAssistantInDefaultChannelCreatesDispatchSession(t *testing.T) {
	if !tmuxRuntimeAvailable() {
		t.Skip("tmux not installed")
	}

	workspaceDir := t.TempDir()
	router, deps := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name":             "Dispatch Workspace",
		"path":             workspaceDir,
		"ownerDisplayName": "Dispatch Owner",
	})
	if workspaceResp.Code != http.StatusCreated {
		t.Fatalf("create workspace status = %d body=%s", workspaceResp.Code, workspaceResp.Body.String())
	}
	workspace := decodeJSON[map[string]any](t, workspaceResp)
	workspaceID := workspace["id"].(string)

	memberResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            "Dispatch Shell",
		"roleType":        "assistant",
		"terminalType":    "native",
		"terminalCommand": "/bin/cat",
		"acpEnabled":      true,
		"acpCommand":      "/bin/cat",
	})
	if memberResp.Code != http.StatusCreated {
		t.Fatalf("create member status = %d body=%s", memberResp.Code, memberResp.Body.String())
	}
	member := decodeJSON[map[string]any](t, memberResp)
	memberID := member["id"].(string)

	conversationsResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations", nil)
	if conversationsResp.Code != http.StatusOK {
		t.Fatalf("list conversations status = %d body=%s", conversationsResp.Code, conversationsResp.Body.String())
	}
	conversations := decodeJSON[map[string]any](t, conversationsResp)
	conversationID := conversations["defaultChannelId"].(string)

	marker := "ORCH_DISPATCH_RUNTIME"
	messageResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/conversations/"+conversationID+"/messages", map[string]any{
		"text":       "Please inspect " + marker,
		"senderId":   "owner",
		"senderName": "Dispatch Owner",
		"mentionIds": []string{memberID},
	})
	if messageResp.Code != http.StatusCreated {
		t.Fatalf("send message status = %d body=%s", messageResp.Code, messageResp.Body.String())
	}

	var sessionID string
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if session := deps.Registry.GetByMember(workspaceID, memberID); session != nil {
			sessionID = session.ID
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if sessionID == "" {
		t.Fatalf("expected mentioned assistant %s to get a dispatch session", memberID)
	}
	defer func() {
		if s := deps.Registry.GetByID(sessionID); s != nil {
			_ = s.Kill()
			deps.Registry.Unregister(sessionID)
		}
	}()

	var snapshot map[string]any
	deadline = time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		snapshotResp := requestJSON(t, router, http.MethodGet, "/api/terminals/"+sessionID+"/snapshot?lines=120", nil)
		if snapshotResp.Code != http.StatusOK {
			t.Fatalf("snapshot status = %d body=%s", snapshotResp.Code, snapshotResp.Body.String())
		}
		snapshot = decodeJSON[map[string]any](t, snapshotResp)
		content, _ := snapshot["content"].(string)
		if strings.Contains(content, "#conversationId{"+conversationID+"}") && strings.Contains(content, marker) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	content, _ := snapshot["content"].(string)
	t.Fatalf("dispatch prompt did not reach assistant terminal; wanted conversation marker and %q in %#v", marker, content)
}

func TestAssistantResultCompletesTaskAndForwardsToSecretary(t *testing.T) {
	if !tmuxRuntimeAvailable() {
		t.Skip("tmux not installed")
	}

	workspaceDir := t.TempDir()
	router, deps := newRuntimeRouter(t, workspaceDir)

	workspaceResp := requestJSON(t, router, http.MethodPost, "/api/workspaces", map[string]any{
		"name":             "Result Loop Workspace",
		"path":             workspaceDir,
		"ownerDisplayName": "Loop Owner",
	})
	if workspaceResp.Code != http.StatusCreated {
		t.Fatalf("create workspace status = %d body=%s", workspaceResp.Code, workspaceResp.Body.String())
	}
	workspace := decodeJSON[map[string]any](t, workspaceResp)
	workspaceID := workspace["id"].(string)

	secretaryResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            "Loop Secretary",
		"roleType":        "secretary",
		"terminalType":    "native",
		"terminalCommand": "/bin/cat",
		"acpEnabled":      true,
		"acpCommand":      "/bin/cat",
	})
	if secretaryResp.Code != http.StatusCreated {
		t.Fatalf("create secretary status = %d body=%s", secretaryResp.Code, secretaryResp.Body.String())
	}
	secretary := decodeJSON[map[string]any](t, secretaryResp)
	secretaryID := secretary["id"].(string)

	assistantResp := requestJSON(t, router, http.MethodPost, "/api/workspaces/"+workspaceID+"/members", map[string]any{
		"name":            "Loop Assistant",
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

	conversationsResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations", nil)
	if conversationsResp.Code != http.StatusOK {
		t.Fatalf("list conversations status = %d body=%s", conversationsResp.Code, conversationsResp.Body.String())
	}
	conversations := decodeJSON[map[string]any](t, conversationsResp)
	conversationID := conversations["defaultChannelId"].(string)

	taskResp := requestJSON(t, router, http.MethodPost, "/api/internal/tasks/create", map[string]any{
		"workspaceId":    workspaceID,
		"conversationId": conversationID,
		"secretaryId":    secretaryID,
		"title":          "Verify result loop",
		"description":    "Exercise the assistant -> task -> chat -> secretary return path.",
		"assigneeId":     assistantID,
	})
	if taskResp.Code != http.StatusCreated {
		t.Fatalf("create task status = %d body=%s", taskResp.Code, taskResp.Body.String())
	}
	task := decodeJSON[map[string]any](t, taskResp)
	taskID := task["taskId"].(string)

	startResp := requestJSON(t, router, http.MethodPost, "/api/internal/tasks/start", map[string]any{
		"taskId": taskID,
	})
	if startResp.Code != http.StatusOK {
		t.Fatalf("start task status = %d body=%s", startResp.Code, startResp.Body.String())
	}

	resultMarker := "ORCH_RESULT_LOOP"
	completeResp := requestJSON(t, router, http.MethodPost, "/api/internal/tasks/complete", map[string]any{
		"taskId":        taskID,
		"resultSummary": "completed " + resultMarker,
	})
	if completeResp.Code != http.StatusOK {
		t.Fatalf("complete task status = %d body=%s", completeResp.Code, completeResp.Body.String())
	}

	tasksResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/tasks", nil)
	if tasksResp.Code != http.StatusOK {
		t.Fatalf("list tasks status = %d body=%s", tasksResp.Code, tasksResp.Body.String())
	}
	tasksBody := decodeJSON[map[string]any](t, tasksResp)
	tasks, ok := tasksBody["tasks"].([]any)
	if !ok || len(tasks) != 1 {
		t.Fatalf("unexpected tasks response: %#v", tasksBody)
	}
	listedTask := tasks[0].(map[string]any)
	if listedTask["status"] != "completed" || !strings.Contains(listedTask["resultSummary"].(string), resultMarker) {
		t.Fatalf("task did not retain completed result: %#v", listedTask)
	}

	reportText := "Assistant finished " + resultMarker
	chatResp := requestJSON(t, router, http.MethodPost, "/api/internal/chat/send", map[string]any{
		"workspaceId":      workspaceID,
		"conversationId":   conversationID,
		"senderId":         assistantID,
		"senderName":       "Loop Assistant",
		"text":             reportText,
		"depth":            0,
		"visitedMemberIDs": []string{},
	})
	if chatResp.Code != http.StatusOK {
		t.Fatalf("internal chat send status = %d body=%s", chatResp.Code, chatResp.Body.String())
	}

	messagesResp := requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations/"+conversationID+"/messages", nil)
	if messagesResp.Code != http.StatusOK {
		t.Fatalf("list messages status = %d body=%s", messagesResp.Code, messagesResp.Body.String())
	}
	messages := decodeJSON[[]map[string]any](t, messagesResp)
	if len(messages) != 1 {
		t.Fatalf("expected one persisted assistant message, got %#v", messages)
	}
	content := messages[0]["content"].(map[string]any)
	if messages[0]["senderId"] != assistantID || messages[0]["isAi"] != true || content["text"] != reportText {
		t.Fatalf("assistant report was not persisted as an AI message: %#v", messages[0])
	}

	conversationsResp = requestJSON(t, router, http.MethodGet, "/api/workspaces/"+workspaceID+"/conversations", nil)
	if conversationsResp.Code != http.StatusOK {
		t.Fatalf("reload conversations status = %d body=%s", conversationsResp.Code, conversationsResp.Body.String())
	}
	conversations = decodeJSON[map[string]any](t, conversationsResp)
	timeline := conversations["timeline"].([]any)
	if len(timeline) != 1 {
		t.Fatalf("unexpected conversation list: %#v", conversations)
	}
	defaultConversation := timeline[0].(map[string]any)
	if defaultConversation["lastMessagePreview"] != reportText {
		t.Fatalf("last message preview = %q, want %q", defaultConversation["lastMessagePreview"], reportText)
	}

	var secretarySessionID string
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if session := deps.Registry.GetByMember(workspaceID, secretaryID); session != nil {
			secretarySessionID = session.ID
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if secretarySessionID == "" {
		t.Fatalf("expected assistant report to create a secretary session")
	}
	defer func() {
		if s := deps.Registry.GetByID(secretarySessionID); s != nil {
			_ = s.Kill()
			deps.Registry.Unregister(secretarySessionID)
		}
	}()

	var snapshot map[string]any
	deadline = time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		snapshotResp := requestJSON(t, router, http.MethodGet, "/api/terminals/"+secretarySessionID+"/snapshot?lines=300", nil)
		if snapshotResp.Code != http.StatusOK {
			t.Fatalf("snapshot status = %d body=%s", snapshotResp.Code, snapshotResp.Body.String())
		}
		snapshot = decodeJSON[map[string]any](t, snapshotResp)
		terminalContent, _ := snapshot["content"].(string)
		if strings.Contains(terminalContent, "#conversationId{"+conversationID+"}") &&
			strings.Contains(terminalContent, "[助手汇报结果]") &&
			strings.Contains(terminalContent, reportText) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	terminalContent, _ := snapshot["content"].(string)
	t.Fatalf("assistant report did not reach secretary terminal; wanted conversation marker and %q in %#v", reportText, terminalContent)
}
