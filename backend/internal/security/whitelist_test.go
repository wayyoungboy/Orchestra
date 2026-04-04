package security

import (
	"testing"
)

func TestWhitelist_ValidateCommand(t *testing.T) {
	w := NewWhitelist([]string{"claude", "gemini", "python"}, nil)

	if err := w.ValidateCommand("claude"); err != nil {
		t.Errorf("claude should be allowed: %v", err)
	}
	if err := w.ValidateCommand("/usr/bin/claude"); err != nil {
		t.Errorf("/usr/bin/claude should be allowed: %v", err)
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
}

func TestWhitelist_Empty(t *testing.T) {
	w := NewWhitelist(nil, nil)

	if err := w.ValidateCommand("claude"); err == nil {
		t.Error("empty whitelist should deny all commands")
	}
}