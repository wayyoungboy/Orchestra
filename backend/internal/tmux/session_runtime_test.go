package tmux_test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/tmux"
)

func tmuxRuntimeAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestTmuxSessionRuntimeRawInputAndResize(t *testing.T) {
	if !tmuxRuntimeAvailable() {
		t.Skip("tmux not installed")
	}

	ctx := context.Background()
	mgr := tmux.NewManager("")
	sessionName := fmt.Sprintf("orchestra-runtime-%d", time.Now().UnixNano())
	_ = mgr.KillSession(ctx, sessionName)

	if err := mgr.CreateSession(ctx, sessionName, "/tmp", "/bin/bash", nil); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer mgr.KillSession(ctx, sessionName)

	session := tmux.NewTmuxSession("runtime-session", "ws-runtime", "member-runtime", "Runtime Agent", "bash", sessionName, "/tmp", "/bin/bash", nil)
	if err := session.Resize(100, 30); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	cols, rows, err := mgr.GetPaneSize(ctx, sessionName)
	if err != nil {
		t.Fatalf("GetPaneSize: %v", err)
	}
	if cols != 100 || rows != 30 {
		t.Fatalf("pane size = %dx%d, want 100x30", cols, rows)
	}

	marker := fmt.Sprintf("ORCH_RUNTIME_%d", time.Now().UnixNano())
	if err := session.SendRawInput(fmt.Sprintf("printf '%s\\n'", marker)); err != nil {
		t.Fatalf("SendRawInput command: %v", err)
	}
	if err := session.SendRawInput("\r"); err != nil {
		t.Fatalf("SendRawInput enter: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		output, err := session.CaptureScrollback(ctx, 80)
		if err != nil {
			t.Fatalf("CaptureScrollback: %v", err)
		}
		if strings.Contains(output, marker) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("raw input marker %q not found in tmux scrollback", marker)
}
