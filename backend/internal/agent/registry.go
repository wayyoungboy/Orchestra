package agent

import (
	"context"
	"log"
	"sync"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/tmux"
	"github.com/orchestra/backend/pkg/utils"
)

// Registry manages agent sessions with indexed lookups.
// It replaces the a2a.Pool's unindexed map traversal.
type Registry struct {
	mu             sync.RWMutex
	sessions       map[string]*AgentSession            // sessionID → session
	memberIndex    map[string]map[string]*AgentSession // workspaceID → memberID → session
	workspaceIndex map[string][]string                 // workspaceID → sessionIDs

	tmuxMgr     *tmux.Manager
	toolHandler *a2a.ToolHandler
	outputFn    func(msg *a2a.ACPMessage)
	bridge      OutputProcessor
}

// OutputProcessor handles agent output messages (e.g., bridging to chat).
type OutputProcessor interface {
	OnMessage(sess SessionView, msg *a2a.ACPMessage)
}

// SessionView is the session interface the bridge needs.
type SessionView interface {
	LastChatTargetConversation() string
	GetWorkspaceID() string
	GetMemberID() string
	GetMemberName() string
	TrySendChatStream([]byte)
	NextStreamSeq() uint64
	StreamSpanID() string
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		sessions:       make(map[string]*AgentSession),
		memberIndex:    make(map[string]map[string]*AgentSession),
		workspaceIndex: make(map[string][]string),
	}
}

// SetTmuxManager sets the tmux manager for session creation.
func (r *Registry) SetTmuxManager(mgr *tmux.Manager) {
	r.tmuxMgr = mgr
}

// SetToolHandler sets the tool handler for all current and future sessions.
func (r *Registry) SetToolHandler(h *a2a.ToolHandler) {
	r.mu.Lock()
	r.toolHandler = h
	r.mu.Unlock()

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		s.SetToolHandler(h)
	}
}

// SetBridge sets the output processor for all current and future sessions.
func (r *Registry) SetBridge(b OutputProcessor) {
	r.mu.Lock()
	r.bridge = b
	r.mu.Unlock()

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		s.SetBridge(b)
	}
}
func (r *Registry) SetOutputCallback(fn func(msg *a2a.ACPMessage)) {
	r.mu.Lock()
	r.outputFn = fn
	r.mu.Unlock()

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		s.SetOutputCallback(fn)
	}
}

// Register adds a session to all indexes.
func (r *Registry) Register(s *AgentSession) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[s.ID] = s

	if r.memberIndex[s.WorkspaceID] == nil {
		r.memberIndex[s.WorkspaceID] = make(map[string]*AgentSession)
	}
	r.memberIndex[s.WorkspaceID][s.MemberID] = s

	r.workspaceIndex[s.WorkspaceID] = append(r.workspaceIndex[s.WorkspaceID], s.ID)

	// Apply current tool handler, output callback, and bridge
	if r.toolHandler != nil {
		s.SetToolHandler(r.toolHandler)
	}
	if r.outputFn != nil {
		s.SetOutputCallback(r.outputFn)
	}
	if r.bridge != nil {
		s.SetBridge(r.bridge)
	}
}

// Unregister removes a session from all indexes.
func (r *Registry) Unregister(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.sessions[sessionID]
	if !ok {
		return
	}

	delete(r.sessions, sessionID)

	if wsMap, ok := r.memberIndex[s.WorkspaceID]; ok {
		delete(wsMap, s.MemberID)
		if len(wsMap) == 0 {
			delete(r.memberIndex, s.WorkspaceID)
		}
	}

	wsList := r.workspaceIndex[s.WorkspaceID]
	for i, id := range wsList {
		if id == sessionID {
			r.workspaceIndex[s.WorkspaceID] = append(wsList[:i], wsList[i+1:]...)
			break
		}
	}
	if len(r.workspaceIndex[s.WorkspaceID]) == 0 {
		delete(r.workspaceIndex, s.WorkspaceID)
	}
}

// GetByMember returns the session for a workspace member, or nil.
func (r *Registry) GetByMember(workspaceID, memberID string) *AgentSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if wsMap, ok := r.memberIndex[workspaceID]; ok {
		return wsMap[memberID]
	}
	return nil
}

// GetByID returns a session by its session ID.
func (r *Registry) GetByID(sessionID string) *AgentSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sessions[sessionID]
}

// ListByWorkspace returns all sessions for a workspace.
func (r *Registry) ListByWorkspace(workspaceID string) []*AgentSession {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.workspaceIndex[workspaceID]
	result := make([]*AgentSession, 0, len(ids))
	for _, id := range ids {
		if s, ok := r.sessions[id]; ok {
			result = append(result, s)
		}
	}
	return result
}

// AcquireOrCreate gets an existing session for a member, or creates a new one.
func (r *Registry) AcquireOrCreate(ctx context.Context, member *models.Member, workspaceDir string) (*AgentSession, error) {
	r.mu.Lock()
	// Check if session already exists
	if wsMap, ok := r.memberIndex[member.WorkspaceID]; ok {
		if sess, ok := wsMap[member.ID]; ok {
			if sess.IsAlive() {
				r.mu.Unlock()
				return sess, nil
			}
			// Dead session, remove it
			delete(r.sessions, sess.ID)
			delete(wsMap, member.ID)
			log.Printf("[registry] Removed dead session %s for member %s", sess.ID, member.Name)
		}
	}
	r.mu.Unlock()

	if !member.ACPEnabled || member.ACPCommand == "" {
		return nil, nil
	}

	// Create new AgentSession
	sessionID := "tmux_" + utils.GenerateID()
	sess := NewAgentSession(sessionID, member.WorkspaceID, member.ID, member.Name)
	sess.SetTmuxManager(r.tmuxMgr)

	if err := sess.Start(ctx, member, workspaceDir); err != nil {
		return nil, err
	}

	r.Register(sess)
	log.Printf("[registry] Created session %s for member %s", sessionID, member.Name)
	return sess, nil
}

// Release removes a session by ID.
func (r *Registry) Release(sessionID string) {
	r.mu.RLock()
	sess := r.sessions[sessionID]
	r.mu.RUnlock()

	if sess != nil {
		sess.Release()
		r.Unregister(sessionID)
	}
}

// ReleaseByTransport finds a session by its transport ID and releases it.
func (r *Registry) ReleaseByTransport(transportID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, s := range r.sessions {
		if s.Transport() != nil && s.Transport().ID == transportID {
			s.Release()
			r.mu.RUnlock()
			r.Unregister(id)
			r.mu.RLock()
			return
		}
	}
}

// RecoverFromTmux scans for existing orchestra tmux sessions and
// reconstructs AgentSession objects for them.
func (r *Registry) RecoverFromTmux(ctx context.Context, mgr *tmux.Manager) error {
	names, err := mgr.ListOrchestraSessions(ctx)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	log.Printf("[registry] Recovering %d tmux session(s)", len(names))

	for _, tmuxName := range names {
		wsID, memberID, ok := tmux.ParseSessionName(tmuxName)
		if !ok {
			log.Printf("[registry] Skipping unknown session name: %s", tmuxName)
			continue
		}

		sessionID := "recovered_" + tmuxName
		sess := &AgentSession{
			ID:          sessionID,
			WorkspaceID: wsID,
			MemberID:    memberID,
			MemberName:  "recovered",
			tmuxName:    tmuxName,
			tmuxMgr:     mgr,
			sm:          NewStateMachine(),
			queue:       NewDispatchQueue(),
		}
		sess.sm.Transition(StateOnline) // recovered sessions are assumed online

		r.Register(sess)
		log.Printf("[registry] Recovered session %s (tmux: %s)", sessionID, tmuxName)
	}

	return nil
}

// SessionForWorkspaceMember implements a2a.SessionLookup by returning
// the a2a.Session transport for the given workspace member.
func (r *Registry) SessionForWorkspaceMember(workspaceID, memberID string) *a2a.Session {
	sess := r.GetByMember(workspaceID, memberID)
	if sess == nil {
		return nil
	}
	return sess.Transport()
}

// Acquire creates a new a2a.Session for the given config, wrapping it in an
// AgentSession via AcquireOrCreate.
func (r *Registry) Acquire(ctx context.Context, config a2a.SessionConfig) (*a2a.Session, error) {
	if config.Member == nil {
		return nil, nil
	}
	sess, err := r.AcquireOrCreate(ctx, config.Member, config.WorkspaceDir)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	return sess.Transport(), nil
}

// ListSessionsForWorkspace returns session info for a workspace.
func (r *Registry) ListSessionsForWorkspace(workspaceID string) []WorkspaceSessionInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := r.workspaceIndex[workspaceID]
	infos := make([]WorkspaceSessionInfo, 0, len(sessions))
	for _, id := range sessions {
		if s, ok := r.sessions[id]; ok {
			infos = append(infos, WorkspaceSessionInfo{
				MemberID:  s.MemberID,
				SessionID: s.ID,
			})
		}
	}
	return infos
}

// WorkspaceSessionInfo provides session metadata for the frontend.
type WorkspaceSessionInfo struct {
	MemberID  string `json:"memberId"`
	SessionID string `json:"sessionId"`
}
