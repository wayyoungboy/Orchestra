package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db   *sql.DB
	path string
}

func NewDatabase(path string) (*Database, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Database{db: db, path: path}, nil
}

func (d *Database) DB() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Migrate(migrationsDir string) error {
	// 创建迁移记录表
	if _, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// 读取迁移文件
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		version := entry.Name()
		// 检查是否已应用
		var count int
		if err := d.db.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			version,
		).Scan(&count); err != nil {
			return fmt.Errorf("check migration: %w", err)
		}
		if count > 0 {
			continue
		}

		// 读取并执行迁移文件
		content, err := os.ReadFile(filepath.Join(migrationsDir, version))
		if err != nil {
			return fmt.Errorf("read migration file: %w", err)
		}

		if _, err := d.db.Exec(string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", version, err)
		}

		// 记录迁移
		if _, err := d.db.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			version,
			time.Now().Unix(),
		); err != nil {
			return fmt.Errorf("record migration: %w", err)
		}
	}

	return nil
}