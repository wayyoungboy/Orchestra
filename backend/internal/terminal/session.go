package terminal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessSession struct {
	ID            string
	PID           int
	Cmd           *exec.Cmd
	Workspace     string
	WorkspaceID   string
	MemberID      string
	MemberName    string
	TerminalType  string
	CreatedAt     time.Time
	LastActive    time.Time

	mu         sync.Mutex
	pty        *os.File
	OutputChan chan []byte
	ErrorChan  chan error
	DoneChan   chan struct{}
	done       bool

	chatTargetMu           sync.Mutex
	lastChatConversationID string
	chatInjectMu           sync.Mutex
	lastChatInjectedText   string
	lastChatInjectedAt     time.Time
	outputHook             func(*ProcessSession, []byte)

	streamSpanMu sync.Mutex
	streamSpanID string
	streamSeq    uint64

	chatStreamMu sync.Mutex
	chatStream   chan<- []byte // optional; set by terminal WS write loop

	// terminalScrollback retains recent PTY bytes so new WebSocket clients can replay history.
	scrollbackMu       sync.Mutex
	terminalScrollback []byte
}

// maxTerminalScrollbackBytes caps memory per session; oldest bytes are dropped.
const maxTerminalScrollbackBytes = 512 * 1024

func createSession(config ProcessConfig) (*ProcessSession, error) {
	session := &ProcessSession{
		ID:            config.ID,
		Workspace:     config.Workspace,
		WorkspaceID:   config.WorkspaceID,
		MemberID:      config.MemberID,
		MemberName:    config.MemberName,
		TerminalType:  config.TerminalType,
		CreatedAt:     time.Now(),
		LastActive:    time.Now(),
		OutputChan:    make(chan []byte, 4096),
		ErrorChan:     make(chan error, 16),
		DoneChan:      make(chan struct{}),
	}

	cmd := exec.Command(config.Command, config.Args...)
	cmd.Dir = config.Workspace
	cmd.Env = append(os.Environ(), config.Env...)

	pty, err := startPty(cmd)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	session.Cmd = cmd
	session.pty = pty
	session.PID = cmd.Process.Pid

	go session.readOutput()
	go session.waitProcess()

	return session, nil
}

// SetOutputHook registers a callback for each PTY read chunk (runs in the PTY reader goroutine).
func (s *ProcessSession) SetOutputHook(fn func(*ProcessSession, []byte)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outputHook = fn
}

// SetLastChatTargetConversation records which chat conversation should receive AI lines from this PTY.
func (s *ProcessSession) SetLastChatTargetConversation(convID string) {
	s.chatTargetMu.Lock()
	defer s.chatTargetMu.Unlock()
	s.lastChatConversationID = convID
}

// LastChatTargetConversation returns the bound conversation id for chat bridging, if any.
func (s *ProcessSession) LastChatTargetConversation() string {
	s.chatTargetMu.Lock()
	defer s.chatTargetMu.Unlock()
	return s.lastChatConversationID
}

// NoteChatInjectedLine records text forwarded from chat into this PTY (for echo suppression).
func (s *ProcessSession) NoteChatInjectedLine(text string) {
	t := strings.TrimSpace(text)
	if t == "" {
		return
	}
	s.chatInjectMu.Lock()
	defer s.chatInjectMu.Unlock()
	s.lastChatInjectedText = t
	s.lastChatInjectedAt = time.Now()
}

// ShouldSuppressAsChatEcho returns true if displayLine is likely the PTY echo of the last injected line.
func (s *ProcessSession) ShouldSuppressAsChatEcho(displayLine string) bool {
	s.chatInjectMu.Lock()
	defer s.chatInjectMu.Unlock()
	if s.lastChatInjectedText == "" {
		return false
	}
	if time.Since(s.lastChatInjectedAt) > 15*time.Second {
		return false
	}
	injected := normalizeEchoKey(s.lastChatInjectedText)
	got := normalizeEchoKey(displayLine)
	if injected == "" || got == "" {
		return false
	}
	if got == injected {
		s.lastChatInjectedText = ""
		return true
	}
	if strings.Contains(got, injected) && len(got)-len(injected) <= 8 {
		s.lastChatInjectedText = ""
		return true
	}
	if strings.Contains(injected, got) && len(injected)-len(got) <= 8 {
		s.lastChatInjectedText = ""
		return true
	}
	return false
}

func normalizeEchoKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, ">")
	s = strings.TrimSpace(s)
	// TUI often echoes without spaces (e.g. "@监工现在几点了" vs "@监工 现在几点了")
	return strings.Join(strings.Fields(s), "")
}

// SetStreamSpanID starts a new terminal↔chat output span (reference-desktop-style); resets stream seq.
func (s *ProcessSession) SetStreamSpanID(id string) {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSpanID = id
	s.streamSeq = 0
}

// StreamSpanID returns the active span for streaming/final chat payloads, if any.
func (s *ProcessSession) StreamSpanID() string {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	return s.streamSpanID
}

// NextStreamSeq returns the next monotonic seq for terminal chat stream payloads.
func (s *ProcessSession) NextStreamSeq() uint64 {
	s.streamSpanMu.Lock()
	defer s.streamSpanMu.Unlock()
	s.streamSeq++
	return s.streamSeq
}

// SetChatStreamSink receives JSON messages ({type, payload}) for the terminal WebSocket client.
func (s *ProcessSession) SetChatStreamSink(ch chan<- []byte) {
	s.chatStreamMu.Lock()
	defer s.chatStreamMu.Unlock()
	s.chatStream = ch
}

// TrySendChatStream delivers a pre-serialized JSON object to the active WebSocket, non-blocking.
func (s *ProcessSession) appendTerminalScrollback(data []byte) {
	if len(data) == 0 {
		return
	}
	s.scrollbackMu.Lock()
	defer s.scrollbackMu.Unlock()
	s.terminalScrollback = append(s.terminalScrollback, data...)
	if len(s.terminalScrollback) > maxTerminalScrollbackBytes {
		s.terminalScrollback = s.terminalScrollback[len(s.terminalScrollback)-maxTerminalScrollbackBytes:]
	}
}

// SnapshotTerminalScrollback returns a copy of buffered output for WebSocket replay.
func (s *ProcessSession) SnapshotTerminalScrollback() []byte {
	s.scrollbackMu.Lock()
	defer s.scrollbackMu.Unlock()
	out := make([]byte, len(s.terminalScrollback))
	copy(out, s.terminalScrollback)
	return out
}

func (s *ProcessSession) TrySendChatStream(data []byte) {
	s.chatStreamMu.Lock()
	ch := s.chatStream
	s.chatStreamMu.Unlock()
	if ch == nil || len(data) == 0 {
		return
	}
	select {
	case ch <- data:
	default:
	}
}

func (s *ProcessSession) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActive = time.Now()
	return s.pty.Write(data)
}

func (s *ProcessSession) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActive = time.Now()
	return resizePty(s.pty, cols, rows)
}

func (s *ProcessSession) Kill() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}
	s.done = true

	if s.Cmd != nil && s.Cmd.Process != nil {
		s.Cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-time.After(2 * time.Second):
			s.Cmd.Process.Kill()
		case <-s.DoneChan:
		}
	}

	if s.pty != nil {
		s.pty.Close()
	}

	close(s.DoneChan)
}

func (s *ProcessSession) readOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.LastActive = time.Now()
			data := make([]byte, n)
			copy(data, buf[:n])
			s.appendTerminalScrollback(data)
			select {
			case s.OutputChan <- data:
			default:
				// 实时通道满时丢弃该块（滚动缓冲里仍保留，重连可回放）
			}
			s.mu.Lock()
			hook := s.outputHook
			s.mu.Unlock()
			if hook != nil {
				// Log before calling hook
				log.Printf("[terminal] Calling output hook for session %s, dataLen=%d", s.ID, len(data))
				hook(s, data)
			} else {
				log.Printf("[terminal] No output hook set for session %s", s.ID)
			}
		}
		if err != nil {
			s.ErrorChan <- err
			return
		}
	}
}

func (s *ProcessSession) waitProcess() {
	err := s.Cmd.Wait()
	if err != nil {
		s.ErrorChan <- err
	}
	s.Kill()
}