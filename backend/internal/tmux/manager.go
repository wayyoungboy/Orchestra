// Package tmux provides tmux session management for Orchestra agent processes.
package tmux

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Manager wraps tmux CLI commands for session lifecycle management.
type Manager struct {
	mu         sync.Mutex
	socketPath string
	prefix     string // session name prefix, default "orchestra-"
}

// NewManager creates a tmux manager.
func NewManager(socketPath string) *Manager {
	m := &Manager{
		socketPath: socketPath,
		prefix:     "orchestra-",
	}
	return m
}

// SetPrefix changes the session name prefix (default "orchestra-").
func (m *Manager) SetPrefix(prefix string) {
	m.prefix = prefix
}

// CreateSession creates a new detached tmux session running the given command.
func (m *Manager) CreateSession(ctx context.Context, name, cwd, command string, args []string) error {
	if m.SessionExists(name) {
		return fmt.Errorf("tmux session %s already exists", name)
	}

	allArgs := []string{"new-session", "-d", "-s", name}
	if cwd != "" {
		allArgs = append(allArgs, "-c", cwd)
	}
	allArgs = append(allArgs, command)
	allArgs = append(allArgs, args...)

	_, err := m.exec(ctx, allArgs...)
	return err
}

// KillSession terminates a tmux session.
func (m *Manager) KillSession(ctx context.Context, name string) error {
	_, err := m.exec(ctx, "kill-session", "-t", name)
	if err != nil && strings.Contains(err.Error(), "can't find session") {
		return nil // already gone
	}
	return err
}

// ListSessions returns all tmux session names.
func (m *Manager) ListSessions(ctx context.Context) ([]string, error) {
	out, err := m.exec(ctx, "list-sessions", "-F", "#{session_name}")
	if err != nil {
		// tmux returns exit code 1 when no sessions exist
		if strings.Contains(err.Error(), "no server running") || strings.Contains(err.Error(), "no sessions") {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names, nil
}

// ListOrchestraSessions returns only Orchestra-managed tmux sessions.
func (m *Manager) ListOrchestraSessions(ctx context.Context) ([]string, error) {
	all, err := m.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []string
	for _, name := range all {
		if strings.HasPrefix(name, m.prefix) {
			filtered = append(filtered, name)
		}
	}
	return filtered, nil
}

// SendKeys sends text to a tmux session's active pane.
// Uses -l (literal) mode to avoid key binding interference.
func (m *Manager) SendKeys(ctx context.Context, name, text string) error {
	_, err := m.exec(ctx, "send-keys", "-t", name, "-l", text)
	return err
}

// SendEnter sends an Enter key to the session.
func (m *Manager) SendEnter(ctx context.Context, name string) error {
	_, err := m.exec(ctx, "send-keys", "-t", name, "Enter")
	return err
}

// CapturePane captures visible output from the tmux pane.
func (m *Manager) CapturePane(ctx context.Context, name string, lines int) (string, error) {
	start := fmt.Sprintf("-%d", lines)
	return m.exec(ctx, "capture-pane", "-t", name, "-p", "-S", start)
}

// CapturePaneRaw captures pane output with ANSI escape codes included.
func (m *Manager) CapturePaneRaw(ctx context.Context, name string, lines int) (string, error) {
	start := fmt.Sprintf("-%d", lines)
	return m.exec(ctx, "capture-pane", "-t", name, "-p", "-e", "-S", start)
}

// SessionExists checks if a tmux session exists.
func (m *Manager) SessionExists(name string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := m.exec(ctx, "has-session", "-t", name)
	return err == nil
}

// GetPID returns the PID of the process in the tmux pane.
func (m *Manager) GetPID(ctx context.Context, name string) (int, error) {
	out, err := m.exec(ctx, "display-message", "-t", name, "-p", "#{pane_pid}")
	if err != nil {
		return 0, err
	}
	var pid int
	_, err = fmt.Sscanf(strings.TrimSpace(out), "%d", &pid)
	return pid, err
}

// SetupPipePane configures pipe-pane to append output to a log file.
// The -o flag ensures pipe-pane only activates if not already piped.
func (m *Manager) SetupPipePane(ctx context.Context, name, logFile string) error {
	cmd := fmt.Sprintf("cat >> %s", logFile)
	_, err := m.exec(ctx, "pipe-pane", "-t", name, "-o", cmd)
	return err
}

// exec runs a tmux command with optional socket path.
func (m *Manager) exec(ctx context.Context, args ...string) (string, error) {
	m.mu.Lock()
	baseArgs := []string{"tmux"}
	if m.socketPath != "" {
		baseArgs = append(baseArgs, "-S", m.socketPath)
	}
	baseArgs = append(baseArgs, args...)
	m.mu.Unlock()

	cmd := exec.CommandContext(ctx, baseArgs[0], baseArgs[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tmux %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// BuildSessionName creates a tmux session name from Orchestra IDs.
func BuildSessionName(workspaceID, memberID string) string {
	// Shorten IDs to avoid tmux name length limits
	shortWS := workspaceID
	if len(workspaceID) > 8 {
		shortWS = workspaceID[:8]
	}
	shortM := memberID
	if len(memberID) > 8 {
		shortM = memberID[:8]
	}
	return fmt.Sprintf("orchestra-%s-%s", shortWS, shortM)
}

// ParseSessionName extracts workspaceID and memberID from a tmux session name.
// Returns ok=false if the name doesn't match the expected format.
func ParseSessionName(name string) (workspaceID, memberID string, ok bool) {
	prefix := "orchestra-"
	if !strings.HasPrefix(name, prefix) {
		return "", "", false
	}
	rest := name[len(prefix):]
	parts := strings.SplitN(rest, "-", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// IsOrchestraSession checks if a tmux session belongs to Orchestra.
func IsOrchestraSession(name string) bool {
	return strings.HasPrefix(name, "orchestra-")
}
