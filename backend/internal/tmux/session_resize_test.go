package tmux

import "testing"

func TestTmuxSessionResizeRejectsInvalidDimensions(t *testing.T) {
	sess := NewTmuxSession("sess-1", "ws-1", "member-1", "Agent", "codex", "orchestra-test", "/tmp", "bash", nil)

	if err := sess.Resize(0, 24); err == nil {
		t.Fatal("Resize accepted zero columns")
	}
	if err := sess.Resize(80, 0); err == nil {
		t.Fatal("Resize accepted zero rows")
	}
	if err := sess.Resize(-1, 24); err == nil {
		t.Fatal("Resize accepted negative columns")
	}
}
