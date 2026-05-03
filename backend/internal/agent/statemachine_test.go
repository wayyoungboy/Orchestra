package agent

import (
	"sync/atomic"
	"testing"
)

func TestNewStateMachine_StartsOffline(t *testing.T) {
	sm := NewStateMachine()
	if sm.Current() != StateOffline {
		t.Errorf("expected StateOffline, got %s", sm.Current())
	}
}

func TestTransition_ValidOfflineToConnecting(t *testing.T) {
	sm := NewStateMachine()
	err := sm.Transition(StateConnecting)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Current() != StateConnecting {
		t.Errorf("expected StateConnecting, got %s", sm.Current())
	}
}

func TestTransition_InvalidOfflineToWorking(t *testing.T) {
	sm := NewStateMachine()
	err := sm.Transition(StateWorking)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if sm.Current() != StateOffline {
		t.Errorf("state should not change, got %s", sm.Current())
	}
}

func TestTransition_WorkingToOnline(t *testing.T) {
	sm := NewStateMachine()
	sm.Transition(StateConnecting)
	sm.Transition(StateOnline)
	sm.Transition(StateWorking)

	err := sm.Transition(StateOnline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Current() != StateOnline {
		t.Errorf("expected StateOnline, got %s", sm.Current())
	}
}

func TestTransition_SameStateNoop(t *testing.T) {
	sm := NewStateMachine()
	sm.Transition(StateConnecting)
	err := sm.Transition(StateConnecting)
	if err != nil {
		t.Fatalf("same-state transition should be no-op, got: %v", err)
	}
}

func TestTransition_ProcessExitToOffline(t *testing.T) {
	sm := NewStateMachine()
	sm.Transition(StateConnecting)
	sm.Transition(StateOnline)
	sm.Transition(StateWorking)

	err := sm.Transition(StateOffline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Current() != StateOffline {
		t.Errorf("expected StateOffline, got %s", sm.Current())
	}
}

func TestOnEnter_HandlerCalled(t *testing.T) {
	sm := NewStateMachine()
	var called atomic.Int32

	sm.OnEnter(StateWorking, StateOnline, func() {
		called.Add(1)
	})

	sm.Transition(StateConnecting)
	sm.Transition(StateOnline)
	sm.Transition(StateWorking)
	sm.Transition(StateOnline)

	if called.Load() != 1 {
		t.Errorf("expected handler called once, got %d", called.Load())
	}
}

func TestToolInFlight(t *testing.T) {
	sm := NewStateMachine()
	if sm.ToolInFlight() {
		t.Error("expected false initially")
	}
	sm.SetToolInFlight(true)
	if !sm.ToolInFlight() {
		t.Error("expected true after set")
	}
	sm.SetToolInFlight(false)
	if sm.ToolInFlight() {
		t.Error("expected false after clear")
	}
}

func TestFullLifecycle(t *testing.T) {
	sm := NewStateMachine()

	// Offline → Connecting → Online → Working → Online → Offline
	steps := []AgentState{StateConnecting, StateOnline, StateWorking, StateOnline, StateOffline}
	for _, target := range steps {
		if err := sm.Transition(target); err != nil {
			t.Fatalf("transition to %s failed: %v", target, err)
		}
	}
}
