// Package messagequeue provides a channel-based message queue for agent sessions.
package messagequeue

import (
	"context"
)

// Queue is a bounded message queue that blocks when full.
// Messages are delivered in FIFO order via a read channel.
type Queue[T any] struct {
	ch chan T
}

// New creates a new message queue with the given capacity.
func New[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		ch: make(chan T, capacity),
	}
}

// Push adds a message to the queue.
// Blocks if the queue is full until space is available.
func (q *Queue[T]) Push(msg T) {
	q.ch <- msg
}

// PushNonBlocking attempts to push a message without blocking.
// Returns true if the message was queued, false if the queue is full.
func (q *Queue[T]) PushNonBlocking(msg T) bool {
	select {
	case q.ch <- msg:
		return true
	default:
		return false
	}
}

// Messages returns the read channel for consuming messages.
func (q *Queue[T]) Messages() <-chan T {
	return q.ch
}

// Len returns the current number of messages in the queue.
func (q *Queue[T]) Len() int {
	return len(q.ch)
}

// Capacity returns the queue capacity.
func (q *Queue[T]) Capacity() int {
	return cap(q.ch)
}

// Close closes the queue. No more messages can be pushed after this.
func (q *Queue[T]) Close() {
	close(q.ch)
}

// Drain reads and discards all pending messages.
func (q *Queue[T]) Drain() int {
	count := 0
	for {
		select {
		case <-q.ch:
			count++
		default:
			return count
		}
	}
}

// WaitOrContext waits for a message or context cancellation.
func (q *Queue[T]) WaitOrContext(ctx context.Context) (T, error) {
	select {
	case msg, ok := <-q.ch:
		var zero T
		if !ok {
			return zero, ErrQueueClosed
		}
		return msg, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// ErrQueueClosed is returned when trying to read from a closed queue.
var ErrQueueClosed = &queueClosedError{}

type queueClosedError struct{}

func (e *queueClosedError) Error() string {
	return "message queue is closed"
}
