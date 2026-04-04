package chatbridge

import (
	"strings"
	"testing"
)

func TestSanitizeTTYForChat_CSI(t *testing.T) {
	in := "\x1b[31mhello\x1b[0m world"
	out := SanitizeTTYForChat(in)
	if out != "hello world" {
		t.Fatalf("got %q want %q", out, "hello world")
	}
}

func TestSanitizeTTYForChat_OrphanTokens(t *testing.T) {
	in := "[?2026h\x1b[2D@监工 [1C你会干嘛"
	out := SanitizeTTYForChat(in)
	if strings.Contains(out, "[?2026") || strings.Contains(out, "[1C") || strings.Contains(out, "\x1b") {
		t.Fatalf("expected escapes removed, got %q", out)
	}
}
