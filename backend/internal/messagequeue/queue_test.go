package messagequeue

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPushAndRead(t *testing.T) {
	q := New[string](10)
	q.Push("hello")
	q.Push("world")

	msg1 := <-q.Messages()
	if msg1 != "hello" {
		t.Errorf("expected %q, got %q", "hello", msg1)
	}

	msg2 := <-q.Messages()
	if msg2 != "world" {
		t.Errorf("expected %q, got %q", "world", msg2)
	}
}

func TestFIFOOrder(t *testing.T) {
	q := New[int](100)
	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	for i := 0; i < 10; i++ {
		got := <-q.Messages()
		if got != i {
			t.Errorf("expected %d, got %d", i, got)
		}
	}
}

func TestLen(t *testing.T) {
	q := New[string](10)
	if q.Len() != 0 {
		t.Errorf("expected 0, got %d", q.Len())
	}

	q.Push("a")
	q.Push("b")
	if q.Len() != 2 {
		t.Errorf("expected 2, got %d", q.Len())
	}
}

func TestCapacity(t *testing.T) {
	q := New[int](42)
	if q.Capacity() != 42 {
		t.Errorf("expected capacity 42, got %d", q.Capacity())
	}
}

func TestPushNonBlocking(t *testing.T) {
	q := New[string](1)
	if !q.PushNonBlocking("a") {
		t.Error("expected push to succeed")
	}

	if q.PushNonBlocking("b") {
		t.Error("expected push to fail on full queue")
	}
}

func TestClose(t *testing.T) {
	q := New[string](10)
	q.Push("before close")
	q.Close()

	// Should be able to read remaining messages
	msg := <-q.Messages()
	if msg != "before close" {
		t.Errorf("expected %q, got %q", "before close", msg)
	}

	// Channel should be closed
	_, ok := <-q.Messages()
	if ok {
		t.Error("expected channel to be closed")
	}
}

func TestDrain(t *testing.T) {
	q := New[string](10)
	q.Push("a")
	q.Push("b")
	q.Push("c")

	count := q.Drain()
	if count != 3 {
		t.Errorf("expected to drain 3, got %d", count)
	}
	if q.Len() != 0 {
		t.Errorf("expected empty queue, got %d", q.Len())
	}
}

func TestWaitOrContextWithMessage(t *testing.T) {
	q := New[string](10)
	q.Push("ready")

	ctx := context.Background()
	msg, err := q.WaitOrContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "ready" {
		t.Errorf("expected %q, got %q", "ready", msg)
	}
}

func TestWaitOrContextWithCancellation(t *testing.T) {
	q := New[string](10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_, err := q.WaitOrContext(ctx)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancellation")
	}
}

func TestConcurrentPushPop(t *testing.T) {
	q := New[int](1000)
	const n = 1000

	var wg sync.WaitGroup
	// Pushers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < n/10; j++ {
				q.Push(start*100 + j)
			}
		}(i)
	}

	// Consumer
	received := make(map[int]bool)
	var mu sync.Mutex
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for msg := range q.Messages() {
			mu.Lock()
			received[msg] = true
			mu.Unlock()
			count++
			if count >= n {
				return
			}
		}
	}()

	wg.Wait()

	mu.Lock()
	if len(received) != n {
		t.Errorf("expected %d unique messages, got %d", n, len(received))
	}
	mu.Unlock()
}

func TestQueueWithType(t *testing.T) {
	type Msg struct {
		Type    string
		Content string
	}

	q := New[Msg](10)
	q.Push(Msg{Type: "text", Content: "hello"})

	msg := <-q.Messages()
	if msg.Type != "text" || msg.Content != "hello" {
		t.Errorf("unexpected message: %+v", msg)
	}
}
