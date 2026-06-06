package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/tmux"
)

// AgentSession manages a single agent's lifecycle: state machine, dispatch queue,
// output processing, and terminal interaction. It wraps an a2a.Session for transport
// and adds dispatch queue + state machine + stability poller on top.
type AgentSession struct {
	ID          string
	WorkspaceID string
	MemberID    string
	MemberName  string

	sm     *StateMachine
	queue  *DispatchQueue
	poller *StabilityPoller

	mu         sync.Mutex
	lastActive time.Time
	createdAt  time.Time

	// Underlying transport session
	transport *a2a.Session
	tmuxMgr   *tmux.Manager
	tmuxName  string

	// Output buffer for semantic message batching
	outputBuffer *OutputBuffer

	// External dependencies
	toolHandler *a2a.ToolHandler
	onStatus    func(workspaceID, memberID string, state AgentState)
	onOutput    func(msg *a2a.ACPMessage)
	bridge      OutputProcessor
}

// NewAgentSession creates a new session in Offline state.
func NewAgentSession(id, workspaceID, memberID, memberName string) *AgentSession {
	s := &AgentSession{
		ID:          id,
		WorkspaceID: workspaceID,
		MemberID:    memberID,
		MemberName:  memberName,
		sm:          NewStateMachine(),
		queue:       NewDispatchQueue(),
		createdAt:   time.Now(),
		lastActive:  time.Now(),
	}

	// Create output buffer (bridge will be set later via SetBridge)
	s.outputBuffer = NewOutputBuffer(s, nil)

	s.queue.SetForceFlush(func(items []DispatchItem) {
		s.dispatchItems(items)
	})

	// When transitioning to Online, flush any queued messages
	s.sm.OnEnter(StateConnecting, StateOnline, func() {
		s.flushQueue()
	})
	s.sm.OnEnter(StateWorking, StateOnline, func() {
		s.flushQueue()
	})

	return s
}

// SetTmuxManager sets the tmux manager for this session.
func (s *AgentSession) SetTmuxManager(mgr *tmux.Manager) {
	s.tmuxMgr = mgr
}

// SetToolHandler sets the tool execution handler.
func (s *AgentSession) SetToolHandler(h *a2a.ToolHandler) {
	s.toolHandler = h
}

// SetStatusCallback sets the callback for state changes.
func (s *AgentSession) SetStatusCallback(fn func(workspaceID, memberID string, state AgentState)) {
	s.onStatus = fn
}

// SetOutputCallback sets the callback for output messages (e.g., chat bridge).
func (s *AgentSession) SetOutputCallback(fn func(msg *a2a.ACPMessage)) {
	s.onOutput = fn
}

// SetBridge sets the output processor for this session.
func (s *AgentSession) SetBridge(b OutputProcessor) {
	s.bridge = b
	// Also update the output buffer's bridge
	if s.outputBuffer != nil {
		s.outputBuffer.bridge = b
	}
}

// Transport returns the underlying a2a.Session for WS/chat integration.
func (s *AgentSession) Transport() *a2a.Session {
	return s.transport
}

// SendUserMessage sends a user message directly to the agent (bypasses dispatch queue).
func (s *AgentSession) SendUserMessage(content string) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured for %s", s.ID)
	}
	return s.transport.SendUserMessage(content)
}

// AgentState returns the current agent state.
func (s *AgentSession) AgentState() AgentState {
	return s.sm.Current()
}

// IsAlive returns true if the tmux session still exists.
func (s *AgentSession) IsAlive() bool {
	if s.transport != nil {
		return s.transport.IsAlive()
	}
	if s.tmuxMgr == nil || s.tmuxName == "" {
		return false
	}
	return s.tmuxMgr.SessionExists(s.tmuxName)
}

// CaptureScrollback returns recent visible tmux pane output for inspection.
func (s *AgentSession) CaptureScrollback(ctx context.Context, lines int) (string, error) {
	if s.transport != nil {
		return s.transport.CaptureScrollback(ctx, lines)
	}
	if s.tmuxMgr == nil || s.tmuxName == "" {
		return "", fmt.Errorf("tmux session not configured")
	}
	return s.tmuxMgr.CapturePane(ctx, s.tmuxName, lines)
}

// Dispatch enqueues a message for the agent. If the agent is Online,
// the queue is flushed immediately.
func (s *AgentSession) Dispatch(content string, senderID string) error {
	s.mu.Lock()
	s.lastActive = time.Now()
	s.mu.Unlock()

	err := s.queue.Enqueue(DispatchItem{
		Content:  content,
		SenderID: senderID,
	})
	if err != nil {
		return err
	}

	// If agent is online, flush immediately
	if s.sm.Current() == StateOnline {
		s.flushQueue()
	}

	return nil
}

// flushQueue sends all queued messages to the agent.
func (s *AgentSession) flushQueue() {
	items := s.queue.Flush()
	if len(items) > 0 {
		s.dispatchItems(items)
	}
}

// dispatchItems sends items to the tmux session.
func (s *AgentSession) dispatchItems(items []DispatchItem) {
	if s.transport == nil {
		s.queue.ClearInflight()
		return
	}

	for _, item := range items {
		if err := s.transport.SendUserMessage(item.Content); err != nil {
			log.Printf("[agent-session] SendUserMessage failed for %s: %v", s.MemberName, err)
			s.queue.ClearInflight()
			return
		}

		// Transition to Working on dispatch
		if s.sm.Current() == StateOnline {
			_ = s.sm.Transition(StateWorking)
			s.onStateChange()
		}
	}
}

// Start creates the tmux session and begins output processing.
func (s *AgentSession) Start(ctx context.Context, member *models.Member, workspaceDir string) error {
	if s.tmuxMgr == nil {
		return fmt.Errorf("tmux manager not configured")
	}

	command := member.ACPCommand
	args := member.ACPArgs
	if args == nil {
		args = []string{}
	}
	command, args = buildAgentCommand(command, args)

	s.tmuxName = tmux.BuildSessionName(s.WorkspaceID, s.MemberID)

	// Transition: Offline → Connecting
	if err := s.sm.Transition(StateConnecting); err != nil {
		return err
	}
	s.onStateChange()

	// Create tmux session
	if err := s.tmuxMgr.CreateSession(ctx, s.tmuxName, workspaceDir, command, args); err != nil {
		_ = s.sm.Transition(StateOffline)
		s.onStateChange()
		return fmt.Errorf("create tmux session: %w", err)
	}

	// Create the a2a.Session transport wrapper
	sessionID := "tmux_" + s.ID
	tmuxSess := tmux.NewTmuxSession(
		sessionID,
		s.WorkspaceID,
		s.MemberID,
		s.MemberName,
		member.TerminalType,
		s.tmuxName,
		workspaceDir,
		command,
		args,
	)

	// Setup pipe-pane for output capture
	tmuxSess.LogFile = "/tmp/orch-" + s.tmuxName + ".log"
	if err := tmuxSess.SetupPipePane(ctx); err != nil {
		log.Printf("[agent-session] Failed to setup pipe-pane: %v", err)
	}

	// Start output reader
	if err := tmuxSess.StartOutputReader(ctx, false); err != nil {
		log.Printf("[agent-session] Failed to start output reader: %v", err)
	}

	s.transport = a2a.NewSession(sessionID, s.WorkspaceID, s.MemberID, s.MemberName, member.TerminalType, tmuxSess)

	// Start output buffer for semantic batching
	if s.outputBuffer != nil {
		s.outputBuffer.Start(ctx)
	}

	// Start output processing goroutine
	go s.processOutput(ctx)

	// Start post-ready automation
	steps := tmux.DefaultPostReadySteps(member.TerminalType)
	go func() {
		auto := tmux.NewPostReadyAutomation(tmuxSess, steps)
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := auto.Execute(bgCtx); err != nil {
			log.Printf("[agent-session] Post-ready automation failed for %s: %v", s.MemberName, err)
		}
	}()

	// Shell ready timeout: force Online after 3s
	go func() {
		select {
		case <-time.After(ShellReadyTimeout):
			if s.sm.Current() == StateConnecting {
				log.Printf("[agent-session] Shell ready timeout for %s, forcing Online", s.MemberName)
				_ = s.sm.Transition(StateOnline)
				s.onStateChange()
			}
		case <-s.transport.DoneChan:
			return
		}
	}()

	// Start stability poller
	s.poller = NewStabilityPoller(s, s.tmuxMgr)
	s.poller.Start(ctx)

	return nil
}

// Kill stops the session and kills the tmux process.
func (s *AgentSession) Kill() error {
	if s.outputBuffer != nil {
		s.outputBuffer.Stop()
	}
	if s.poller != nil {
		s.poller.Stop()
	}

	if s.transport != nil {
		s.transport.Kill()
	} else if s.tmuxMgr != nil && s.tmuxName != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.tmuxMgr.KillSession(ctx, s.tmuxName)
	}

	_ = s.sm.Transition(StateOffline)
	s.onStateChange()

	return nil
}

// Release stops output processing but keeps the tmux session alive.
func (s *AgentSession) Release() {
	if s.outputBuffer != nil {
		s.outputBuffer.Stop()
	}
	if s.poller != nil {
		s.poller.Stop()
	}
	if s.transport != nil {
		s.transport.Release()
	}
}

// SetLastChatTargetConversation binds a conversation for chat output routing.
func (s *AgentSession) SetLastChatTargetConversation(convID string) {
	if s.transport != nil {
		s.transport.SetLastChatTargetConversation(convID)
	}
}

// LastChatTargetConversation returns the bound conversation ID.
func (s *AgentSession) LastChatTargetConversation() string {
	if s.transport != nil {
		return s.transport.LastChatTargetConversation()
	}
	return ""
}

// GetWorkspaceID implements chatbridge.SessionInterface.
func (s *AgentSession) GetWorkspaceID() string { return s.WorkspaceID }

// GetMemberID implements chatbridge.SessionInterface.
func (s *AgentSession) GetMemberID() string { return s.MemberID }

// GetMemberName implements chatbridge.SessionInterface.
func (s *AgentSession) GetMemberName() string { return s.MemberName }

// TrySendChatStream implements chatbridge.SessionInterface.
func (s *AgentSession) TrySendChatStream(data []byte) {
	if s.transport != nil {
		s.transport.TrySendChatStream(data)
	}
}

// NextStreamSeq implements chatbridge.SessionInterface.
func (s *AgentSession) NextStreamSeq() uint64 {
	if s.transport != nil {
		return s.transport.NextStreamSeq()
	}
	return 0
}

// StreamSpanID implements chatbridge.SessionInterface.
func (s *AgentSession) StreamSpanID() string {
	if s.transport != nil {
		return s.transport.StreamSpanID()
	}
	return ""
}

// onStateChange fires the status callback.
func (s *AgentSession) onStateChange() {
	if s.onStatus != nil {
		s.onStatus(s.WorkspaceID, s.MemberID, s.sm.Current())
	}
}

// processOutput reads from the transport session's OutputChan, adds state
// transitions and poller notifications, then forwards to chat bridge.
func (s *AgentSession) processOutput(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.transport.DoneChan:
			// Transport died → Offline
			if s.sm.Current() != StateOffline {
				_ = s.sm.Transition(StateOffline)
				s.onStateChange()
			}
			return
		case msg, ok := <-s.transport.OutputChan:
			if !ok {
				return
			}
			s.handleACPMessage(msg)
		case err := <-s.transport.ErrorChan:
			if err != nil {
				log.Printf("[agent-session] Error from %s: %v", s.MemberName, err)
			}
		}
	}
}

// handleACPMessage processes an ACP message from the agent output.
func (s *AgentSession) handleACPMessage(msg *a2a.ACPMessage) {
	// Notify poller of activity
	if s.poller != nil {
		s.poller.NotifyOutput()
	}

	// Fire output callback (e.g., chat bridge)
	if s.onOutput != nil {
		s.onOutput(msg)
	}

	switch msg.Type {
	case a2a.TypeSystem:
		// System init → shell ready
		parsed, err := msg.ParseErrorMessage()
		if err == nil && parsed != nil {
			if s.sm.Current() == StateConnecting {
				_ = s.sm.Transition(StateOnline)
				s.onStateChange()
			}
		}

	case a2a.TypeAssistantMessage:
		// Agent output → transition to Working
		if s.sm.Current() == StateOnline {
			_ = s.sm.Transition(StateWorking)
			s.onStateChange()
		}
		s.sm.SetToolInFlight(false)
		// Push to output buffer for semantic batching
		if s.outputBuffer != nil {
			s.outputBuffer.Push(msg)
		} else if s.bridge != nil {
			s.bridge.OnMessage(s, msg)
		}

	case a2a.TypeToolUse:
		// Flush any pending assistant message buffer before tool_use
		if s.outputBuffer != nil {
			s.outputBuffer.Flush()
		}
		s.sm.SetToolInFlight(true)
		// Execute tool if handler is configured
		if s.toolHandler != nil {
			go s.executeTool(msg)
		}

	case a2a.TypeResult:
		// Flush any pending assistant message buffer before result
		if s.outputBuffer != nil {
			s.outputBuffer.Flush()
		}
		s.sm.SetToolInFlight(false)
		// Bridge completion notification
		if s.bridge != nil {
			s.bridge.OnMessage(s, msg)
		}

	case a2a.TypeError:
		// Error from agent — bridge for logging
		if s.bridge != nil {
			s.bridge.OnMessage(s, msg)
		}
	}
}

// executeTool runs a tool and sends the result back to the agent.
func (s *AgentSession) executeTool(msg *a2a.ACPMessage) {
	result := s.toolHandler.ExecuteTool(msg, s.transport)
	if result == nil {
		return
	}
	toolUseParsed, _ := msg.ParseToolUseMessage()
	if toolUseParsed == nil {
		return
	}
	if err := s.transport.SendToolResultToAgent(toolUseParsed.ToolUseID, result.Content, result.IsError); err != nil {
		log.Printf("[agent-session] Failed to send tool result: %v", err)
	}
}

// buildAgentCommand augments the base command with headless flags.
func buildAgentCommand(command string, args []string) (string, []string) {
	base := args
	switch command {
	case "claude":
		hasOutputFormat := false
		hasInputFormat := false
		hasSkipPerms := false
		hasVerbose := false
		for _, a := range base {
			switch a {
			case "--output-format":
				hasOutputFormat = true
			case "--input-format":
				hasInputFormat = true
			case "--dangerously-skip-permissions":
				hasSkipPerms = true
			case "--verbose":
				hasVerbose = true
			}
		}
		if !hasOutputFormat {
			base = append([]string{"--output-format", "stream-json"}, base...)
		}
		if !hasInputFormat {
			base = append([]string{"--input-format", "stream-json"}, base...)
		}
		if !hasSkipPerms {
			base = append([]string{"--dangerously-skip-permissions"}, base...)
		}
		if !hasVerbose {
			base = append([]string{"--verbose"}, base...)
		}
	}
	return command, base
}
