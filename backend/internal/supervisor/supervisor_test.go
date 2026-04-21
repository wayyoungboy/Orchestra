package supervisor

import (
	"testing"
	"time"

	"github.com/orchestra/backend/internal/eventbus"
	"github.com/orchestra/backend/internal/provider"
)

func TestNewSupervisor(t *testing.T) {
	eb := eventbus.New()
	reg := provider.NewRegistry()
	cfg := DefaultConfig()
	cfg.MaxWorkers = 3

	s := New(cfg, eb, reg)
	if s == nil {
		t.Fatal("expected non-nil supervisor")
	}
	if s.ActiveSessionCount() != 0 {
		t.Error("expected 0 active sessions")
	}
	if s.IsAtCapacity() {
		t.Error("should not be at capacity with 0 sessions and maxWorkers=3")
	}
}

func TestActiveSessions(t *testing.T) {
	eb := eventbus.New()
	reg := provider.NewRegistry()
	s := New(DefaultConfig(), eb, reg)

	ids := s.ActiveSessions()
	if len(ids) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(ids))
	}
}

func TestGetSessionNotFound(t *testing.T) {
	eb := eventbus.New()
	reg := provider.NewRegistry()
	s := New(DefaultConfig(), eb, reg)

	_, ok := s.GetSession("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestTerminateNonExistentSession(t *testing.T) {
	eb := eventbus.New()
	reg := provider.NewRegistry()
	s := New(DefaultConfig(), eb, reg)

	err := s.TerminateSession("nonexistent")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestIsAtCapacity_ZeroMaxWorkers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWorkers = 0 // unlimited
	s := New(cfg, nil, nil)

	if s.IsAtCapacity() {
		t.Error("should never be at capacity with maxWorkers=0")
	}
}

func TestIsAtCapacity_AtLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxWorkers = 2
	s := New(cfg, nil, nil)

	// Manually inject sessions for testing
	s.mu.Lock()
	s.sessions["sess-1"] = &runningSession{}
	s.sessions["sess-2"] = &runningSession{}
	s.mu.Unlock()

	if !s.IsAtCapacity() {
		t.Error("expected to be at capacity with 2/2 sessions")
	}
}

func TestCheckStaleSessions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StaleThreshold = 100 * time.Millisecond
	s := New(cfg, nil, nil)

	// Inject a stale dead session
	s.mu.Lock()
	s.sessions["stale-session"] = &runningSession{
		info: SessionInfo{
			SessionID: "stale-session",
			State:     StateInTurn,
		},
		lastMsgTime: time.Now().Add(-time.Hour),
		cancel:      func() {},
	}
	s.mu.Unlock()

	s.CheckStaleSessions()

	_, ok := s.GetSession("stale-session")
	if ok {
		t.Error("expected stale session to be removed")
	}
}

func TestCheckStaleSessions_SkipsNonStale(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StaleThreshold = time.Hour
	s := New(cfg, nil, nil)

	s.mu.Lock()
	s.sessions["fresh-session"] = &runningSession{
		info: SessionInfo{
			SessionID: "fresh-session",
			State:     StateInTurn,
		},
		lastMsgTime: time.Now(),
		cancel:      func() {},
	}
	s.mu.Unlock()

	s.CheckStaleSessions()

	_, ok := s.GetSession("fresh-session")
	if !ok {
		t.Error("expected fresh session to remain")
	}
}

func TestCheckStaleSessions_SkipsWaitingInput(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StaleThreshold = 100 * time.Millisecond
	s := New(cfg, nil, nil)

	s.mu.Lock()
	s.sessions["waiting-session"] = &runningSession{
		info: SessionInfo{
			SessionID: "waiting-session",
			State:     StateWaitingInput,
		},
		lastMsgTime: time.Now().Add(-time.Hour),
		cancel:      func() {},
	}
	s.mu.Unlock()

	s.CheckStaleSessions()

	_, ok := s.GetSession("waiting-session")
	if !ok {
		t.Error("expected waiting-input session to remain even if stale")
	}
}

func TestEverOwnedSessions(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, nil, nil)

	s.mu.Lock()
	s.everOwnedSessions["old-session"] = true
	s.mu.Unlock()

	s.mu.RLock()
	_, ok := s.everOwnedSessions["old-session"]
	s.mu.RUnlock()
	if !ok {
		t.Error("expected old session to be tracked")
	}
}
