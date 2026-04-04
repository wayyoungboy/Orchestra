package terminal

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock validator for testing
type mockValidator struct {
	allowedCommands map[string]bool
}

func (m *mockValidator) ValidateCommand(cmd string) error {
	if !m.allowedCommands[cmd] {
		return errors.New("command not allowed")
	}
	return nil
}

func TestProcessPool_AcquireAndRelease(t *testing.T) {
	pool := NewProcessPool(2, 30*time.Minute)

	config := ProcessConfig{
		ID:        "test-1",
		Command:   "echo",
		Args:      []string{"hello"},
		Workspace: "/tmp",
	}

	session, err := pool.Acquire(context.Background(), config)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}

	if pool.Count() != 1 {
		t.Errorf("expected pool count 1, got %d", pool.Count())
	}

	pool.Release(session.ID)

	if pool.Count() != 0 {
		t.Errorf("expected pool count 0 after release, got %d", pool.Count())
	}
}

func TestProcessPool_Full(t *testing.T) {
	pool := NewProcessPool(1, 30*time.Minute)

	config := ProcessConfig{
		ID:        "test-1",
		Command:   "sleep",
		Args:      []string{"10"},
		Workspace: "/tmp",
	}

	_, err := pool.Acquire(context.Background(), config)
	if err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}

	config2 := ProcessConfig{
		ID:        "test-2",
		Command:   "sleep",
		Args:      []string{"10"},
		Workspace: "/tmp",
	}

	_, err = pool.Acquire(context.Background(), config2)
	if err != ErrProcessPoolFull {
		t.Errorf("expected ErrProcessPoolFull, got %v", err)
	}

	pool.Release("test-1")
}

func TestProcessPool_GetNotFound(t *testing.T) {
	pool := NewProcessPool(2, 30*time.Minute)

	_, err := pool.Get("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestProcessPool_CommandValidation(t *testing.T) {
	pool := NewProcessPool(2, 30*time.Minute)
	pool.SetValidator(&mockValidator{
		allowedCommands: map[string]bool{"echo": true, "ls": true},
	})

	tests := []struct {
		name      string
		command   string
		expectErr error
	}{
		{
			name:      "allowed command",
			command:   "echo",
			expectErr: nil,
		},
		{
			name:      "disallowed command",
			command:   "rm",
			expectErr: ErrCommandNotAllowed,
		},
		{
			name:      "another allowed command",
			command:   "ls",
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ProcessConfig{
				ID:        "test-" + tt.name,
				Command:   tt.command,
				Args:      []string{},
				Workspace: "/tmp",
			}

			session, err := pool.Acquire(context.Background(), config)
			if err != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}

			if err == nil && session != nil {
				pool.Release(session.ID)
			}
		})
	}
}

func TestProcessPool_NoValidator(t *testing.T) {
	pool := NewProcessPool(2, 30*time.Minute)
	// No validator set

	config := ProcessConfig{
		ID:        "test-no-validator",
		Command:   "echo", // Use a valid command that exists
		Args:      []string{"test"},
		Workspace: "/tmp",
	}

	session, err := pool.Acquire(context.Background(), config)
	if err != nil {
		t.Errorf("without validator, valid command should be allowed, got error: %v", err)
	}

	if session != nil {
		pool.Release(session.ID)
	}
}