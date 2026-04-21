package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ClaudeProvider implements AgentProvider for Claude Code.
type ClaudeProvider struct {
	command string
	args    []string
}

// NewClaudeProvider creates a Claude Code provider.
// If command is empty, defaults to "claude".
func NewClaudeProvider(command string, args ...string) *ClaudeProvider {
	if command == "" {
		command = "claude"
	}
	return &ClaudeProvider{command: command, args: args}
}

func (p *ClaudeProvider) Name() ProviderName {
	return ProviderClaude
}

func (p *ClaudeProvider) DisplayName() string {
	return "Claude Code"
}

func (p *ClaudeProvider) IsInstalled() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}

func (p *ClaudeProvider) SupportsPermissionMode() bool {
	return true
}

func (p *ClaudeProvider) StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error) {
	if !p.IsInstalled() {
		return nil, fmt.Errorf("claude is not installed")
	}

	cmdArgs := append([]string{}, p.args...)
	if opts.WorkspacePath != "" {
		cmdArgs = append(cmdArgs, "--cwd", opts.WorkspacePath)
	}
	if opts.PermissionMode != "" {
		cmdArgs = append(cmdArgs, "--permission-mode", opts.PermissionMode)
	}

	cmd := exec.CommandContext(ctx, p.command, cmdArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CLAUDE_CODE_WORKSPACE=%s", opts.WorkspacePath))

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	return &claudeSession{
		cmd:  cmd,
		pipe: pipe,
		msgCh: make(chan AgentMessage, 64),
		done:  make(chan struct{}),
	}, nil
}

type claudeSession struct {
	cmd   *exec.Cmd
	pipe  io.WriteCloser
	msgCh chan AgentMessage
	done  chan struct{}
}

func (s *claudeSession) Messages() <-chan AgentMessage {
	return s.msgCh
}

func (s *claudeSession) Send(msg UserMessage) error {
	// Write user message to Claude's stdin in ACP format
	_, err := s.pipe.Write([]byte(msg.Text + "\n"))
	return err
}

func (s *claudeSession) Abort() error {
	if s.pipe != nil {
		s.pipe.Close()
	}
	return s.cmd.Process.Kill()
}

func (s *claudeSession) IsAlive() bool {
	if s.cmd.Process == nil {
		return false
	}
	return s.cmd.Process.Signal(os.Signal(nil)) == nil
}

func (s *claudeSession) PID() int {
	if s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}
