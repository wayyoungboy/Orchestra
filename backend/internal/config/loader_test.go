package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.HTTPAddr != ":8080" {
		t.Errorf("expected default HTTPAddr :8080, got %s", cfg.Server.HTTPAddr)
	}
	if cfg.Terminal.MaxSessions != 10 {
		t.Errorf("expected default MaxSessions 10, got %d", cfg.Terminal.MaxSessions)
	}
}

func TestLoadFromFile(t *testing.T) {
	content := `
server:
  http_addr: ":9090"
terminal:
  max_sessions: 5
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.HTTPAddr != ":9090" {
		t.Errorf("expected HTTPAddr :9090, got %s", cfg.Server.HTTPAddr)
	}
	if cfg.Terminal.MaxSessions != 5 {
		t.Errorf("expected MaxSessions 5, got %d", cfg.Terminal.MaxSessions)
	}
}

func TestEncryptionKeyFromEnv(t *testing.T) {
	os.Setenv("ORCHESTRA_ENCRYPTION_KEY", "test-key-32-bytes-long-12345678")
	defer os.Unsetenv("ORCHESTRA_ENCRYPTION_KEY")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Security.EncryptionKey != "test-key-32-bytes-long-12345678" {
		t.Errorf("expected encryption key from env, got %s", cfg.Security.EncryptionKey)
	}
}