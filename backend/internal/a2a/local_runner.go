package a2a

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// LocalRunner manages a local CLI agent subprocess with stdio communication.
// Inspired by cc-connect's ACP agent pattern.
// Supports both Claude's native stream-json format and generic JSON-RPC.
type LocalRunner struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	nextID   atomic.Int64
	pendingMu sync.Mutex
	pending   map[string]chan *json.RawMessage

	onOutput func(string)       // callback for agent output lines
	onError  func(error)        // callback for errors
	onEvent  func(string, any)  // callback for structured events (type, data)

	mu     sync.Mutex
	done   bool
	doneCh chan struct{}
}

// NewLocalRunner creates a new local CLI agent runner.
func NewLocalRunner(command string, args []string, workspacePath string) *LocalRunner {
	cmd := exec.Command(command, args...)
	if workspacePath != "" {
		cmd.Dir = workspacePath
	}
	cmd.Env = append(cmd.Env, "HOME="+getHomeDir())

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	return &LocalRunner{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		pending: make(map[string]chan *json.RawMessage),
		doneCh:  make(chan struct{}),
	}
}

// Start launches the subprocess and begins the read loop.
func (r *LocalRunner) Start(ctx context.Context) error {
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent %s: %w", r.cmd.Path, err)
	}

	log.Printf("[local-runner] Agent started: PID=%d", r.cmd.Process.Pid)

	// Start stderr reader
	go r.readStderr()

	// Start stdout read loop with background context so it survives request cancellation
	go r.readLoop(context.Background())

	return nil
}

// SendUserMessage sends a user message to the agent.
// Tries Claude's native stream-json format first, falls back to JSON-RPC.
func (r *LocalRunner) SendUserMessage(text string) error {
	r.mu.Lock()
	if r.done {
		r.mu.Unlock()
		return fmt.Errorf("runner is closed")
	}
	r.mu.Unlock()

	// Send in Claude's native stream-json format
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": text,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if _, err := r.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write to agent stdin: %w", err)
	}

	// No synchronous response expected - Claude streams responses
	return nil
}

// readLoop reads newline-delimited JSON from stdout.
func (r *LocalRunner) readLoop(ctx context.Context) {
	scanner := bufio.NewScanner(r.stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB buffer

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil && err != io.EOF {
				log.Printf("[local-runner] Read error: %v", err)
			}
			break
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		r.handleJSONLine(line)
	}

	r.close()
}

// handleJSONLine processes a single JSON line from the agent.
func (r *LocalRunner) handleJSONLine(line []byte) {
	var raw map[string]any
	if err := json.Unmarshal(line, &raw); err != nil {
		// Not JSON, treat as raw output
		if r.onOutput != nil {
			r.onOutput(string(line))
		}
		return
	}

	typ, _ := raw["type"].(string)
	if typ == "" {
		// Check for JSON-RPC format (fallback for non-Claude agents)
		if _, hasJSONRPC := raw["jsonrpc"]; hasJSONRPC {
			r.handleJSONRPC(raw)
		}
		return
	}

	// Dispatch by Claude's native type field
	switch typ {
	case "system":
		// System initialization message with session_id
		sessionID, _ := raw["session_id"].(string)
		if sessionID != "" {
			log.Printf("[local-runner] Claude session initialized: %s", sessionID)
		}
		if r.onEvent != nil {
			r.onEvent("system", raw)
		}

	case "assistant":
		// Assistant message with text content and tool_use blocks
		if r.onEvent != nil {
			r.onEvent("assistant", raw)
		}
		// Also extract text for backward compatibility
		if text := extractAssistantText(raw); text != "" && r.onOutput != nil {
			r.onOutput(text)
		}

	case "user":
		// Tool result echoed back
		if r.onEvent != nil {
			r.onEvent("tool_result", raw)
		}

	case "result":
		// Final result with usage info - handled via onEvent, don't also emit raw output
		if r.onEvent != nil {
			r.onEvent("result", raw)
		}

	case "control_request":
		// Claude requesting permission for tool use
		if r.onEvent != nil {
			r.onEvent("control_request", raw)
		}
		// Auto-approve tool use
		r.autoApproveTool(raw)

	case "control_cancel_request":
		// Permission request cancelled
		log.Printf("[local-runner] Control request cancelled")

	default:
		// Unknown type, pass as raw output
		if r.onOutput != nil {
			r.onOutput(string(line))
		}
	}
}

// handleJSONRPC handles JSON-RPC 2.0 responses (for non-Claude agents).
func (r *LocalRunner) handleJSONRPC(raw map[string]any) {
	id, hasID := raw["id"].(string)
	if !hasID {
		// Float64 ID
		if idF, ok := raw["id"].(float64); ok {
			id = fmt.Sprintf("%.0f", idF)
			hasID = true
		}
	}

	if hasID {
		r.pendingMu.Lock()
		ch, ok := r.pending[id]
		if ok {
			delete(r.pending, id)
			data := json.RawMessage{}
			_ = json.Unmarshal([]byte("{}"), &data) // placeholder
			select {
			case ch <- &data:
			default:
			}
		}
		r.pendingMu.Unlock()
	}

	// Check for method (notification)
	if method, ok := raw["method"].(string); ok {
		if r.onEvent != nil {
			r.onEvent(method, raw)
		}
	}
}

// extractAssistantText extracts text content from an assistant message.
func extractAssistantText(raw map[string]any) string {
	message, ok := raw["message"].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := message["content"].([]any)
	if !ok {
		// Might be a string directly
		if text, ok := message["content"].(string); ok {
			return text
		}
		return ""
	}

	var parts []string
	for _, block := range content {
		if blockMap, ok := block.(map[string]any); ok {
			switch blockMap["type"] {
			case "text":
				if text, ok := blockMap["text"].(string); ok {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

// autoApproveTool responds to Claude's tool permission requests.
// In Orchestra we auto-approve all tool use by design.
func (r *LocalRunner) autoApproveTool(raw map[string]any) {
	requestID, _ := raw["request_id"].(string)
	if requestID == "" {
		return
	}

	resp := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype": "success",
			"request_id": requestID,
			"response": map[string]any{
				"behavior": "allow",
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[local-runner] Failed to marshal auto-approve: %v", err)
		return
	}
	data = append(data, '\n')

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return
	}
	if _, err := r.stdin.Write(data); err != nil {
		log.Printf("[local-runner] Failed to write auto-approve: %v", err)
	}
}

// readStderr reads stderr for error messages.
func (r *LocalRunner) readStderr() {
	scanner := bufio.NewScanner(r.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			log.Printf("[local-runner] stderr: %s", line)
			if r.onError != nil {
				r.onError(fmt.Errorf("agent stderr: %s", line))
			}
		}
	}
}

// close cleans up the runner.
func (r *LocalRunner) close() {
	r.mu.Lock()
	if r.done {
		r.mu.Unlock()
		return
	}
	r.done = true
	r.mu.Unlock()

	// Cancel all pending requests
	r.pendingMu.Lock()
	for _, ch := range r.pending {
		select {
		case ch <- nil:
		default:
		}
	}
	r.pending = make(map[string]chan *json.RawMessage)
	r.pendingMu.Unlock()

	close(r.doneCh)
}

// Stop terminates the subprocess.
func (r *LocalRunner) Stop() {
	r.close()

	if r.stdin != nil {
		r.stdin.Close()
	}

	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
		r.cmd.Wait()
	}

	log.Printf("[local-runner] Agent stopped")
}

// Done returns a channel that closes when the runner stops.
func (r *LocalRunner) Done() <-chan struct{} {
	return r.doneCh
}

// PID returns the process ID.
func (r *LocalRunner) PID() int {
	if r.cmd != nil && r.cmd.Process != nil {
		return r.cmd.Process.Pid
	}
	return 0
}

func getHomeDir() string {
	if home, err := exec.Command("sh", "-c", "echo $HOME").Output(); err == nil && len(home) > 0 {
		return string(home[:len(home)-1])
	}
	return "/root"
}

// extractAssistantTextFromRaw extracts text content from a Claude assistant message.
// Exported version of extractAssistantText for use by pool.go.
func extractAssistantTextFromRaw(raw any) string {
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	return extractAssistantText(rawMap)
}
