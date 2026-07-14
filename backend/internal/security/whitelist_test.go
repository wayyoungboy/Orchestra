package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWhitelist_ValidateCommand(t *testing.T) {
	w := NewWhitelist([]string{"claude", "gemini", "python", "/bin/sh"}, nil)

	if err := w.ValidateCommand("claude"); err != nil {
		t.Errorf("claude should be allowed: %v", err)
	}
	if err := w.ValidateCommand("./claude"); err == nil {
		t.Error("an unapproved relative command path should not be allowed")
	}
	if err := w.ValidateCommand("/tmp/claude"); err == nil {
		t.Error("an unapproved absolute command path should not be allowed")
	}
	if err := w.ValidateCommand("/bin/sh"); err != nil {
		t.Errorf("an explicitly allowed absolute command should be allowed: %v", err)
	}
	if err := w.ValidateCommand("rm"); err == nil {
		t.Error("rm should not be allowed")
	}
}

func TestWhitelist_ValidatePath(t *testing.T) {
	w := NewWhitelist(nil, []string{"/home/user/projects", "/var/workspaces"})

	if err := w.ValidatePath("/home/user/projects/myapp"); err != nil {
		t.Errorf("projects/myapp should be allowed: %v", err)
	}
	if err := w.ValidatePath("/etc/passwd"); err == nil {
		t.Error("/etc/passwd should not be allowed")
	}
	if err := w.ValidatePath("/home/user/projects-evil"); err == nil {
		t.Error("a lookalike path outside an allowed directory should not be allowed")
	}
}

func TestWhitelist_Empty(t *testing.T) {
	w := NewWhitelist(nil, nil)

	if err := w.ValidateCommand("claude"); err == nil {
		t.Error("empty whitelist should deny all commands")
	}
}

func TestWhitelistRejectsSymlinkEscape(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(allowed, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	w := NewWhitelist(nil, []string{allowed})
	if err := w.ValidatePath(filepath.Join(link, "new-file")); err == nil {
		t.Fatal("path through a symlink outside an allowed root should be denied")
	}
}

func TestWhitelistExpandsHomeDirectoryPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	w := NewWhitelist(nil, []string{"~/projects"})

	if err := w.ValidatePath(filepath.Join(home, "projects", "orchestra")); err != nil {
		t.Fatalf("expected home-relative path to be allowed: %v", err)
	}
}
