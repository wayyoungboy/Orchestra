package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase() error = %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file not created")
	}
}

func TestMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	migrationsDir := filepath.Join(tmpDir, "migrations")

	// 创建迁移目录和文件
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		t.Fatalf("create migrations dir: %v", err)
	}
	migrationSQL := `CREATE TABLE test (id TEXT PRIMARY KEY);`
	if err := os.WriteFile(filepath.Join(migrationsDir, "001_test.sql"), []byte(migrationSQL), 0644); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("NewDatabase() error = %v", err)
	}
	defer db.Close()

	if err := db.Migrate(migrationsDir); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// 验证表存在
	var count int
	if err := db.DB().QueryRow("SELECT COUNT(*) FROM test").Scan(&count); err != nil {
		t.Errorf("test table not created: %v", err)
	}
}