package agent

import (
	"context"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/a2a"
)

// OutputBuffer batches assistant message fragments within a time window
// and flushes them as single coherent messages.
//
// This addresses fragmented assistant messages that arrive as individual
// lines from the agent. The buffer collects fragments within ChatSilenceTimeout
// and flushes them as one merged message when silence is detected.
type OutputBuffer struct {
	mu        sync.Mutex
	fragments []string
	lastMsg   time.Time
	timer     *time.Timer
	session   SessionView
	bridge    OutputProcessor
	acpMeta   *a2a.ACPMessage // store last raw msg for reconstruction
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
}

// NewOutputBuffer creates a new output buffer for a session.
func NewOutputBuffer(session SessionView, bridge OutputProcessor) *OutputBuffer {
	return &OutputBuffer{
		session:   session,
		bridge:    bridge,
		fragments: make([]string, 0, 16),
	}
}

// Start begins the background silence detection goroutine.
func (b *OutputBuffer) Start(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return
	}

	b.ctx, b.cancel = context.WithCancel(ctx)
	b.running = true
	b.lastMsg = time.Now()

	// Start the debounce timer
	b.timer = time.AfterFunc(ChatIdleDebounce, b.checkSilence)
}

// Stop stops the buffer and flushes any pending content.
func (b *OutputBuffer) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	b.running = false
	if b.timer != nil {
		b.timer.Stop()
	}
	if b.cancel != nil {
		b.cancel()
	}

	// Flush any remaining fragments
	b.flushLocked()
}

// Push adds an assistant message fragment to the buffer.
// If the message is not an assistant_message, it passes through directly.
func (b *OutputBuffer) Push(msg *a2a.ACPMessage) {
	if msg == nil {
		return
	}

	// Only buffer assistant_message types
	if msg.Type != a2a.TypeAssistantMessage {
		// Pass through directly
		if b.bridge != nil {
			b.bridge.OnMessage(b.session, msg)
		}
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Extract content from the message
	parsed, err := msg.ParseAssistantMessage()
	if err != nil || parsed == nil {
		// Can't parse, pass through
		if b.bridge != nil {
			b.bridge.OnMessage(b.session, msg)
		}
		return
	}

	// Add fragment to buffer
	b.fragments = append(b.fragments, parsed.Content)
	b.lastMsg = time.Now()
	b.acpMeta = msg

	// Reset the silence detection timer
	if b.timer != nil {
		b.timer.Reset(ChatIdleDebounce)
	}
}

// Flush immediately flushes any buffered content.
// Call this before sending result or tool_use messages.
func (b *OutputBuffer) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

// flushLocked flushes the buffer. Must be called with mu held.
func (b *OutputBuffer) flushLocked() {
	if len(b.fragments) == 0 {
		return
	}
	if b.bridge == nil {
		b.fragments = b.fragments[:0]
		return
	}

	// Merge all fragments
	merged := b.mergeFragments()

	// Create a new ACPMessage with merged content
	mergedMsg := &a2a.ACPMessage{
		Type:    a2a.TypeAssistantMessage,
		Content: merged,
	}

	// Deliver to bridge
	b.bridge.OnMessage(b.session, mergedMsg)

	// Clear buffer
	b.fragments = b.fragments[:0]
	b.acpMeta = nil
}

// mergeFragments joins all buffered fragments.
// Must be called with mu held.
func (b *OutputBuffer) mergeFragments() []byte {
	if len(b.fragments) == 0 {
		return nil
	}

	// Simple join with newlines
	// Build the JSON structure for assistant_message
	result := make([]byte, 0, 256)
	result = append(result, `{"type":"assistant_message","content":"`...)

	// Escape and concatenate content
	merged := ""
	for i, f := range b.fragments {
		if i > 0 {
			merged += "\n"
		}
		merged += f
	}

	// JSON-escape the content
	escaped := escapeJSONString(merged)
	result = append(result, escaped...)
	result = append(result, '"', '}')

	return result
}

// escapeJSONString escapes special characters for JSON strings.
func escapeJSONString(s string) string {
	result := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			result = append(result, '\\', '"')
		case '\\':
			result = append(result, '\\', '\\')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		default:
			if c < 0x20 {
				// Control character, skip
				continue
			}
			result = append(result, c)
		}
	}
	return string(result)
}

// checkSilence is called by the timer to check if silence timeout has elapsed.
func (b *OutputBuffer) checkSilence() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	// Check if silence timeout has elapsed since last message
	elapsed := time.Since(b.lastMsg)
	if elapsed >= ChatSilenceTimeout {
		// Silence detected, flush
		b.flushLocked()
	}

	// Reschedule the check
	if b.timer != nil && b.running {
		b.timer.Reset(ChatIdleDebounce)
	}
}

// IsEmpty returns true if there are no buffered fragments.
func (b *OutputBuffer) IsEmpty() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.fragments) == 0
}

// BufferedCount returns the number of buffered fragments.
func (b *OutputBuffer) BufferedCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.fragments)
}
