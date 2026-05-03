package agent

import (
	"errors"
	"sync"
	"time"
)

var ErrQueueFull = errors.New("dispatch queue full")

// DispatchItem represents a queued message waiting to be sent to an agent.
type DispatchItem struct {
	Content   string
	SenderID  string
	CreatedAt time.Time
	MergeOf   int // how many original messages were merged into this item
}

// RingBuffer is a fixed-size circular buffer for dedup tracking.
type RingBuffer struct {
	mu    sync.Mutex
	buf   []string
	size  int
	head  int
	count int
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:  make([]string, size),
		size: size,
	}
}

// Contains checks if the key exists in the buffer.
func (r *RingBuffer) Contains(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < r.count; i++ {
		idx := (r.head - r.count + i + r.size) % r.size
		if r.buf[idx] == key {
			return true
		}
	}
	return false
}

// Push adds a key to the buffer, evicting the oldest if full.
func (r *RingBuffer) Push(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = key
	r.head = (r.head + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

// DispatchQueue buffers messages when an agent is busy, with merge and dedup.
type DispatchQueue struct {
	mu          sync.Mutex
	items       []DispatchItem
	maxSize     int
	dedupRing   *RingBuffer
	inflight    *DispatchItem
	forceTimer  *time.Timer
	forceFlush  func([]DispatchItem)
}

// NewDispatchQueue creates a queue with the given maximum size.
func NewDispatchQueue() *DispatchQueue {
	q := &DispatchQueue{
		items:     make([]DispatchItem, 0, DispatchQueueSize),
		maxSize:   DispatchQueueSize,
		dedupRing: NewRingBuffer(DedupRingSize),
	}
	return q
}

// SetForceFlush sets the callback for the force-flush timer.
func (q *DispatchQueue) SetForceFlush(fn func([]DispatchItem)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.forceFlush = fn
}

// Enqueue adds a message to the queue. Within MergeWindow, messages from
// the same sender are merged. Duplicate content within DedupWindow is dropped.
func (q *DispatchQueue) Enqueue(item DispatchItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if item.Content == "" {
		return nil
	}

	// Dedup check
	dedupKey := item.SenderID + ":" + item.Content
	if q.dedupRing.Contains(dedupKey) {
		return nil
	}
	q.dedupRing.Push(dedupKey)

	// Try merge with last item from same sender within merge window
	if len(q.items) > 0 {
		last := &q.items[len(q.items)-1]
		if last.SenderID == item.SenderID && time.Since(last.CreatedAt) < MergeWindow {
			merged := last.Content + "\n" + item.Content
			if len(merged) > MaxMergedLength {
				merged = merged[:MaxMergedLength]
			}
			last.Content = merged
			last.MergeOf++
			return nil
		}
	}

	if len(q.items) >= q.maxSize {
		return ErrQueueFull
	}

	item.CreatedAt = time.Now()
	item.MergeOf = 1
	q.items = append(q.items, item)

	// Start force-flush timer on first item
	if len(q.items) == 1 && q.forceFlush != nil {
		if q.forceTimer != nil {
			q.forceTimer.Stop()
		}
		q.forceTimer = time.AfterFunc(ForceFlushTimeout, func() {
			items := q.Flush()
			if q.forceFlush != nil && len(items) > 0 {
				q.forceFlush(items)
			}
		})
	}

	return nil
}

// Flush drains and returns all queued items in FIFO order.
func (q *DispatchQueue) Flush() []DispatchItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.forceTimer != nil {
		q.forceTimer.Stop()
		q.forceTimer = nil
	}

	items := q.items
	q.items = make([]DispatchItem, 0, DispatchQueueSize)
	if len(items) > 0 {
		q.inflight = &items[0]
	}
	return items
}

// ClearInflight returns the inflight item to the front of the queue
// (called when a send fails).
func (q *DispatchQueue) ClearInflight() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.inflight != nil {
		restored := *q.inflight
		q.inflight = nil
		q.items = append([]DispatchItem{restored}, q.items...)
	}
}

// Len returns the number of items in the queue.
func (q *DispatchQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// IsFull returns whether the queue has reached capacity.
func (q *DispatchQueue) IsFull() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items) >= q.maxSize
}
