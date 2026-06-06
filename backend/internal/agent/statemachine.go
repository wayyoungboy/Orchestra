package agent

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// AgentState represents the lifecycle state of an agent session.
type AgentState string

const (
	StateOffline    AgentState = "offline"
	StateConnecting AgentState = "connecting"
	StateOnline     AgentState = "online"
	StateWorking    AgentState = "working"
)

// validTransitions defines the allowed state transition graph.
var validTransitions = map[AgentState][]AgentState{
	StateOffline:    {StateConnecting},
	StateConnecting: {StateOnline, StateOffline},
	StateOnline:     {StateWorking, StateOffline},
	StateWorking:    {StateOnline, StateOffline},
}

// TransitionKey identifies a state transition for handler registration.
type TransitionKey struct {
	From AgentState
	To   AgentState
}

// StateMachine manages agent session state transitions with validation and hooks.
type StateMachine struct {
	mu       sync.Mutex
	current  AgentState
	handlers map[TransitionKey][]func()

	// toolInFlight extends SilenceTimeout to SilenceTimeoutBusy when true.
	toolInFlight atomic.Bool
}

// NewStateMachine creates a state machine starting in StateOffline.
func NewStateMachine() *StateMachine {
	return &StateMachine{
		current:  StateOffline,
		handlers: make(map[TransitionKey][]func()),
	}
}

// Current returns the current state.
func (sm *StateMachine) Current() AgentState {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.current
}

// Transition attempts to move to a new state. Returns an error if the
// transition is not allowed. On success, any registered OnEnter handlers
// for the (old→new) transition are called synchronously.
func (sm *StateMachine) Transition(to AgentState) error {
	sm.mu.Lock()

	if to == sm.current {
		sm.mu.Unlock()
		return nil // no-op
	}

	allowed := validTransitions[sm.current]
	valid := false
	for _, s := range allowed {
		if s == to {
			valid = true
			break
		}
	}
	if !valid {
		current := sm.current
		sm.mu.Unlock()
		return fmt.Errorf("invalid state transition: %s → %s", current, to)
	}

	from := sm.current
	sm.current = to

	key := TransitionKey{From: from, To: to}
	handlers := append([]func(){}, sm.handlers[key]...)
	sm.mu.Unlock()

	for _, fn := range handlers {
		fn()
	}

	return nil
}

// OnEnter registers a callback for a specific state transition.
func (sm *StateMachine) OnEnter(from, to AgentState, fn func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	key := TransitionKey{From: from, To: to}
	sm.handlers[key] = append(sm.handlers[key], fn)
}

// SetToolInFlight sets the tool-in-flight flag. When true, the poller
// uses SilenceTimeoutBusy instead of SilenceTimeout.
func (sm *StateMachine) SetToolInFlight(v bool) {
	sm.toolInFlight.Store(v)
}

// ToolInFlight returns whether a tool is currently being executed.
func (sm *StateMachine) ToolInFlight() bool {
	return sm.toolInFlight.Load()
}
