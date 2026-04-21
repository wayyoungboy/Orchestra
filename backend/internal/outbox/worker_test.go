package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE outbox (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			attempt_count INTEGER NOT NULL DEFAULT 0,
			last_error TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewWorker(t *testing.T) {
	db := testDB(t)
	cfg := DefaultConfig()
	var wg sync.WaitGroup
	wg.Add(1)
	w := New(db, cfg, func(ctx context.Context, item *Item) error {
		wg.Done()
		return nil
	})

	if w == nil {
		t.Fatal("expected non-nil worker")
	}
	if w.config.MaxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", w.config.MaxRetries)
	}
}

func TestEnqueueAndProcess(t *testing.T) {
	db := testDB(t)

	var mu sync.Mutex
	processed := []string{}

	cfg := DefaultConfig()
	cfg.PollInterval = 50 * time.Millisecond

	w := New(db, cfg, func(ctx context.Context, item *Item) error {
		mu.Lock()
		processed = append(processed, item.ID)
		mu.Unlock()
		return nil
	})

	ctx := context.Background()
	id, err := w.Enqueue(ctx, "conv-1", "sender-1", `{"text":"hello"}`)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	w.Start(ctx)

	// Wait for processing
	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		if len(processed) > 0 {
			mu.Unlock()
			break
		}
		mu.Unlock()
	}
	w.Stop()

	mu.Lock()
	if len(processed) != 1 {
		t.Errorf("expected 1 processed, got %d", len(processed))
	} else if processed[0] != id {
		t.Errorf("expected id %s, got %s", id, processed[0])
	}
	mu.Unlock()
}

func TestEnqueueJSON(t *testing.T) {
	db := testDB(t)

	w := New(db, DefaultConfig(), func(ctx context.Context, item *Item) error {
		return nil
	})

	ctx := context.Background()
	id, err := w.EnqueueJSON(ctx, "conv-1", "sender-1", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("enqueue json: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}
}

func TestEnqueueJSONInvalid(t *testing.T) {
	db := testDB(t)
	w := New(db, DefaultConfig(), func(ctx context.Context, item *Item) error {
		return nil
	})

	// channel cannot be JSON marshaled
	_, err := w.EnqueueJSON(context.Background(), "conv-1", "sender-1", make(chan int))
	if err == nil {
		t.Error("expected error for unmarshalable type")
	}
}

func TestRetryExhaustionBecomesDead(t *testing.T) {
	db := testDB(t)

	cfg := DefaultConfig()
	cfg.MaxRetries = 2
	cfg.PollInterval = 50 * time.Millisecond
	cfg.MinBackoff = 10 * time.Millisecond

	callCount := 0
	w := New(db, cfg, func(ctx context.Context, item *Item) error {
		callCount++
		return fmt.Errorf("always fails")
	})

	ctx := context.Background()
	_, err := w.Enqueue(ctx, "conv-1", "sender-1", "fail me")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	w.Start(ctx)
	time.Sleep(800 * time.Millisecond)
	w.Stop()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM outbox WHERE status = 'dead'").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 dead item, got %d", count)
	}
}

func TestStats(t *testing.T) {
	db := testDB(t)

	// Insert items with different statuses
	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('1', 'c', 's', 'x', 'pending', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())
	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('2', 'c', 's', 'x', 'sent', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())
	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('3', 'c', 's', 'x', 'sent', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())

	w := New(db, DefaultConfig(), nil)
	stats, err := w.Stats(context.Background())
	if err != nil {
		t.Fatalf("stats: %v", err)
	}

	if stats[StatusPending] != 1 {
		t.Errorf("expected 1 pending, got %d", stats[StatusPending])
	}
	if stats[StatusSent] != 2 {
		t.Errorf("expected 2 sent, got %d", stats[StatusSent])
	}
}

func TestClearCompleted(t *testing.T) {
	db := testDB(t)

	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('1', 'c', 's', 'x', 'sent', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())
	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('2', 'c', 's', 'x', 'pending', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())

	w := New(db, DefaultConfig(), nil)
	n, err := w.ClearCompleted(context.Background())
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted, got %d", n)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM outbox").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 remaining, got %d", count)
	}
}

func TestBackoff(t *testing.T) {
	w := &Worker{config: Config{MinBackoff: time.Second, MaxBackoff: 30 * time.Second}}

	if d := w.backoff(0); d != time.Second {
		t.Errorf("backoff(0) = %v, want 1s", d)
	}
	if d := w.backoff(1); d != 2*time.Second {
		t.Errorf("backoff(1) = %v, want 2s", d)
	}
	if d := w.backoff(5); d != 30*time.Second {
		t.Errorf("backoff(5) = %v, want 30s (capped)", d)
	}
}

func TestShouldRetry(t *testing.T) {
	w := &Worker{config: Config{MaxRetries: 3, MinBackoff: time.Millisecond}}

	// Fresh failed item with 0 attempts, backoff elapsed
	item := &Item{Status: StatusFailed, AttemptCount: 0, UpdatedAt: time.Now().Add(-10 * time.Millisecond)}
	if !w.shouldRetry(item) {
		t.Error("expected should retry for fresh item")
	}

	// Maxed out attempts
	item = &Item{Status: StatusFailed, AttemptCount: 3, UpdatedAt: time.Now().Add(-time.Hour)}
	if w.shouldRetry(item) {
		t.Error("expected no retry for max attempts")
	}

	// Failed item too recent for backoff
	item = &Item{Status: StatusFailed, AttemptCount: 1, UpdatedAt: time.Now()}
	w.config.MinBackoff = time.Hour
	if w.shouldRetry(item) {
		t.Error("expected no retry during backoff window")
	}
}

func TestProcessNext_ContextCancellation(t *testing.T) {
	db := testDB(t)
	db.Exec(`INSERT INTO outbox (id, conversation_id, sender_id, content, status, attempt_count, created_at, updated_at) VALUES ('1', 'c', 's', 'x', 'pending', 0, ?, ?)`, time.Now().Unix(), time.Now().Unix())

	slow := make(chan struct{})
	w := New(db, DefaultConfig(), func(ctx context.Context, item *Item) error {
		<-slow // block forever
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
	w.Stop()
}
