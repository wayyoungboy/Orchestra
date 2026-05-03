package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/orchestra/backend/internal/tmux"
)

// StabilityPoller detects when an agent in Working state has gone silent
// and transitions it back to Online.
type StabilityPoller struct {
	session  *AgentSession
	mgr      *tmux.Manager
	lastSnap string

	mu      sync.Mutex
	silent  bool
	since   time.Time // when silence started
	cancel  context.CancelFunc
	stopped chan struct{}
}

// NewStabilityPoller creates a poller for the given session.
func NewStabilityPoller(session *AgentSession, mgr *tmux.Manager) *StabilityPoller {
	return &StabilityPoller{
		session: session,
		mgr:     mgr,
		stopped: make(chan struct{}),
	}
}

// Start begins the polling loop. It stops when the context is cancelled
// or Stop() is called.
func (p *StabilityPoller) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	go p.run(ctx)
}

// Stop halts the poller.
func (p *StabilityPoller) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	<-p.stopped
}

// NotifyOutput resets the silence timer. Call this whenever the session
// receives output from the agent.
func (p *StabilityPoller) NotifyOutput() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.silent = false
	p.since = time.Time{}
}

func (p *StabilityPoller) run(ctx context.Context) {
	defer close(p.stopped)

	ticker := time.NewTicker(StatusPollInterval)
	defer ticker.Stop()

	var debounceTimer *time.Timer
	var debounceC <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case <-ticker.C:
			if p.session.sm.Current() != StateWorking {
				if debounceTimer != nil {
					debounceTimer.Stop()
					debounceTimer = nil
					debounceC = nil
				}
				continue
			}

			snap, err := p.mgr.CapturePane(ctx, p.session.tmuxName, 1)
			if err != nil {
				continue
			}

			p.mu.Lock()
			isSilent := (snap == p.lastSnap)
			p.mu.Unlock()

			if isSilent {
				p.mu.Lock()
				if !p.silent {
					p.silent = true
					p.since = time.Now()
				}
				elapsed := time.Since(p.since)
				p.mu.Unlock()

				timeout := SilenceTimeout
				if p.session.sm.ToolInFlight() {
					timeout = SilenceTimeoutBusy
				}

				if elapsed >= timeout && debounceTimer == nil {
					debounceTimer = time.NewTimer(StabilizeDebounce)
					debounceC = debounceTimer.C
				}
			} else {
				p.NotifyOutput()
				if debounceTimer != nil {
					debounceTimer.Stop()
					debounceTimer = nil
					debounceC = nil
				}
			}

			p.mu.Lock()
			p.lastSnap = snap
			p.mu.Unlock()

		case <-debounceC:
			debounceTimer = nil
			debounceC = nil

			if p.session.sm.Current() != StateWorking {
				continue
			}

			// Re-check: still silent after debounce?
			p.mu.Lock()
			stillSilent := p.silent
			p.mu.Unlock()

			if !stillSilent {
				continue
			}

			log.Printf("[poller] Agent %s: Working→Online (silence confirmed)", p.session.MemberName)
			if err := p.session.sm.Transition(StateOnline); err != nil {
				log.Printf("[poller] Transition error: %v", err)
			}
			p.session.onStateChange()

			// Flush any queued messages
			p.session.flushQueue()
		}
	}
}
