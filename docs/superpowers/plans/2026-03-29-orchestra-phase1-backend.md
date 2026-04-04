# Orchestra Phase 1 - 后端基础架构实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 Go 后端基础架构，包括配置管理、数据库、API 框架、安全模块和进程池。

**Architecture:** Go 单体服务，模块化设计，预留多用户扩展接口。

**Tech Stack:** Go 1.21+, Gin, SQLite, gorilla/websocket, PTY

---

## 文件结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 入口
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置结构
│   │   └── loader.go            # 配置加载
│   ├── storage/
│   │   ├── database.go          # 数据库连接
│   │   └── migrations/
│   │       └── 001_init.sql     # 初始化迁移
│   ├── security/
│   │   ├── crypto.go            # 密钥加密
│   │   └── whitelist.go         # 命令/路径白名单
│   ├── terminal/
│   │   ├── pool.go              # 进程池
│   │   ├── session.go           # CLI 会话
│   │   └── pty.go               # PTY 处理
│   ├── api/
│   │   ├── router.go            # 路由注册
│   │   └── handlers/
│   │       └── health.go        # 健康检查
│   └── ws/
│       └── gateway.go           # WebSocket 网关
├── pkg/
│   └── utils/
│       └── ulid.go              # ULID 生成
├── configs/
│   └── config.yaml              # 配置文件
├── go.mod
└── go.sum
```

---

### Task 1: 项目初始化

**Files:**
- Create: `backend/go.mod`
- Create: `backend/.gitignore`
- Create: `backend/Makefile`

- [ ] **Step 1: 创建 go.mod**

```go
module github.com/orchestra/backend

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/gorilla/websocket v1.5.1
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/oklog/ulid/v2 v2.1.0
    gopkg.in/yaml.v3 v3.0.1
    golang.org/x/crypto v0.21.0
    github.com/creack/pty v1.1.21
)
```

- [ ] **Step 2: 创建 .gitignore**

```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Go
*.test
*.out
go.work
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo

# Environment
.env
.env.local
*.local.yaml

# Data
data/
*.db
*.db-journal
workspaces/

# Logs
*.log
logs/
```

- [ ] **Step 3: 创建 Makefile**

```makefile
.PHONY: build run test clean

build:
	go build -o bin/orchestra ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf data/*.db
```

- [ ] **Step 4: 下载依赖**

Run: `cd backend && go mod tidy`
Expected: 成功下载所有依赖

- [ ] **Step 5: Commit**

```bash
git add backend/go.mod backend/.gitignore backend/Makefile
git commit -m "chore: initialize Go module and project structure"
```

---

### Task 2: 配置管理模块

**Files:**
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/loader.go`
- Create: `backend/configs/config.yaml`

- [ ] **Step 1: 创建配置结构 config.go**

```go
package config

import "time"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Terminal TerminalConfig `yaml:"terminal"`
	Security SecurityConfig `yaml:"security"`
	Storage  StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	WSAddr   string `yaml:"ws_addr"`
}

type TerminalConfig struct {
	MaxSessions int           `yaml:"max_sessions"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type SecurityConfig struct {
	EncryptionKey    string   `yaml:"encryption_key"`
	AllowedCommands  []string `yaml:"allowed_commands"`
	AllowedPaths     []string `yaml:"allowed_paths"`
}

type StorageConfig struct {
	Database   string `yaml:"database"`
	Workspaces string `yaml:"workspaces"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPAddr: ":8080",
			WSAddr:   ":8081",
		},
		Terminal: TerminalConfig{
			MaxSessions: 10,
			IdleTimeout: 30 * time.Minute,
		},
		Security: SecurityConfig{
			EncryptionKey:   "",
			AllowedCommands: []string{"claude", "gemini", "codex", "qwen"},
			AllowedPaths:    []string{},
		},
		Storage: StorageConfig{
			Database:   "./data/orchestra.db",
			Workspaces: "./workspaces",
		},
	}
}
```

- [ ] **Step 2: 创建配置加载器 loader.go**

```go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		path = os.Getenv("ORCHESTRA_CONFIG")
		if path == "" {
			return cfg, nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 从环境变量读取加密密钥
	if key := os.Getenv("ORCHESTRA_ENCRYPTION_KEY"); key != "" {
		cfg.Security.EncryptionKey = key
	}

	// 解析路径中的环境变量
	cfg.Storage.Database = expandPath(cfg.Storage.Database)
	cfg.Storage.Workspaces = expandPath(cfg.Storage.Workspaces)

	return cfg, nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
```

- [ ] **Step 3: 创建默认配置文件 config.yaml**

```yaml
server:
  http_addr: ":8080"
  ws_addr: ":8081"

terminal:
  max_sessions: 10
  idle_timeout: 30m

security:
  encryption_key: "${ORCHESTRA_ENCRYPTION_KEY}"
  allowed_commands:
    - claude
    - gemini
    - codex
    - qwen
  allowed_paths:
    - ~/projects
    - ~/workspaces

storage:
  database: "./data/orchestra.db"
  workspaces: "./workspaces"
```

- [ ] **Step 4: 创建配置测试**

Create: `backend/internal/config/loader_test.go`

```go
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
```

- [ ] **Step 5: 运行测试**

Run: `cd backend && go test -v ./internal/config/...`
Expected: 所有测试通过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/config/ backend/configs/
git commit -m "feat: add configuration module with YAML and env support"
```

---

### Task 3: 数据库初始化

**Files:**
- Create: `backend/internal/storage/database.go`
- Create: `backend/internal/storage/migrations/001_init.sql`

- [ ] **Step 1: 创建数据库连接 database.go**

```go
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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
```

- [ ] **Step 2: 创建初始化迁移 001_init.sql**

```sql
-- 工作区表
CREATE TABLE IF NOT EXISTS workspaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    last_opened_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);

-- 成员表（Agent/CLI 配置）
CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    role_type TEXT NOT NULL,
    role_key TEXT,
    avatar TEXT,
    terminal_type TEXT,
    terminal_command TEXT,
    terminal_path TEXT,
    auto_start_terminal INTEGER DEFAULT 1,
    status TEXT DEFAULT 'online',
    created_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- 会话表
CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    type TEXT NOT NULL,
    name TEXT,
    target_id TEXT,
    member_ids TEXT NOT NULL,
    pinned INTEGER DEFAULT 0,
    muted INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL,
    is_ai INTEGER DEFAULT 0,
    attachment TEXT,
    status TEXT DEFAULT 'sent',
    created_at INTEGER NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

-- 工作流模板表
CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    steps TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- 设置表
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- API 密钥表（加密存储）
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    encrypted_key TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_members_workspace ON members(workspace_id);
CREATE INDEX IF NOT EXISTS idx_conversations_workspace ON conversations(workspace_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at);
```

- [ ] **Step 3: 创建数据库测试**

Create: `backend/internal/storage/database_test.go`

```go
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
```

- [ ] **Step 4: 运行测试**

Run: `cd backend && go test -v ./internal/storage/...`
Expected: 所有测试通过

- [ ] **Step 5: Commit**

```bash
git add backend/internal/storage/
git commit -m "feat: add SQLite database with migration support"
```

---

### Task 4: 安全模块

**Files:**
- Create: `backend/internal/security/crypto.go`
- Create: `backend/internal/security/whitelist.go`
- Create: `backend/internal/security/crypto_test.go`
- Create: `backend/internal/security/whitelist_test.go`

- [ ] **Step 1: 创建加密模块 crypto.go**

```go
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

type KeyEncryptor struct {
	key []byte
}

func NewKeyEncryptor(key string) (*KeyEncryptor, error) {
	if len(key) < 32 {
		return nil, ErrInvalidKey
	}

	k := make([]byte, 32)
	copy(k, key[:32])

	return &KeyEncryptor{key: k}, nil
}

func (e *KeyEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *KeyEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
```

- [ ] **Step 2: 创建白名单模块 whitelist.go**

```go
package security

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrCommandNotAllowed = errors.New("command not in whitelist")
	ErrPathNotAllowed    = errors.New("path not in whitelist")
)

type Whitelist struct {
	commands map[string]bool
	paths    []string
}

func NewWhitelist(commands, paths []string) *Whitelist {
	cmdMap := make(map[string]bool)
	for _, cmd := range commands {
		cmdMap[filepath.Base(cmd)] = true
	}

	return &Whitelist{
		commands: cmdMap,
		paths:    paths,
	}
}

func (w *Whitelist) ValidateCommand(cmd string) error {
	base := filepath.Base(cmd)
	if !w.commands[base] {
		return ErrCommandNotAllowed
	}
	return nil
}

func (w *Whitelist) ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range w.paths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absAllowed) {
			return nil
		}
	}

	return ErrPathNotAllowed
}

func (w *Whitelist) AllowedCommands() []string {
	result := make([]string, 0, len(w.commands))
	for cmd := range w.commands {
		result = append(result, cmd)
	}
	return result
}

func (w *Whitelist) AllowedPaths() []string {
	return w.paths
}
```

- [ ] **Step 3: 创建加密测试 crypto_test.go**

```go
package security

import (
	"testing"
)

func TestKeyEncryptor_EncryptDecrypt(t *testing.T) {
	key := "test-key-32-bytes-long-12345678"
	encryptor, err := NewKeyEncryptor(key)
	if err != nil {
		t.Fatalf("NewKeyEncryptor() error = %v", err)
	}

	plaintext := "my-secret-api-key"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ciphertext == plaintext {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestKeyEncryptor_InvalidKey(t *testing.T) {
	_, err := NewKeyEncryptor("short")
	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestKeyEncryptor_InvalidCiphertext(t *testing.T) {
	encryptor, _ := NewKeyEncryptor("test-key-32-bytes-long-12345678")

	_, err := encryptor.Decrypt("invalid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}
```

- [ ] **Step 4: 创建白名单测试 whitelist_test.go**

```go
package security

import (
	"testing"
)

func TestWhitelist_ValidateCommand(t *testing.T) {
	w := NewWhitelist([]string{"claude", "gemini", "python"}, nil)

	if err := w.ValidateCommand("claude"); err != nil {
		t.Errorf("claude should be allowed: %v", err)
	}
	if err := w.ValidateCommand("/usr/bin/claude"); err != nil {
		t.Errorf("/usr/bin/claude should be allowed: %v", err)
	}
	if err := w.ValidateCommand("rm"); err == nil {
		t.Error("rm should not be allowed")
	}
}

func TestWhitelist_ValidatePath(t *testing.T) {
	w := NewWhitelist(nil, []string{"/home/user/projects", "/var/workspaces"})

	if err := w.ValidatePath("/home/user/projects/myapp"); err != nil {
		t.Errorf("projects/myapp should be allowed: %v", err)
	}
	if err := w.ValidatePath("/etc/passwd"); err == nil {
		t.Error("/etc/passwd should not be allowed")
	}
}

func TestWhitelist_Empty(t *testing.T) {
	w := NewWhitelist(nil, nil)

	if err := w.ValidateCommand("claude"); err == nil {
		t.Error("empty whitelist should deny all commands")
	}
}
```

- [ ] **Step 5: 运行测试**

Run: `cd backend && go test -v ./internal/security/...`
Expected: 所有测试通过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/security/
git commit -m "feat: add security module with AES encryption and whitelist validation"
```

---

### Task 5: 进程池模块

**Files:**
- Create: `backend/internal/terminal/pool.go`
- Create: `backend/internal/terminal/session.go`
- Create: `backend/internal/terminal/pty.go`
- Create: `backend/internal/terminal/pool_test.go`

- [ ] **Step 1: 创建进程池 pool.go**

```go
package terminal

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrProcessPoolFull  = errors.New("process pool is full")
	ErrSessionNotFound  = errors.New("session not found")
)

type ProcessConfig struct {
	ID           string
	Command      string
	Args         []string
	Workspace    string
	TerminalType string
	Env          []string
}

type ProcessPool struct {
	mu          sync.RWMutex
	sessions    map[string]*ProcessSession
	maxSessions int
	idleTimeout time.Duration
}

func NewProcessPool(maxSessions int, idleTimeout time.Duration) *ProcessPool {
	p := &ProcessPool{
		sessions:    make(map[string]*ProcessSession),
		maxSessions: maxSessions,
		idleTimeout: idleTimeout,
	}
	go p.cleanupIdleSessions()
	return p
}

func (p *ProcessPool) Acquire(ctx context.Context, config ProcessConfig) (*ProcessSession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.sessions) >= p.maxSessions {
		return nil, ErrProcessPoolFull
	}

	session, err := createSession(config)
	if err != nil {
		return nil, err
	}

	p.sessions[session.ID] = session
	return session, nil
}

func (p *ProcessPool) Get(id string) (*ProcessSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, ok := p.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (p *ProcessPool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if session, ok := p.sessions[id]; ok {
		session.Kill()
		delete(p.sessions, id)
	}
}

func (p *ProcessPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

func (p *ProcessPool) cleanupIdleSessions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		for id, session := range p.sessions {
			if time.Since(session.LastActive) > p.idleTimeout {
				session.Kill()
				delete(p.sessions, id)
			}
		}
		p.mu.Unlock()
	}
}
```

- [ ] **Step 2: 创建会话 session.go**

```go
package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type ProcessSession struct {
	ID           string
	PID          int
	Cmd          *exec.Cmd
	Workspace    string
	TerminalType string
	CreatedAt    time.Time
	LastActive   time.Time

	mu         sync.Mutex
	pty        *os.File
	OutputChan chan []byte
	ErrorChan  chan error
	DoneChan   chan struct{}
	done       bool
}

func createSession(config ProcessConfig) (*ProcessSession, error) {
	session := &ProcessSession{
		ID:           config.ID,
		Workspace:    config.Workspace,
		TerminalType: config.TerminalType,
		CreatedAt:    time.Now(),
		LastActive:   time.Now(),
		OutputChan:   make(chan []byte, 1024),
		ErrorChan:    make(chan error, 16),
		DoneChan:     make(chan struct{}),
	}

	cmd := exec.Command(config.Command, config.Args...)
	cmd.Dir = config.Workspace
	cmd.Env = append(os.Environ(), config.Env...)

	pty, err := startPty(cmd)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	session.Cmd = cmd
	session.pty = pty
	session.PID = cmd.Process.Pid

	go session.readOutput()
	go session.waitProcess()

	return session, nil
}

func (s *ProcessSession) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActive = time.Now()
	return s.pty.Write(data)
}

func (s *ProcessSession) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActive = time.Now()
	return resizePty(s.pty, cols, rows)
}

func (s *ProcessSession) Kill() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done {
		return
	}
	s.done = true

	if s.Cmd != nil && s.Cmd.Process != nil {
		s.Cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-time.After(2 * time.Second):
			s.Cmd.Process.Kill()
		case <-s.DoneChan:
		}
	}

	if s.pty != nil {
		s.pty.Close()
	}

	close(s.DoneChan)
}

func (s *ProcessSession) readOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.LastActive = time.Now()
			data := make([]byte, n)
			copy(data, buf[:n])
			select {
			case s.OutputChan <- data:
			default:
				// 缓冲区满，丢弃旧数据
			}
		}
		if err != nil {
			s.ErrorChan <- err
			return
		}
	}
}

func (s *ProcessSession) waitProcess() {
	err := s.Cmd.Wait()
	if err != nil {
		s.ErrorChan <- err
	}
	s.Kill()
}
```

- [ ] **Step 3: 创建 PTY 处理 pty.go**

```go
package terminal

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

func startPty(cmd *exec.Cmd) (*os.File, error) {
	// 设置进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func resizePty(f *os.File, cols, rows int) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct {
			Rows uint16
			Cols uint16
			X    uint16
			Y    uint16
		}{
			Rows: uint16(rows),
			Cols: uint16(cols),
		})),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func init() {
	// 忽略 SIGCHLD，让子进程自动回收
	signal.Ignore(syscall.SIGCHLD)
}
```

- [ ] **Step 4: 创建进程池测试 pool_test.go**

```go
package terminal

import (
	"context"
	"testing"
	"time"
)

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
```

- [ ] **Step 5: 运行测试**

Run: `cd backend && go test -v ./internal/terminal/...`
Expected: 所有测试通过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/terminal/
git commit -m "feat: add terminal process pool with PTY support"
```

---

### Task 6: API 框架和入口

**Files:**
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/api/handlers/health.go`
- Create: `backend/internal/api/middleware/logger.go`
- Create: `backend/internal/api/middleware/cors.go`
- Create: `backend/cmd/server/main.go`

- [ ] **Step 1: 创建路由 router.go**

```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/api/handlers"
	"github.com/orchestra/backend/internal/api/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", handlers.HealthCheck)

	// API 路由组
	api := r.Group("/api")
	{
		// TODO: 添加业务路由
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

	return r
}
```

- [ ] **Step 2: 创建健康检查 handlers/health.go**

```go
package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Goroutine int    `json:"goroutine"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Version:   "0.1.0",
		GoVersion: runtime.Version(),
		Goroutine: runtime.NumGoroutine(),
	})
}
```

- [ ] **Step 3: 创建日志中间件 middleware/logger.go**

```go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		gin.DefaultWriter.Write([]byte(
			"[API] " + method + " " + path +
				" | " + string(rune(status)) +
				" | " + latency.String() + "\n",
		))
	}
}
```

- [ ] **Step 4: 创建 CORS 中间件 middleware/cors.go**

```go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

- [ ] **Step 5: 创建入口 cmd/server/main.go**

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra/backend/internal/api"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/storage"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	db, err := storage.NewDatabase(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	defer db.Close()

	// 执行迁移
	if err := db.Migrate("internal/storage/migrations"); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	// 启动 HTTP 服务器
	router := api.SetupRouter()

	go func() {
		log.Printf("Starting server on %s", cfg.Server.HTTPAddr)
		if err := router.Run(cfg.Server.HTTPAddr); err != nil {
			log.Fatalf("start server: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	fmt.Println("Goodbye!")
}
```

- [ ] **Step 6: 构建并测试**

Run: `cd backend && go build -o bin/orchestra ./cmd/server && ./bin/orchestra &`
Run: `curl http://localhost:8080/health`
Expected: `{"status":"ok","version":"0.1.0",...}`

- [ ] **Step 7: Commit**

```bash
git add backend/cmd/ backend/internal/api/
git commit -m "feat: add HTTP API framework with health check endpoint"
```

---

### Task 7: WebSocket 网关

**Files:**
- Create: `backend/internal/ws/gateway.go`
- Create: `backend/internal/ws/terminal.go`
- Create: `backend/internal/ws/gateway_test.go`

- [ ] **Step 1: 创建 WebSocket 网关 gateway.go**

```go
package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 生产环境需要验证来源
	},
}

type Gateway struct {
	mu         sync.RWMutex
	terminal   *TerminalHandler
}

func NewGateway(terminal *TerminalHandler) *Gateway {
	return &Gateway{
		terminal: terminal,
	}
}

func (g *Gateway) HandleTerminal(c *gin.Context) {
	sessionID := c.Param("sessionId")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if err := g.terminal.Handle(sessionID, conn); err != nil {
		log.Printf("Terminal handler error: %v", err)
	}
}
```

- [ ] **Step 2: 创建终端 WebSocket 处理 terminal.go**

```go
package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/orchestra/backend/internal/terminal"
)

type TerminalHandler struct {
	pool *terminal.ProcessPool
}

func NewTerminalHandler(pool *terminal.ProcessPool) *TerminalHandler {
	return &TerminalHandler{pool: pool}
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type TerminalResponse struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	Code      int    `json:"code,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

func (h *TerminalHandler) Handle(sessionID string, conn *websocket.Conn) error {
	// 发送连接成功消息
	h.sendMessage(conn, TerminalResponse{
		Type:      "connected",
		SessionID: sessionID,
	})

	session, err := h.pool.Get(sessionID)
	if err != nil {
		h.sendMessage(conn, TerminalResponse{
			Type:    "error",
			Message: err.Error(),
		})
		return err
	}

	// 读取客户端消息
	go h.readLoop(conn, session)

	// 写入终端输出到客户端
	return h.writeLoop(conn, session)
}

func (h *TerminalHandler) readLoop(conn *websocket.Conn, session *terminal.ProcessSession) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v", err)
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Unmarshal message error: %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			session.Write([]byte(msg.Data))
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				session.Resize(msg.Cols, msg.Rows)
			}
		case "close":
			h.pool.Release(session.ID)
			return
		}
	}
}

func (h *TerminalHandler) writeLoop(conn *websocket.Conn, session *terminal.ProcessSession) error {
	for {
		select {
		case data := <-session.OutputChan:
			if err := h.sendMessage(conn, TerminalResponse{
				Type: "output",
				Data: string(data),
			}); err != nil {
				return err
			}
		case err := <-session.ErrorChan:
			h.sendMessage(conn, TerminalResponse{
				Type:    "error",
				Message: err.Error(),
			})
			return err
		case <-session.DoneChan:
			h.sendMessage(conn, TerminalResponse{
				Type: "exit",
				Code: 0,
			})
			return nil
		}
	}
}

func (h *TerminalHandler) sendMessage(conn *websocket.Conn, msg TerminalResponse) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
```

- [ ] **Step 3: 创建 WebSocket 测试 gateway_test.go**

```go
package ws

import (
	"testing"
	"time"

	"github.com/orchestra/backend/internal/terminal"
)

func TestTerminalHandler_Handle(t *testing.T) {
	pool := terminal.NewProcessPool(10, 30*time.Minute)
	handler := NewTerminalHandler(pool)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}
```

- [ ] **Step 4: 更新路由添加 WebSocket**

Update: `backend/internal/api/router.go`

```go
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/api/handlers"
	"github.com/orchestra/backend/internal/api/middleware"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

func SetupRouter(pool *terminal.ProcessPool, gateway *ws.Gateway) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	r.GET("/health", handlers.HealthCheck)

	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

	// WebSocket 路由
	r.GET("/ws/terminal/:sessionId", gateway.HandleTerminal)

	return r
}
```

- [ ] **Step 5: 运行测试**

Run: `cd backend && go test -v ./internal/ws/...`
Expected: 测试通过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/ws/ backend/internal/api/router.go
git commit -m "feat: add WebSocket gateway for terminal streaming"
```

---

### Task 8: 更新入口集成所有模块

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 更新 main.go 集成所有模块**

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra/backend/internal/api"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	db, err := storage.NewDatabase(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate("internal/storage/migrations"); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	// 初始化安全模块
	whitelist := security.NewWhitelist(
		cfg.Security.AllowedCommands,
		cfg.Security.AllowedPaths,
	)

	var encryptor *security.KeyEncryptor
	if cfg.Security.EncryptionKey != "" {
		encryptor, err = security.NewKeyEncryptor(cfg.Security.EncryptionKey)
		if err != nil {
			log.Fatalf("init encryptor: %v", err)
		}
	}

	// 初始化进程池
	pool := terminal.NewProcessPool(
		cfg.Terminal.MaxSessions,
		cfg.Terminal.IdleTimeout,
	)

	// 初始化 WebSocket 网关
	terminalHandler := ws.NewTerminalHandler(pool)
	gateway := ws.NewGateway(terminalHandler)

	// 启动 HTTP 服务器
	router := api.SetupRouter(pool, gateway)

	log.Printf("Orchestra starting on %s", cfg.Server.HTTPAddr)
	log.Printf("Allowed commands: %v", whitelist.AllowedCommands())
	log.Printf("Allowed paths: %v", whitelist.AllowedPaths())

	go func() {
		if err := router.Run(cfg.Server.HTTPAddr); err != nil {
			log.Fatalf("start server: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	_ = encryptor // TODO: 使用加密器
	fmt.Println("Goodbye!")
}
```

- [ ] **Step 2: 创建 ULID 工具 pkg/utils/ulid.go**

```go
package utils

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy = ulid.Monotonic(rand.Reader, 0)

func GenerateID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
```

- [ ] **Step 3: 构建并测试**

Run: `cd backend && make build && ./bin/orchestra &`
Run: `curl http://localhost:8080/health`
Expected: 健康检查返回正常

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/server/main.go backend/pkg/utils/
git commit -m "feat: integrate all modules in main entry point"
```

---

### Task 9: 创建 CLAUDE.md

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: 创建 CLAUDE.md**

```markdown
# Orchestra - 多Agent协作Web系统

## 项目简介

Orchestra 是对照用桌面参考实现的 Web 版复刻，支持多 Agent（Claude Code、Gemini CLI 等）并行执行和编排。

## 技术栈

### 后端
- Go 1.21+
- Gin (HTTP 框架)
- gorilla/websocket
- SQLite (存储)
- PTY (终端模拟)

### 前端
- Vue 3 + TypeScript
- Pinia (状态管理)
- Tailwind CSS
- xterm.js (终端)

## 项目结构

```
Orchestra/
├── backend/           # Go 后端
│   ├── cmd/           # 入口
│   ├── internal/      # 内部模块
│   ├── pkg/           # 公共工具
│   └── configs/       # 配置
├── frontend/          # Vue 前端
├── docs/              # 文档
│   └── superpowers/
│       ├── specs/     # 设计规范
│       └── plans/     # 实现计划
└── CLAUDE.md          # 本文件
```

## 开发命令

### 后端
```bash
cd backend
make build    # 构建
make run      # 运行
make test     # 测试
```

### 前端
```bash
cd frontend
pnpm install  # 安装依赖
pnpm dev      # 开发服务器
pnpm build    # 构建
```

## 代码规范

- Go: gofmt, goimports
- TypeScript: ESLint, Prettier
- 提交信息: Conventional Commits

## API 端点

- `GET /health` - 健康检查
- `GET /api/ping` - 测试连接
- `GET /ws/terminal/:sessionId` - 终端 WebSocket

## 配置

配置文件: `backend/configs/config.yaml`

环境变量:
- `ORCHESTRA_ENCRYPTION_KEY` - API 密钥加密密钥（32字节）
- `ORCHESTRA_CONFIG` - 配置文件路径

## 参考资源

- 设计文档: `docs/superpowers/specs/2026-03-29-orchestra-design.md`
- 对照端参考仓库：本地路径由团队约定
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add CLAUDE.md project documentation"
```

---

## 阶段完成检查

- [ ] 后端项目结构完整
- [ ] 配置管理模块正常工作
- [ ] 数据库初始化和迁移正常
- [ ] 安全模块（加密、白名单）测试通过
- [ ] 进程池模块测试通过
- [ ] HTTP API 框架正常工作
- [ ] WebSocket 网关正常工作
- [ ] 所有测试通过
- [ ] CLAUDE.md 已创建

---

**完成后继续:** Phase 2 - 工作区和成员 API