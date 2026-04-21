package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestSubscribeAndEmit(t *testing.T) {
	eb := New()
	var received Event
	eb.Subscribe(EventProcessStateChanged, func(e Event) {
		received = e
	})

	eb.EmitPayload(EventProcessStateChanged, map[string]any{"state": "idle"})

	if received.Type != EventProcessStateChanged {
		t.Errorf("expected event type %q, got %q", EventProcessStateChanged, received.Type)
	}
	if received.Payload["state"] != "idle" {
		t.Errorf("expected state %q, got %q", "idle", received.Payload["state"])
	}
}

func TestUnsubscribe(t *testing.T) {
	eb := New()
	callCount := 0
	handler := func(e Event) {
		callCount++
	}

	unsub := eb.Subscribe(EventProcessStateChanged, handler)
	eb.EmitPayload(EventProcessStateChanged, nil)
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}

	unsub()
	eb.EmitPayload(EventProcessStateChanged, nil)
	if callCount != 1 {
		t.Fatalf("expected 1 call after unsubscribe, got %d", callCount)
	}
}

func TestUnsubscribeViaReturnedFunc(t *testing.T) {
	eb := New()
	callCount := 0
	handler := func(e Event) {
		callCount++
	}

	unsub := eb.Subscribe(EventProcessStateChanged, handler)
	eb.EmitPayload(EventProcessStateChanged, nil)
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}

	unsub()
	eb.EmitPayload(EventProcessStateChanged, nil)
	if callCount != 1 {
		t.Fatalf("expected 1 call after unsub, got %d", callCount)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	eb := New()
	var mu sync.Mutex
	results := []string{}

	eb.Subscribe(EventProcessStateChanged, func(e Event) {
		mu.Lock()
		results = append(results, "a")
		mu.Unlock()
	})
	eb.Subscribe(EventProcessStateChanged, func(e Event) {
		mu.Lock()
		results = append(results, "b")
		mu.Unlock()
	})

	eb.EmitPayload(EventProcessStateChanged, nil)

	mu.Lock()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	mu.Unlock()
}

func TestDifferentEventTypes(t *testing.T) {
	eb := New()
	called := false
	eb.Subscribe(EventFileChanged, func(e Event) {
		called = true
	})

	// Emit a different event type — should not trigger FileChanged handler
	eb.EmitPayload(EventProcessStateChanged, nil)
	if called {
		t.Error("handler should not have been called for different event type")
	}
}

func TestEmitWithNoSubscribers(t *testing.T) {
	eb := New()
	// Should not panic
	eb.EmitPayload(EventProcessStateChanged, nil)
}

func TestHandlerCount(t *testing.T) {
	eb := New()
	if eb.HandlerCount(EventProcessStateChanged) != 0 {
		t.Error("expected 0 handlers")
	}

	eb.Subscribe(EventProcessStateChanged, func(e Event) {})
	if eb.HandlerCount(EventProcessStateChanged) != 1 {
		t.Error("expected 1 handler")
	}
}

func TestPanicInHandlerDoesNotCrash(t *testing.T) {
	eb := New()
	eb.Subscribe(EventProcessStateChanged, func(e Event) {
		panic("test panic")
	})

	// Should not panic
	eb.EmitPayload(EventProcessStateChanged, nil)
}

func TestConcurrentSubscribeEmit(t *testing.T) {
	eb := New()
	var mu sync.Mutex
	count := 0

	// Concurrent subscribers
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Subscribe(EventProcessStateChanged, func(e Event) {
				mu.Lock()
				count++
				mu.Unlock()
			})
		}()
	}
	wg.Wait()

	// Emit from multiple goroutines
	var emitWg sync.WaitGroup
	for i := 0; i < 10; i++ {
		emitWg.Add(1)
		go func() {
			defer emitWg.Done()
			eb.EmitPayload(EventProcessStateChanged, nil)
		}()
	}
	emitWg.Wait()

	mu.Lock()
	// 10 subscribers * 10 emits = 100 calls
	if count != 100 {
		t.Errorf("expected 100 calls, got %d", count)
	}
	mu.Unlock()
}

func TestEventTimestamp(t *testing.T) {
	eb := New()
	before := time.Now()
	var ts time.Time

	eb.Subscribe(EventProcessStateChanged, func(e Event) {
		ts = e.Timestamp
	})
	eb.EmitPayload(EventProcessStateChanged, nil)

	after := time.Now()

	if ts.Before(before) || ts.After(after) {
		t.Error("event timestamp should be between before and after emit")
	}
}
