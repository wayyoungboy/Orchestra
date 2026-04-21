package tmux

import (
	"context"
	"log"
	"strings"
	"time"
)

// PostReadyStep defines a step to execute after agent becomes ready.
type PostReadyStep struct {
	Type    string        // "send", "wait_pattern", "mark_ready"
	Content string        // text to send, or pattern to wait for
	Timeout time.Duration
}

// PostReadyAutomation executes startup steps after agent becomes ready.
type PostReadyAutomation struct {
	session *TmuxSession
	steps   []PostReadyStep
	current int
	done    bool
}

// DefaultPostReadySteps returns standard startup steps for an agent type.
func DefaultPostReadySteps(terminalType string) []PostReadyStep {
	switch terminalType {
	case "claude":
		// Claude sends system/init automatically; we just need to wait for it
		return []PostReadyStep{
			{Type: "wait_pattern", Content: "session_id", Timeout: 15 * time.Second},
			{Type: "mark_ready"},
		}
	case "gemini":
		return []PostReadyStep{
			{Type: "send", Content: "/status", Timeout: 5 * time.Second},
			{Type: "wait_pattern", Content: "ready", Timeout: 10 * time.Second},
			{Type: "mark_ready"},
		}
	default:
		// Generic: just wait for any output then mark ready
		return []PostReadyStep{
			{Type: "wait_pattern", Content: "", Timeout: 10 * time.Second}, // any output
			{Type: "mark_ready"},
		}
	}
}

// NewPostReadyAutomation creates a post-ready automation for the given session.
func NewPostReadyAutomation(session *TmuxSession, steps []PostReadyStep) *PostReadyAutomation {
	return &PostReadyAutomation{
		session: session,
		steps:   steps,
	}
}

// Execute runs all post-ready steps sequentially.
func (p *PostReadyAutomation) Execute(ctx context.Context) error {
	for i, step := range p.steps {
		if p.done {
			return nil
		}
		p.current = i

		switch step.Type {
		case "send":
			if err := p.session.SendInput(step.Content); err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)

		case "wait_pattern":
			if err := p.waitForPattern(ctx, step.Content, step.Timeout); err != nil {
				log.Printf("[post-ready] Pattern %q not found for %s (may be ok)", step.Content, p.session.MemberName)
			}

		case "mark_ready":
			p.session.SetState(StateOnline)
			p.done = true
			log.Printf("[post-ready] Agent %s is ready", p.session.MemberName)
		}
	}
	p.done = true
	return nil
}

// waitForPattern waits until the session output contains the given pattern.
// Empty pattern means wait for any output.
func (p *PostReadyAutomation) waitForPattern(ctx context.Context, pattern string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if shell is already ready
		p.session.mu.Lock()
		ready := p.session.shellReady
		p.session.mu.Unlock()
		if ready && pattern == "" {
			return nil
		}

		// If we have a specific pattern, check recent output
		if pattern != "" {
			mgr := NewManager("")
			output, err := mgr.CapturePane(ctx, p.session.Name, 50)
			if err == nil && strings.Contains(output, pattern) {
				p.session.mu.Lock()
				p.session.shellReady = true
				p.session.mu.Unlock()
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return context.DeadlineExceeded
}
