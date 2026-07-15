package a2a

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/tmux"
)

func TestForwardTmuxOutput(t *testing.T) {
	pool := NewPool(time.Hour, "")
	tmuxSession := tmux.NewTmuxSession("session-1", "workspace-1", "member-1", "agent", "", "tmux-1", "", "", nil)
	session := NewSession("session-1", "workspace-1", "member-1", "agent", "", tmuxSession)

	forwarded := make(chan struct{})
	go func() {
		pool.forwardTmuxOutput(session, tmuxSession)
		close(forwarded)
	}()

	want := json.RawMessage(`{"type":"assistant_message","content":"hello"}`)
	tmuxSession.OutputChan <- &tmux.ACPMessage{Type: tmux.TypeAssistantMessage, Content: want}

	select {
	case got := <-session.OutputChan:
		if got.Type != TypeAssistantMessage {
			t.Fatalf("expected assistant message, got %q", got.Type)
		}
		if string(got.Content) != string(want) {
			t.Fatalf("expected %s, got %s", want, got.Content)
		}
	case <-time.After(time.Second):
		t.Fatal("tmux output was not forwarded to the session")
	}

	close(tmuxSession.DoneChan)
	select {
	case <-forwarded:
	case <-time.After(time.Second):
		t.Fatal("forwarder did not stop after tmux shutdown")
	}
}

func TestTerminalSubscribersReceiveIndependentCopies(t *testing.T) {
	session := NewSession("session-1", "workspace-1", "member-1", "agent", "", nil)
	outputA, streamA, errorsA, unsubscribeA := session.SubscribeTerminal()
	outputB, streamB, errorsB, unsubscribeB := session.SubscribeTerminal()
	defer unsubscribeA()
	defer unsubscribeB()

	message := NewUserMessage("hello")
	session.PublishOutput(message)
	for label, ch := range map[string]<-chan *ACPMessage{"A": outputA, "B": outputB} {
		select {
		case got := <-ch:
			if got != message {
				t.Fatalf("subscriber %s received an unexpected message", label)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %s did not receive terminal output", label)
		}
	}

	session.TrySendChatStream([]byte(`{"type":"stream"}`))
	for label, ch := range map[string]<-chan []byte{"A": streamA, "B": streamB} {
		select {
		case got := <-ch:
			if string(got) != `{"type":"stream"}` {
				t.Fatalf("subscriber %s received %q", label, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %s did not receive chat stream output", label)
		}
	}

	wantErr := &testError{}
	session.PublishError(wantErr)
	for label, ch := range map[string]<-chan error{"A": errorsA, "B": errorsB} {
		select {
		case got := <-ch:
			if got != wantErr {
				t.Fatalf("subscriber %s received an unexpected error", label)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %s did not receive terminal error", label)
		}
	}
}

func TestExecutionPolicyRejectsUntrustedCommandsAndPaths(t *testing.T) {
	workspace := t.TempDir()
	pool := NewPool(time.Hour, "")
	pool.SetExecutionPolicy(security.NewWhitelist([]string{"codex"}, []string{workspace}))

	if err := pool.validateExecutionConfig("codex", workspace); err != nil {
		t.Fatalf("expected allowed command and workspace: %v", err)
	}
	if err := pool.validateExecutionConfig("sh", workspace); err == nil {
		t.Fatal("expected an untrusted command to be rejected")
	}
	if err := pool.validateExecutionConfig("codex", workspace+"-other"); err == nil {
		t.Fatal("expected a lookalike path outside the workspace to be rejected")
	}
}

type testError struct{}

func (*testError) Error() string { return "test error" }
