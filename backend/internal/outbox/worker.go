// Package outbox implements the outbox pattern for reliable message delivery.
package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ItemStatus represents the state of an outbox item.
type ItemStatus string

const (
	StatusPending  ItemStatus = "pending"
	StatusSending  ItemStatus = "sending"
	StatusSent     ItemStatus = "sent"
	StatusFailed   ItemStatus = "failed"
	StatusDead     ItemStatus = "dead"
)

// Item represents a single outbox entry.
type Item struct {
	ID            string     `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID      string     `json:"sender_id"`
	Content       string     `json:"content"`
	Status        ItemStatus `json:"status"`
	AttemptCount  int        `json:"attempt_count"`
	LastError     string     `json:"last_error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SenderFunc is the function that actually sends a message.
// It returns an error if the send fails.
type SenderFunc func(ctx context.Context, item *Item) error

// Config holds outbox worker configuration.
type Config struct {
	MaxRetries    int
	MinBackoff    time.Duration
	MaxBackoff    time.Duration
	PollInterval  time.Duration
	MaxConcurrent int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries:    5,
		MinBackoff:    1 * time.Second,
		MaxBackoff:    30 * time.Second,
		PollInterval:  5 * time.Second,
		MaxConcurrent: 5,
	}
}

// Worker is a background worker that processes outbox items.
type Worker struct {
	db         *sql.DB
	config     Config
	sender     SenderFunc
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	processing map[string]bool
}

// New creates a new outbox worker.
func New(db *sql.DB, config Config, sender SenderFunc) *Worker {
	return &Worker{
		db:         db,
		config:     config,
		sender:     sender,
		processing: make(map[string]bool),
	}
}

// Start begins processing outbox items in the background.
func (w *Worker) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.run(workerCtx)
	}()
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
}

func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *Worker) processNext(ctx context.Context) {
	items, err := w.pendingItems(ctx)
	if err != nil {
		return
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !w.tryClaim(item.ID) {
			continue
		}

		w.wg.Add(1)
		go func(it *Item) {
			defer w.wg.Done()
			defer w.release(it.ID)
			w.sendWithRetry(ctx, it)
		}(item)
	}
}

func (w *Worker) pendingItems(ctx context.Context) ([]*Item, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, status, attempt_count,
		       COALESCE(last_error, ''), created_at, updated_at
		FROM outbox
		WHERE status IN ('pending', 'failed')
		ORDER BY created_at ASC
		LIMIT 50
	`

	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query pending outbox: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var it Item
		var createdAt, updatedAt int64
		if err := rows.Scan(&it.ID, &it.ConversationID, &it.SenderID, &it.Content,
			&it.Status, &it.AttemptCount, &it.LastError, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox item: %w", err)
		}
		it.CreatedAt = time.Unix(createdAt, 0)
		it.UpdatedAt = time.Unix(updatedAt, 0)

		// Check backoff
		if it.Status == StatusFailed && !w.shouldRetry(&it) {
			continue
		}
		items = append(items, &it)
	}
	return items, rows.Err()
}

func (w *Worker) shouldRetry(item *Item) bool {
	if item.AttemptCount >= w.config.MaxRetries {
		return false
	}
	backoff := w.backoff(item.AttemptCount)
	return time.Since(item.UpdatedAt) >= backoff
}

func (w *Worker) backoff(attempt int) time.Duration {
	delay := w.config.MinBackoff * (1 << attempt)
	if w.config.MaxBackoff > 0 && delay > w.config.MaxBackoff {
		delay = w.config.MaxBackoff
	}
	return delay
}

func (w *Worker) tryClaim(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.processing[id]; ok {
		return false
	}
	w.processing[id] = true
	return true
}

func (w *Worker) release(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.processing, id)
}

func (w *Worker) sendWithRetry(ctx context.Context, item *Item) {
	var lastErr error
	for item.AttemptCount < w.config.MaxRetries {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := w.sendOne(ctx, item); err != nil {
			lastErr = err
			item.LastError = err.Error()
			// Back off between retries within the same processing cycle
			delay := w.config.MinBackoff
			if delay > time.Second {
				delay = time.Second
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			continue
		}
		w.markSent(item.ID)
		return
	}
	// Exhausted retries
	if lastErr != nil {
		w.markDead(item.ID, fmt.Errorf("max retries exceeded: %w", lastErr))
	}
}

func (w *Worker) sendOne(ctx context.Context, item *Item) error {
	if err := w.markSending(item.ID); err != nil {
		return err
	}

	item.AttemptCount++
	if err := w.sender(ctx, item); err != nil {
		item.UpdatedAt = time.Now()
		_ = w.markFailed(item.ID, err) // record failure in DB
		return err                      // return original sender error
	}
	return nil
}

func (w *Worker) markSending(id string) error {
	_, err := w.db.Exec(`
		UPDATE outbox SET status = 'sending', updated_at = ? WHERE id = ?
	`, time.Now().Unix(), id)
	return err
}

func (w *Worker) markSent(id string) error {
	_, err := w.db.Exec(`
		UPDATE outbox SET status = 'sent', updated_at = ? WHERE id = ?
	`, time.Now().Unix(), id)
	return err
}

func (w *Worker) markFailed(id string, sendErr error) error {
	_, err := w.db.Exec(`
		UPDATE outbox SET status = 'failed', attempt_count = attempt_count + 1,
		                  last_error = ?, updated_at = ? WHERE id = ?
	`, sendErr.Error(), time.Now().Unix(), id)
	return err
}

func (w *Worker) markDead(id string, sendErr error) error {
	_, err := w.db.Exec(`
		UPDATE outbox SET status = 'dead', last_error = ?, updated_at = ? WHERE id = ?
	`, sendErr.Error(), time.Now().Unix(), id)
	return err
}

// Enqueue adds a new message to the outbox.
func (w *Worker) Enqueue(ctx context.Context, conversationID, senderID, content string) (string, error) {
	id := fmt.Sprintf("obx-%d", time.Now().UnixNano())
	now := time.Now().Unix()

	_, err := w.db.ExecContext(ctx, `
		INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'pending', 0, ?, ?)
	`, id, conversationID, senderID, content, now, now)

	return id, err
}

// EnqueueJSON marshals a value and enqueues it as JSON.
func (w *Worker) EnqueueJSON(ctx context.Context, conversationID, senderID string, v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal message: %w", err)
	}
	return w.Enqueue(ctx, conversationID, senderID, string(data))
}

// Stats returns counts by status.
func (w *Worker) Stats(ctx context.Context) (map[ItemStatus]int, error) {
	rows, err := w.db.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM outbox GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("query outbox stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[ItemStatus]int)
	for rows.Next() {
		var status ItemStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan stat: %w", err)
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

// ClearCompleted removes sent items from the outbox.
func (w *Worker) ClearCompleted(ctx context.Context) (int64, error) {
	result, err := w.db.ExecContext(ctx, `DELETE FROM outbox WHERE status = 'sent'`)
	if err != nil {
		return 0, fmt.Errorf("clear completed: %w", err)
	}
	return result.RowsAffected()
}
