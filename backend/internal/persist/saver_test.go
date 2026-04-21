package persist

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewSaver(t *testing.T) {
	saved := 0
	s := New(func() error {
		saved++
		return nil
	}, 50*time.Millisecond)
	defer s.Stop()

	if s == nil {
		t.Fatal("expected non-nil saver")
	}
}

func TestSaveCoalesces(t *testing.T) {
	count := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}, 50*time.Millisecond)
	defer s.Stop()

	// Rapid calls should coalesce into 1 save
	for i := 0; i < 10; i++ {
		s.Save()
	}

	time.Sleep(150 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("expected 1 save after coalescing, got %d", count)
	}
}

func TestSaveSeparatesByDelay(t *testing.T) {
	count := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}, 50*time.Millisecond)
	defer s.Stop()

	s.Save()
	time.Sleep(100 * time.Millisecond) // first save fires
	s.Save()
	time.Sleep(100 * time.Millisecond) // second save fires

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Errorf("expected 2 saves, got %d", count)
	}
}

func TestSaveImmediate(t *testing.T) {
	count := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}, 500*time.Millisecond)
	defer s.Stop()

	s.Save()
	// Immediate should fire right away, not wait for debounce
	err := s.SaveImmediate()
	if err != nil {
		t.Fatalf("save immediate: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Errorf("expected 1 save from immediate, got %d", count)
	}
}

func TestStopCancelsPendingSave(t *testing.T) {
	count := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}, 100*time.Millisecond)

	s.Save()
	s.Stop()

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 0 {
		t.Errorf("expected 0 saves after stop, got %d", count)
	}
}

func TestSaveAtomicJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	data := map[string]interface{}{
		"name": "test",
		"count": 42,
	}

	err := SaveAtomicJSON(path, data)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file exists and contains data
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("expected non-empty file")
	}

	// Verify no .tmp file remains
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected tmp file to be cleaned up")
	}
}

func TestSaveAtomicJSONInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	// channel cannot be marshaled
	err := SaveAtomicJSON(path, make(chan int))
	if err == nil {
		t.Error("expected error for unmarshalable type")
	}
}

func TestConcurrentSave(t *testing.T) {
	count := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}, 50*time.Millisecond)
	defer s.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Save()
		}()
	}
	wg.Wait()

	time.Sleep(150 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	// All concurrent saves should coalesce into 1
	if count != 1 {
		t.Errorf("expected 1 save, got %d", count)
	}
}

func TestSaveWithRetry(t *testing.T) {
	attempts := 0
	var mu sync.Mutex
	s := New(func() error {
		mu.Lock()
		attempts++
		defer mu.Unlock()
		if attempts == 1 {
			return os.ErrPermission
		}
		return nil
	}, 20*time.Millisecond)
	defer s.Stop()

	s.Save()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// First attempt fails, retry should succeed
	if attempts < 2 {
		t.Errorf("expected at least 2 attempts (fail + retry), got %d", attempts)
	}
}

func TestSaveAtomicJSONWriteDir(t *testing.T) {
	// Should fail when target directory doesn't exist
	err := SaveAtomicJSON("/nonexistent/dir/data.json", "test")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}
