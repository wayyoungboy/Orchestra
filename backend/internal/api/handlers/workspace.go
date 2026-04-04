package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/pkg/utils"
)

type WorkspaceHandler struct {
	repo       repository.WorkspaceRepository
	memberRepo repository.MemberRepository
	browser    *filesystem.Browser
}

func NewWorkspaceHandler(repo repository.WorkspaceRepository, memberRepo repository.MemberRepository, browser *filesystem.Browser) *WorkspaceHandler {
	return &WorkspaceHandler{repo: repo, memberRepo: memberRepo, browser: browser}
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	workspaces, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if workspaces == nil {
		workspaces = []*models.Workspace{}
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
	_ = ensureWorkspaceOwner(c.Request.Context(), h.memberRepo, ws, "")
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

	ownerName := req.OwnerDisplayName
	if ownerName == "" {
		ownerName = "Owner"
	}
	if err := ensureWorkspaceOwner(c.Request.Context(), h.memberRepo, ws, ownerName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ws)
}

// ensureWorkspaceOwner creates a single owner member if the workspace has none (covers legacy DBs).
func ensureWorkspaceOwner(ctx context.Context, memberRepo repository.MemberRepository, ws *models.Workspace, displayName string) error {
	members, err := memberRepo.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.RoleType == models.RoleOwner {
			return nil
		}
	}
	name := displayName
	if name == "" {
		name = "Owner"
	}
	m := &models.Member{
		ID:                utils.GenerateID(),
		WorkspaceID:       ws.ID,
		Name:              name,
		RoleType:          models.RoleOwner,
		AutoStartTerminal: false,
		Status:            "online",
		CreatedAt:         time.Now(),
	}
	return memberRepo.Create(ctx, m)
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

	files, err := h.browser.ListDir(fullPath, true)
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

	files, err := h.browser.ListDir(path, true)
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