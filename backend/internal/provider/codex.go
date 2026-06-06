package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// CodexProvider implements AgentProvider for OpenAI Codex CLI.
type CodexProvider struct {
	command string
	args    []string
}

// NewCodexProvider creates a Codex CLI provider.
// If command is empty, defaults to "codex".
func NewCodexProvider(command string, args ...string) *CodexProvider {
	if command == "" {
		command = "codex"
	}
	return &CodexProvider{command: command, args: args}
}

func (p *CodexProvider) Name() ProviderName {
	return ProviderCodex
}

func (p *CodexProvider) DisplayName() string {
	return "OpenAI Codex"
}

func (p *CodexProvider) IsInstalled() bool {
	_, err := exec.LookPath(p.command)
	return err == nil
}

func (p *CodexProvider) SupportsPermissionMode() bool {
	return false
}

func (p *CodexProvider) StartSession(ctx context.Context, opts SessionOptions) (AgentSession, error) {
	if !p.IsInstalled() {
		return nil, fmt.Errorf("codex is not installed")
	}

	cmd := exec.CommandContext(ctx, p.command, p.args...)
	if opts.WorkspacePath != "" {
		cmd.Dir = opts.WorkspacePath
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("CODEX_WORKSPACE=%s", opts.WorkspacePath))

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start codex: %w", err)
	}

	return &codexSession{
		cmd:   cmd,
		pipe:  pipe,
		msgCh: make(chan AgentMessage, 64),
		done:  make(chan struct{}),
	}, nil
}

type codexSession struct {
	cmd   *exec.Cmd
	pipe  io.WriteCloser
	msgCh chan AgentMessage
	done  chan struct{}
}

func (s *codexSession) Messages() <-chan AgentMessage {
	return s.msgCh
}

func (s *codexSession) Send(msg UserMessage) error {
	_, err := s.pipe.Write([]byte(msg.Text + "\n"))
	return err
}

func (s *codexSession) Abort() error {
	if s.pipe != nil {
		s.pipe.Close()
	}
	return s.cmd.Process.Kill()
}

func (s *codexSession) IsAlive() bool {
	if s.cmd.Process == nil {
		return false
	}
	return s.cmd.Process.Signal(os.Signal(nil)) == nil
}

func (s *codexSession) PID() int {
	if s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}
