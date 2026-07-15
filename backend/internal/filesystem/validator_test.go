package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatorRejectsLookalikeAndSymlinkEscapes(t *testing.T) {
	allowed := t.TempDir()
	validator := NewValidator([]string{allowed})

	if err := validator.ValidatePath(filepath.Join(allowed, "project")); err != nil {
		t.Fatalf("expected child path to be allowed: %v", err)
	}
	if err := validator.ValidatePath(allowed + "-other"); err == nil {
		t.Fatal("lookalike sibling path should not be allowed")
	}

	outside := t.TempDir()
	link := filepath.Join(allowed, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	if err := validator.ValidatePath(filepath.Join(link, "child")); err == nil {
		t.Fatal("path through a symlink outside the allowed root should not be allowed")
	}
}
