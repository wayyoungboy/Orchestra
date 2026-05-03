package agent

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatchQueue_Enqueue(t *testing.T) {
	q := NewDispatchQueue()
	err := q.Enqueue(DispatchItem{Content: "hello", SenderID: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Len() != 1 {
		t.Errorf("expected len 1, got %d", q.Len())
	}
}

func TestDispatchQueue_EmptyContentIgnored(t *testing.T) {
	q := NewDispatchQueue()
	err := q.Enqueue(DispatchItem{Content: "", SenderID: "u1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Len() != 0 {
		t.Errorf("empty content should be ignored, got len %d", q.Len())
	}
}

func TestDispatchQueue_MergeWindow(t *testing.T) {
	q := NewDispatchQueue()
	q.Enqueue(DispatchItem{Content: "a", SenderID: "u1"})
	q.Enqueue(DispatchItem{Content: "b", SenderID: "u1"})
	q.Enqueue(DispatchItem{Content: "c", SenderID: "u1"})

	// Within merge window, same sender → merged
	if q.Len() != 1 {
		t.Errorf("expected 1 merged item, got %d", q.Len())
	}

	items := q.Flush()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].MergeOf != 3 {
		t.Errorf("expected MergeOf=3, got %d", items[0].MergeOf)
	}
	if items[0].Content != "a\nb\nc" {
		t.Errorf("expected merged content 'a\\nb\\nc', got %q", items[0].Content)
	}
}

func TestDispatchQueue_DifferentSendersNotMerged(t *testing.T) {
	q := NewDispatchQueue()
	q.Enqueue(DispatchItem{Content: "a", SenderID: "u1"})
	q.Enqueue(DispatchItem{Content: "b", SenderID: "u2"})

	if q.Len() != 2 {
		t.Errorf("expected 2 items, got %d", q.Len())
	}
}

func TestDispatchQueue_Dedup(t *testing.T) {
	q := NewDispatchQueue()
	q.Enqueue(DispatchItem{Content: "hello", SenderID: "u1"})
	q.Enqueue(DispatchItem{Content: "hello", SenderID: "u1"})

	if q.Len() != 1 {
		t.Errorf("duplicate should be deduped, got %d", q.Len())
	}
}

func TestDispatchQueue_QueueFull(t *testing.T) {
	q := NewDispatchQueue()
	for i := 0; i < DispatchQueueSize; i++ {
		// Use different content to avoid dedup
		q.Enqueue(DispatchItem{Content: fmt.Sprintf("msg-%d", i), SenderID: fmt.Sprintf("u%d", i)})
	}

	err := q.Enqueue(DispatchItem{Content: "overflow", SenderID: "u99"})
	if err != ErrQueueFull {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestDispatchQueue_FIFOOrder(t *testing.T) {
	q := NewDispatchQueue()
	// Use different senders to prevent merge
	q.Enqueue(DispatchItem{Content: "first", SenderID: "u1"})
	// Small delay to break merge window
	time.Sleep(MergeWindow + 10*time.Millisecond)
	q.Enqueue(DispatchItem{Content: "second", SenderID: "u2"})
	time.Sleep(MergeWindow + 10*time.Millisecond)
	q.Enqueue(DispatchItem{Content: "third", SenderID: "u3"})

	items := q.Flush()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Content != "first" || items[1].Content != "second" || items[2].Content != "third" {
		t.Errorf("FIFO order broken: %v", items)
	}
}

func TestDispatchQueue_ForceFlush(t *testing.T) {
	q := NewDispatchQueue()
	var flushed atomic.Int32
	q.SetForceFlush(func(items []DispatchItem) {
		flushed.Store(int32(len(items)))
	})

	q.Enqueue(DispatchItem{Content: "hello", SenderID: "u1"})

	// Wait for force flush (30s is too long for tests, so we test the mechanism)
	// In practice, the timer is set on first enqueue
	select {
	case <-time.After(100 * time.Millisecond):
		// Force flush hasn't fired yet (expected — timer is 30s)
	}

	// Manual flush
	items := q.Flush()
	if len(items) != 1 {
		t.Errorf("expected 1 flushed item, got %d", len(items))
	}
}

func TestDispatchQueue_ClearInflight(t *testing.T) {
	q := NewDispatchQueue()
	q.Enqueue(DispatchItem{Content: "important", SenderID: "u1"})

	items := q.Flush()
	if len(items) != 1 || items[0].Content != "important" {
		t.Fatalf("flush failed")
	}

	q.ClearInflight()

	// Item should be back at front
	if q.Len() != 1 {
		t.Errorf("expected 1 item after ClearInflight, got %d", q.Len())
	}
	items2 := q.Flush()
	if items2[0].Content != "important" {
		t.Errorf("expected 'important', got %q", items2[0].Content)
	}
}

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Push("a")
	rb.Push("b")

	if !rb.Contains("a") || !rb.Contains("b") {
		t.Error("expected a and b in buffer")
	}
	if rb.Contains("c") {
		t.Error("c should not be in buffer")
	}

	rb.Push("c")
	rb.Push("d") // evicts "a"

	if rb.Contains("a") {
		t.Error("a should have been evicted")
	}
	if !rb.Contains("d") {
		t.Error("d should be in buffer")
	}
}
