package a2a

import "testing"

func TestSessionResizeRejectsInvalidDimensions(t *testing.T) {
	sess := NewSession("sess-1", "ws-1", "member-1", "Agent", "codex", nil)

	if err := sess.Resize(0, 24); err == nil {
		t.Fatal("Resize accepted zero columns")
	}
	if err := sess.Resize(80, 0); err == nil {
		t.Fatal("Resize accepted zero rows")
	}
}
