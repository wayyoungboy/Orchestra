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
	cfg.Security.AllowedCommands = []string{"/bin/bash"}

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
