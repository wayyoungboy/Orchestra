package tmux

import "testing"

func TestShellJoinQuotesAllArguments(t *testing.T) {
	got := shellJoin("codex", "exec", "--message", "; touch /tmp/owned", "it's literal")
	want := "'codex' 'exec' '--message' '; touch /tmp/owned' 'it'\"'\"'s literal'"
	if got != want {
		t.Fatalf("shellJoin() = %q, want %q", got, want)
	}
}

func TestBuildAndParseSessionNamePreservesFullIDs(t *testing.T) {
	workspaceID := "01KXE881Y5EV1V7RA846P9BWFB"
	memberID := "01KXE881YBFHS0M21WVPS3S13P"
	name := BuildSessionName(workspaceID, memberID)

	gotWorkspaceID, gotMemberID, ok := ParseSessionName(name)
	if !ok || gotWorkspaceID != workspaceID || gotMemberID != memberID {
		t.Fatalf("ParseSessionName(%q) = (%q, %q, %v)", name, gotWorkspaceID, gotMemberID, ok)
	}
}
