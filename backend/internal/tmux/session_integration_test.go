//go:build integration
// +build integration

package tmux_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/tmux"
)

func tmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestTmuxSessionPipePaneOutput(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	ctx := context.Background()
	mgr := tmux.NewManager("")
	sessName := "orchestra-test-pipepane"

	// Cleanup in case of leftover
	_ = mgr.KillSession(ctx, sessName)
	time.Sleep(200 * time.Millisecond)

	// Create a bash session
	if err := mgr.CreateSession(ctx, sessName, "/tmp", "bash", nil); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer mgr.KillSession(ctx, sessName)

	// Create TmuxSession wrapper
	ts := tmux.NewTmuxSession("test-sess", "ws1", "m1", "TestAgent", "bash", sessName, "/tmp", "bash", nil)

	// Setup pipe-pane
	if err := ts.SetupPipePane(ctx); err != nil {
		t.Fatalf("SetupPipePane: %v", err)
	}

	// Start output reader
	if err := ts.StartOutputReader(ctx, false); err != nil {
		t.Fatalf("StartOutputReader: %v", err)
	}

	// Wait for shell to be ready
	time.Sleep(500 * time.Millisecond)

	// Send a unique echo command
	marker := "ORCH_TEST_" + time.Now().Format("150405")
	if err := mgr.SendKeys(ctx, sessName, "echo "+marker); err != nil {
		t.Fatalf("SendKeys: %v", err)
	}
	if err := mgr.SendEnter(ctx, sessName); err != nil {
		t.Fatalf("SendEnter: %v", err)
	}

	// Wait for output to appear
	deadline := time.After(5 * time.Second)
	found := false
	for !found {
		select {
		case msg := <-ts.OutputChan:
			content := string(msg.Content)
			t.Logf("Output: type=%s content=%s", msg.Type, content)
			if len(content) > 0 {
				found = true
			}
		case <-deadline:
			t.Fatal("Timeout waiting for pipe-pane output")
		}
	}

	t.Log("Pipe-pane output capture verified")
}

func TestTmuxManagerBasicLifecycle(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	ctx := context.Background()
	mgr := tmux.NewManager("")
	sessName := "orchestra-test-lifecycle"

	// Cleanup
	_ = mgr.KillSession(ctx, sessName)

	// Create
	if err := mgr.CreateSession(ctx, sessName, "/tmp", "bash", nil); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer mgr.KillSession(ctx, sessName)

	// Verify exists
	if !mgr.SessionExists(sessName) {
		t.Fatal("Session should exist after creation")
	}

	// List
	sessions, err := mgr.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s == sessName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Session %s not found in list: %v", sessName, sessions)
	}

	// Capture pane
	time.Sleep(300 * time.Millisecond)
	output, err := mgr.CapturePane(ctx, sessName, 10)
	if err != nil {
		t.Fatalf("CapturePane: %v", err)
	}
	t.Logf("CapturePane output: %q", output)

	// Kill
	if err := mgr.KillSession(ctx, sessName); err != nil {
		t.Fatalf("KillSession: %v", err)
	}

	// Verify gone
	if mgr.SessionExists(sessName) {
		t.Fatal("Session should not exist after kill")
	}
}
