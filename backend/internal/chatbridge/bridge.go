package chatbridge

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/danielgatis/go-vte"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
)

var debugLog *log.Logger

func init() {
	f, err := os.OpenFile("/tmp/orchestra-chatbridge.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		debugLog = log.New(f, "", log.LstdFlags)
	} else {
		debugLog = log.New(os.Stderr, "[chatbridge] ", log.LstdFlags)
	}
}

// Stream cadence
const streamEmitMinInterval = 500 * time.Millisecond

// streamPayload for WebSocket events
type streamPayload struct {
	TerminalID     string `json:"terminalId"`
	MemberID       string `json:"memberId,omitempty"`
	WorkspaceID    string `json:"workspaceId,omitempty"`
	ConversationID string `json:"conversationId,omitempty"`
	Seq            uint64 `json:"seq"`
	Timestamp      int64  `json:"timestamp"`
	Content        string `json:"content"`
	Type           string `json:"type"`
	Source         string `json:"source"`
	Mode           string `json:"mode"`
	SpanID         string `json:"spanId,omitempty"`
	MessageID      string `json:"messageId,omitempty"`
}

type streamEnvelope struct {
	Type    string        `json:"type"`
	Payload streamPayload `json:"payload"`
}

// semanticState holds per-session semantic processing state
type semanticState struct {
	screen         *ScreenBuffer
	parser         *vte.Parser
	lastExtract    time.Time
	lastContent    string
	lastInputLines []string // Track what user sent for matching
	pendingContent strings.Builder
}

// Bridge persists PTY output as chat messages using semantic extraction
type Bridge struct {
	mu        sync.Mutex
	msg       *repository.MessageRepository
	semantic  map[string]*semanticState
	streamSeq map[string]uint64
}

func New(msg *repository.MessageRepository) *Bridge {
	return &Bridge{
		msg:       msg,
		semantic:  make(map[string]*semanticState),
		streamSeq: make(map[string]uint64),
	}
}

// OnTerminalOutput implements terminal output hook
func (b *Bridge) OnTerminalOutput(s *terminal.ProcessSession, data []byte) {
	convID := s.LastChatTargetConversation()
	memberID := s.MemberID

	debugLog.Printf("OnTerminalOutput: session=%s, convID=%s, memberID=%s, dataLen=%d", s.ID, convID, memberID, len(data))

	if convID == "" || memberID == "" || len(data) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	sid := s.ID

	// Get or create semantic state
	state, ok := b.semantic[sid]
	if !ok {
		state = &semanticState{
			screen: NewScreenBuffer(120, 40),
		}
		state.parser = vte.NewParser(state.screen)
		b.semantic[sid] = state
	}

	// Feed data to parser
	for _, b := range data {
		state.parser.Advance(b)
	}

	// Periodically extract semantic content
	now := time.Now()
	if now.Sub(state.lastExtract) >= streamEmitMinInterval {
		b.extractAndEmitLocked(s, state, convID)
		state.lastExtract = now
	}
}

// NoteUserInput records what user sent for matching in output
func (b *Bridge) NoteUserInput(sessionID string, input string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	state, ok := b.semantic[sessionID]
	if !ok {
		return
	}

	// Store input lines for matching
	lines := strings.Split(input, "\n")
	var nonEmpty []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmpty = append(nonEmpty, trimmed)
		}
	}
	state.lastInputLines = nonEmpty
	state.pendingContent.Reset()
}

// extractAndEmitLocked extracts AI response content from screen
func (b *Bridge) extractAndEmitLocked(s *terminal.ProcessSession, state *semanticState, convID string) {
	content := state.screen.GetVisibleContent()
	if content == "" {
		return
	}

	// Debug: log the screen content
	debugLog.Printf("Screen content for session %s:\n%s", s.ID, content)

	// Extract AI response using prompt/bullet pattern
	response := extractAIResponse(content, state.lastInputLines)
	if response == "" {
		return
	}

	// Skip if we already sent this content
	if response == state.lastContent {
		return
	}
	state.lastContent = response

	debugLog.Printf("Extracted response: %s", response)

	// Create chat message
	memberID := s.MemberID
	msg, err := b.msg.Create(repository.MessageCreate{
		ConversationID: convID,
		SenderID:       memberID,
		Content:        repository.MessageContent{Type: "text", Text: response},
		IsAI:           false,
	})
	if err != nil || msg == nil {
		debugLog.Printf("Failed to create message: %v", err)
		return
	}

	// Emit stream event
	seq := b.streamSeq[s.ID] + 1
	b.streamSeq[s.ID] = seq

	env := streamEnvelope{
		Type: "terminal_chat_stream",
		Payload: streamPayload{
			TerminalID:     s.ID,
			MemberID:       s.MemberID,
			WorkspaceID:    s.WorkspaceID,
			ConversationID: convID,
			Seq:            seq,
			Timestamp:      msg.CreatedAt,
			Content:        response,
			Type:           "info",
			Source:         "pty",
			Mode:           "final",
			MessageID:      msg.ID,
		},
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return
	}
	s.TrySendChatStream(raw)
}

// extractAIResponse extracts AI response content using Claude Code's prompt/bullet pattern
func extractAIResponse(screenContent string, inputLines []string) string {
	lines := strings.Split(screenContent, "\n")

	// Find prompt markers (›) and bullet markers (•, ✦)
	var promptIndices []int
	var bulletIndices []int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isPromptLine(trimmed) {
			promptIndices = append(promptIndices, i)
		}
		if isBulletLine(trimmed) {
			bulletIndices = append(bulletIndices, i)
		}
	}

	// If no bullets found, no AI response yet
	if len(bulletIndices) == 0 {
		return ""
	}

	// Find the prompt that matches user input (if available)
	var startPrompt int
	if len(inputLines) > 0 && len(promptIndices) > 0 {
		for _, idx := range promptIndices {
			if idx+1 < len(lines) && matchesInputSegment(lines, idx, inputLines) {
				startPrompt = idx
				break
			}
		}
	} else if len(promptIndices) > 0 {
		// Use last prompt as fallback
		startPrompt = promptIndices[len(promptIndices)-1]
	}

	// Find the first bullet after start prompt
	var startBullet int = -1
	for _, idx := range bulletIndices {
		if idx > startPrompt {
			startBullet = idx
			break
		}
	}

	if startBullet < 0 {
		return ""
	}

	// Find end of response (next prompt or end of screen)
	var endIdx int = len(lines)
	for _, idx := range promptIndices {
		if idx > startBullet {
			endIdx = idx
			break
		}
	}

	// Extract response lines
	var response []string
	for i := startBullet; i < endIdx && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Skip TUI chrome
		if isTUIChrome(line) {
			continue
		}
		// Remove bullet prefix
		line = stripBulletPrefix(line)
		if line != "" {
			response = append(response, line)
		}
	}

	if len(response) == 0 {
		return ""
	}

	return strings.Join(response, "\n")
}

func isPromptLine(line string) bool {
	return strings.HasPrefix(line, "›") || strings.HasPrefix(line, ">")
}

func isBulletLine(line string) bool {
	return strings.HasPrefix(line, "•") ||
		strings.HasPrefix(line, "✦") ||
		strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "* ")
}

func stripBulletPrefix(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "• ") {
		return strings.TrimPrefix(line, "• ")
	}
	if strings.HasPrefix(line, "✦ ") {
		return strings.TrimPrefix(line, "✦ ")
	}
	if strings.HasPrefix(line, "- ") {
		return strings.TrimPrefix(line, "- ")
	}
	if strings.HasPrefix(line, "* ") {
		return strings.TrimPrefix(line, "* ")
	}
	return line
}

func isTUIChrome(line string) bool {
	// Skip status bar lines
	if strings.Contains(line, "Claude Code") && strings.Contains(line, "v2.") {
		return true
	}
	if strings.Contains(line, "glm-") || strings.Contains(line, "API Usage") {
		return true
	}
	if strings.HasPrefix(line, "▗") || strings.HasPrefix(line, "▘") {
		return true
	}
	// Skip separator lines
	if strings.Count(line, "─") > len(line)/2 {
		return true
	}
	// Skip path indicators
	if strings.HasPrefix(line, "~/") || strings.HasPrefix(line, "/") {
		return true
	}
	return false
}

func matchesInputSegment(lines []string, promptIdx int, inputLines []string) bool {
	if promptIdx+1 >= len(lines) {
		return false
	}

	// Check if the lines after prompt match input
	matchCount := 0
	for i, input := range inputLines {
		if promptIdx+1+i >= len(lines) {
			break
		}
		screenLine := strings.TrimSpace(lines[promptIdx+1+i])
		if strings.Contains(screenLine, input) || strings.Contains(input, screenLine) {
			matchCount++
		}
	}

	return matchCount > 0
}

// CleanupSession removes semantic state for a terminated session
func (b *Bridge) CleanupSession(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.semantic, sessionID)
	delete(b.streamSeq, sessionID)
}