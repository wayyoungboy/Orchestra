package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceBrowsePathStaysWithinWorkspace(t *testing.T) {
	workspace := t.TempDir()
	child := filepath.Join(workspace, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}

	got, err := workspaceBrowsePath(workspace, "child")
	if err != nil || got != child {
		t.Fatalf("workspaceBrowsePath() = %q, %v; want %q, nil", got, err, child)
	}
	if _, err := workspaceBrowsePath(workspace, t.TempDir()); err == nil {
		t.Fatal("an unrelated absolute path must be rejected")
	}

	escape := filepath.Join(workspace, "escape")
	if err := os.Symlink(t.TempDir(), escape); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	if _, err := workspaceBrowsePath(workspace, filepath.Join(escape, "child")); err == nil {
		t.Fatal("a symlink path outside the workspace must be rejected")
	}
}
