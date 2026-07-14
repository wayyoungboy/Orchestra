package tmux

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAvailableLogLinesWaitsForCompleteRecord(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "agent.log")
	if err := os.WriteFile(logFile, []byte("partial"), 0o600); err != nil {
		t.Fatalf("write partial log: %v", err)
	}

	session := NewTmuxSession("session", "workspace", "member", "agent", "", "tmux", "", "", nil)
	file, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	readAny, err := session.readAvailableLogLines(file)
	file.Close()
	if err != nil || readAny || session.readOffset != 0 {
		t.Fatalf("partial record should remain unread: readAny=%v offset=%d err=%v", readAny, session.readOffset, err)
	}
	select {
	case message := <-session.OutputChan:
		t.Fatalf("partial record was emitted: %#v", message)
	default:
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open log for append: %v", err)
	}
	if _, err := f.WriteString(" record\n"); err != nil {
		f.Close()
		t.Fatalf("append log: %v", err)
	}
	f.Close()

	file, err = os.Open(logFile)
	if err != nil {
		t.Fatalf("reopen log: %v", err)
	}
	readAny, err = session.readAvailableLogLines(file)
	file.Close()
	if err != nil || !readAny {
		t.Fatalf("complete record was not read: readAny=%v err=%v", readAny, err)
	}
	if session.readOffset != int64(len("partial record\n")) {
		t.Fatalf("read offset = %d, want %d", session.readOffset, len("partial record\n"))
	}
	message := <-session.OutputChan
	if message.Type != TypeAssistantMessage || string(message.Content) != `{"content":"partial record","type":"assistant_message"}` {
		t.Fatalf("emitted message = %#v", message)
	}
}
