// Package eventbus provides a simple in-memory pub/sub event bus.
package eventbus

import (
	"sync"
	"time"
)

// EventType identifies the kind of event.
type EventType string

const (
	EventProcessStateChanged EventType = "process-state-changed"
	EventFileChanged         EventType = "file-changed"
	EventWorkerActivity      EventType = "worker-activity-changed"
	EventQueueRequestAdded   EventType = "queue-request-added"
	EventQueueRequestRemoved EventType = "queue-request-removed"
	EventQueuePositionChanged EventType = "queue-position-changed"
	EventMemberStatusChanged EventType = "member-status-changed"
	EventSessionCreated      EventType = "session-created"
	EventMessageReceived     EventType = "message-received"
)

// Event is the base interface for all events.
type Event struct {
	Type      EventType
	Timestamp time.Time
	Payload   map[string]any
}

// Handler is a function that receives events.
type Handler func(Event)

// subscriber wraps a handler with a unique ID for removal.
type subscriber struct {
	id      uint64
	handler Handler
}

// EventBus is a simple in-memory pub/sub system.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]subscriber
	nextID      uint64
}

// New creates a new EventBus.
func New() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]subscriber),
	}
}

// Subscribe registers a handler for a specific event type.
// Returns an unsubscribe function.
func (eb *EventBus) Subscribe(eventType EventType, handler Handler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber{id: id, handler: handler})
	return func() {
		eb.unsubscribeByID(eventType, id)
	}
}

// Unsubscribe removes a handler for a specific event type.
// Deprecated: use the unsubscribe function returned by Subscribe instead.
func (eb *EventBus) Unsubscribe(eventType EventType, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	handlers := eb.subscribers[eventType]
	for i, s := range handlers {
		if s.handler == nil && handler == nil {
			eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) unsubscribeByID(eventType EventType, id uint64) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	handlers := eb.subscribers[eventType]
	for i, s := range handlers {
		if s.id == id {
			eb.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

// Emit sends an event to all subscribers of its type.
// Handlers that panic are silently skipped.
func (eb *EventBus) Emit(event Event) {
	eb.mu.RLock()
	subs := append([]subscriber(nil), eb.subscribers[event.Type]...)
	eb.mu.RUnlock()

	for _, s := range subs {
		func() {
			defer func() { recover() }()
			s.handler(event)
		}()
	}
}

// EmitPayload is a convenience method that creates an Event with a payload.
func (eb *EventBus) EmitPayload(eventType EventType, payload map[string]any) {
	eb.Emit(Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

// HandlerCount returns the number of subscribers for an event type.
func (eb *EventBus) HandlerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[eventType])
}

