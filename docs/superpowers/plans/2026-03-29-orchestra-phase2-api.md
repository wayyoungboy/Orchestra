# Orchestra Phase 2 - 工作区和成员 API 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** 实现工作区管理、成员管理、文件浏览 API。

**Architecture:** REST API + Repository 模式。

**Tech Stack:** Go, Gin, SQLite

---

## 文件结构

```
backend/internal/
├── api/handlers/
│   ├── workspace.go
│   └── member.go
├── storage/repository/
│   ├── interface.go
│   ├── workspace.go
│   └── member.go
├── filesystem/
│   ├── browser.go
│   └── validator.go
└── models/
    ├── workspace.go
    └── member.go
```

---

### Task 1: 数据模型定义

**Files:**
- Create: `backend/internal/models/workspace.go`
- Create: `backend/internal/models/member.go`

- [ ] **Step 1: 创建工作区模型**

```go
// backend/internal/models/workspace.go
package models

import "time"

type Workspace struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	LastOpenedAt  time.Time `json:"lastOpenedAt"`
	CreatedAt     time.Time `json:"createdAt"`
}

type WorkspaceCreate struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"path" binding:"required"`
}
```

- [ ] **Step 2: 创建成员模型**

```go
// backend/internal/models/member.go
package models

import "time"

type MemberRole string

const (
	RoleOwner     MemberRole = "owner"
	RoleAdmin     MemberRole = "admin"
	RoleAssistant MemberRole = "assistant"
	RoleMember    MemberRole = "member"
)

type Member struct {
	ID               string     `json:"id"`
	WorkspaceID      string     `json:"workspaceId"`
	Name             string     `json:"name"`
	RoleType         MemberRole `json:"roleType"`
	RoleKey          string     `json:"roleKey,omitempty"`
	Avatar           string     `json:"avatar,omitempty"`
	TerminalType     string     `json:"terminalType,omitempty"`
	TerminalCommand  string     `json:"terminalCommand,omitempty"`
	TerminalPath     string     `json:"terminalPath,omitempty"`
	AutoStartTerminal bool      `json:"autoStartTerminal"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type MemberCreate struct {
	Name             string     `json:"name" binding:"required"`
	RoleType         MemberRole `json:"roleType" binding:"required"`
	TerminalType     string     `json:"terminalType,omitempty"`
	TerminalCommand  string     `json:"terminalCommand,omitempty"`
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/models/
git commit -m "feat: add workspace and member data models"
```

---

### Task 2: 存储接口和实现

**Files:**
- Create: `backend/internal/storage/repository/interface.go`
- Create: `backend/internal/storage/repository/workspace.go`
- Create: `backend/internal/storage/repository/member.go`

- [ ] **Step 1: 创建存储接口**

```go
// backend/internal/storage/repository/interface.go
package repository

import (
	"context"
	"github.com/orchestra/backend/internal/models"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *models.Workspace) error
	GetByID(ctx context.Context, id string) (*models.Workspace, error)
	List(ctx context.Context) ([]*models.Workspace, error)
	Update(ctx context.Context, ws *models.Workspace) error
	Delete(ctx context.Context, id string) error
	UpdateLastOpened(ctx context.Context, id string) error
}

type MemberRepository interface {
	Create(ctx context.Context, m *models.Member) error
	GetByID(ctx context.Context, id string) (*models.Member, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Member, error)
	Update(ctx context.Context, m *models.Member) error
	Delete(ctx context.Context, id string) error
}
```

- [ ] **Step 2: 创建工作区仓库**

```go
// backend/internal/storage/repository/workspace.go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orchestra/backend/internal/models"
)

type sqlWorkspaceRepo struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &sqlWorkspaceRepo{db: db}
}

func (r *sqlWorkspaceRepo) Create(ctx context.Context, ws *models.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, path, last_opened_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		ws.ID, ws.Name, ws.Path,
		ws.LastOpenedAt.Unix(), ws.CreatedAt.Unix(),
	)
	return err
}

func (r *sqlWorkspaceRepo) GetByID(ctx context.Context, id string) (*models.Workspace, error) {
	query := `
		SELECT id, name, path, last_opened_at, created_at
		FROM workspaces WHERE id = ?
	`
	ws := &models.Workspace{}
	var lastOpened, createdAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ws.ID, &ws.Name, &ws.Path, &lastOpened, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	ws.LastOpenedAt = time.Unix(lastOpened, 0)
	ws.CreatedAt = time.Unix(createdAt, 0)
	return ws, nil
}

func (r *sqlWorkspaceRepo) List(ctx context.Context) ([]*models.Workspace, error) {
	query := `
		SELECT id, name, path, last_opened_at, created_at
		FROM workspaces ORDER BY last_opened_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		ws := &models.Workspace{}
		var lastOpened, createdAt int64
		if err := rows.Scan(
			&ws.ID, &ws.Name, &ws.Path, &lastOpened, &createdAt,
		); err != nil {
			return nil, err
		}
		ws.LastOpenedAt = time.Unix(lastOpened, 0)
		ws.CreatedAt = time.Unix(createdAt, 0)
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

func (r *sqlWorkspaceRepo) Update(ctx context.Context, ws *models.Workspace) error {
	query := `UPDATE workspaces SET name = ?, path = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, ws.Name, ws.Path, ws.ID)
	return err
}

func (r *sqlWorkspaceRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM workspaces WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqlWorkspaceRepo) UpdateLastOpened(ctx context.Context, id string) error {
	query := `UPDATE workspaces SET last_opened_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now().Unix(), id)
	return err
}
```

- [ ] **Step 3: 创建成员仓库**

```go
// backend/internal/storage/repository/member.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/orchestra/backend/internal/models"
)

type sqlMemberRepo struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) MemberRepository {
	return &sqlMemberRepo{db: db}
}

func (r *sqlMemberRepo) Create(ctx context.Context, m *models.Member) error {
	query := `
		INSERT INTO members (id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.WorkspaceID, m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status, m.CreatedAt.Unix(),
	)
	return err
}

func (r *sqlMemberRepo) GetByID(ctx context.Context, id string) (*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at
		FROM members WHERE id = ?
	`
	m := &models.Member{}
	var autoStart int
	var createdAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
		&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	m.AutoStartTerminal = autoStart == 1
	m.CreatedAt = time.Unix(createdAt, 0)
	return m, nil
}

func (r *sqlMemberRepo) ListByWorkspace(ctx context.Context, workspaceID string) ([]*models.Member, error) {
	query := `
		SELECT id, workspace_id, name, role_type, role_key, avatar,
			terminal_type, terminal_command, terminal_path, auto_start_terminal, status, created_at
		FROM members WHERE workspace_id = ? ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.Member
	for rows.Next() {
		m := &models.Member{}
		var autoStart int
		var createdAt int64
		if err := rows.Scan(
			&m.ID, &m.WorkspaceID, &m.Name, &m.RoleType, &m.RoleKey, &m.Avatar,
			&m.TerminalType, &m.TerminalCommand, &m.TerminalPath, &autoStart, &m.Status, &createdAt,
		); err != nil {
			return nil, err
		}
		m.AutoStartTerminal = autoStart == 1
		m.CreatedAt = time.Unix(createdAt, 0)
		members = append(members, m)
	}
	return members, nil
}

func (r *sqlMemberRepo) Update(ctx context.Context, m *models.Member) error {
	query := `
		UPDATE members SET name = ?, role_type = ?, role_key = ?, avatar = ?,
			terminal_type = ?, terminal_command = ?, terminal_path = ?, auto_start_terminal = ?, status = ?
		WHERE id = ?
	`
	autoStart := 0
	if m.AutoStartTerminal {
		autoStart = 1
	}
	_, err := r.db.ExecContext(ctx, query,
		m.Name, m.RoleType, m.RoleKey, m.Avatar,
		m.TerminalType, m.TerminalCommand, m.TerminalPath, autoStart, m.Status, m.ID,
	)
	return err
}

func (r *sqlMemberRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM members WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/storage/repository/
git commit -m "feat: add workspace and member repository implementations"
```

---

### Task 3: 文件系统浏览

**Files:**
- Create: `backend/internal/filesystem/browser.go`
- Create: `backend/internal/filesystem/validator.go`

- [ ] **Step 1: 创建文件浏览器**

```go
// backend/internal/filesystem/browser.go
package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileInfo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	IsDir   bool        `json:"isDir"`
	Size    int64       `json:"size"`
	ModTime time.Time   `json:"modTime"`
	Mode    string      `json:"mode"`
}

type Browser struct {
	validator *Validator
}

func NewBrowser(validator *Validator) *Browser {
	return &Browser{validator: validator}
}

func (b *Browser) ListDir(path string) ([]*FileInfo, error) {
	if err := b.validator.ValidatePath(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []*FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, &FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		})
	}

	// 排序：目录在前，然后按名称
	sortFiles(files)
	return files, nil
}

func (b *Browser) GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (b *Browser) PathExists(path string) (bool, error) {
	if err := b.validator.ValidatePath(path); err != nil {
		return false, err
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func sortFiles(files []*FileInfo) {
	// 简单排序：目录优先
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if !files[i].IsDir && files[j].IsDir {
				files[i], files[j] = files[j], files[i]
			} else if files[i].IsDir == files[j].IsDir &&
				strings.Compare(files[i].Name, files[j].Name) > 0 {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}
```

- [ ] **Step 2: 创建路径验证器**

```go
// backend/internal/filesystem/validator.go
package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathNotAllowed = errors.New("path not allowed")
	ErrPathNotExist   = errors.New("path does not exist")
)

type Validator struct {
	allowedPaths []string
}

func NewValidator(allowedPaths []string) *Validator {
	expanded := make([]string, 0, len(allowedPaths))
	for _, p := range allowedPaths {
		expanded = append(expanded, expandPath(p))
	}
	return &Validator{allowedPaths: expanded}
}

func (v *Validator) ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range v.allowedPaths {
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

func (v *Validator) ValidateExists(path string) error {
	if err := v.ValidatePath(path); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrPathNotExist
	}
	return nil
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

- [ ] **Step 3: Commit**

```bash
git add backend/internal/filesystem/
git commit -m "feat: add filesystem browser with path validation"
```

---

### Task 4: API Handlers

**Files:**
- Create: `backend/internal/api/handlers/workspace.go`
- Create: `backend/internal/api/handlers/member.go`

- [ ] **Step 1: 创建工作区 Handler**

```go
// backend/internal/api/handlers/workspace.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

type WorkspaceHandler struct {
	repo     repository.WorkspaceRepository
	browser  *filesystem.Browser
}

func NewWorkspaceHandler(repo repository.WorkspaceRepository, browser *filesystem.Browser) *WorkspaceHandler {
	return &WorkspaceHandler{repo: repo, browser: browser}
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	workspaces, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workspaces)
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	c.JSON(http.StatusOK, ws)
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req models.WorkspaceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证路径存在
	exists, err := h.browser.PathExists(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path does not exist"})
		return
	}

	ws := &models.Workspace{
		ID:           utils.GenerateID(),
		Name:         req.Name,
		Path:         req.Path,
		LastOpenedAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), ws); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ws)
}

func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) Browse(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	subPath := c.Query("path")
	fullPath := ws.Path
	if subPath != "" {
		fullPath = subPath
	}

	files, err := h.browser.ListDir(fullPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"basePath": fullPath,
		"files":    files,
	})
}

func (h *WorkspaceHandler) BrowseRoot(c *gin.Context) {
	home, err := h.browser.GetHomeDir()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	path := c.Query("path")
	if path == "" {
		path = home
	}

	files, err := h.browser.ListDir(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"basePath": path,
		"home":     home,
		"files":    files,
	})
}
```

- [ ] **Step 2: 创建成员 Handler**

```go
// backend/internal/api/handlers/member.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

type MemberHandler struct {
	repo repository.MemberRepository
}

func NewMemberHandler(repo repository.MemberRepository) *MemberHandler {
	return &MemberHandler{repo: repo}
}

func (h *MemberHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	members, err := h.repo.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *MemberHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")

	var req models.MemberCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &models.Member{
		ID:               utils.GenerateID(),
		WorkspaceID:      workspaceID,
		Name:             req.Name,
		RoleType:         req.RoleType,
		TerminalType:     req.TerminalType,
		TerminalCommand:  req.TerminalCommand,
		AutoStartTerminal: true,
		Status:           "online",
		CreatedAt:        time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func (h *MemberHandler) Update(c *gin.Context) {
	id := c.Param("memberId")

	m, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}

	var req models.Member
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m.Name = req.Name
	m.RoleType = req.RoleType
	m.Avatar = req.Avatar
	m.TerminalType = req.TerminalType
	m.TerminalCommand = req.TerminalCommand
	m.TerminalPath = req.TerminalPath
	m.AutoStartTerminal = req.AutoStartTerminal
	m.Status = req.Status

	if err := h.repo.Update(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

func (h *MemberHandler) Delete(c *gin.Context) {
	id := c.Param("memberId")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
```

- [ ] **Step 3: 更新路由**

```go
// 在 router.go 中添加
func SetupRouter(..., wsHandler *handlers.WorkspaceHandler, memberHandler *handlers.MemberHandler) *gin.Engine {
	// ...

	api := r.Group("/api")
	{
		// 工作区
		api.GET("/workspaces", wsHandler.List)
		api.POST("/workspaces", wsHandler.Create)
		api.GET("/workspaces/:id", wsHandler.Get)
		api.DELETE("/workspaces/:id", wsHandler.Delete)
		api.GET("/workspaces/:id/browse", wsHandler.Browse)
		api.GET("/browse", wsHandler.BrowseRoot)

		// 成员
		api.GET("/workspaces/:id/members", memberHandler.List)
		api.POST("/workspaces/:id/members", memberHandler.Create)
		api.PUT("/workspaces/:id/members/:memberId", memberHandler.Update)
		api.DELETE("/workspaces/:id/members/:memberId", memberHandler.Delete)
	}
	// ...
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/api/handlers/ backend/internal/api/router.go
git commit -m "feat: add workspace and member API handlers"
```

---

## 阶段完成检查

- [ ] 数据模型定义完成
- [ ] 存储仓库实现完成
- [ ] 文件系统浏览功能完成
- [ ] API handlers 完成并测试通过
- [ ] 路由更新完成

---

**完成后继续:** Phase 3 - 前端基础架构