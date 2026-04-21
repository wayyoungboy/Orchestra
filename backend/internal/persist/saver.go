// Package persist provides debounced file persistence with concurrent safety.
package persist

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Saver provides debounced file write with atomic replace semantics.
type Saver struct {
	mu         sync.Mutex
	saveFunc   func() error
	delay      time.Duration
	timer      *time.Timer
	savePending bool
}

// New creates a debounced saver. Save() calls will coalesce within the given delay.
func New(saveFn func() error, delay time.Duration) *Saver {
	return &Saver{
		saveFunc: saveFn,
		delay:    delay,
	}
}

// Save schedules a save with debounce. Multiple calls within the delay window coalesce.
func (s *Saver) Save() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.savePending {
		return
	}
	s.savePending = true

	if s.timer != nil {
		s.timer.Stop()
	}

	s.timer = time.AfterFunc(s.delay, func() {
		s.mu.Lock()
		s.savePending = false
		s.timer = nil
		saveFn := s.saveFunc
		s.mu.Unlock()
		if err := saveFn(); err != nil {
			// Retry once after a short delay
			time.Sleep(s.delay / 2)
			_ = saveFn()
		}
	})
}

// Stop cancels any pending save and stops the timer.
func (s *Saver) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.savePending = false
}

// SaveImmediate forces an immediate save, bypassing the debounce.
func (s *Saver) SaveImmediate() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.savePending = false
	return s.saveFunc()
}

// SaveAtomicJSON marshals a value and writes it atomically (write to .tmp then rename).
func SaveAtomicJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
