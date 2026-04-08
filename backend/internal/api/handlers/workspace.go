package handlers

import (
	"context"
	"net/http"
	"strconv"
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
	msgRepo    *repository.MessageRepository
	browser    *filesystem.Browser
}

func NewWorkspaceHandler(repo repository.WorkspaceRepository, memberRepo repository.MemberRepository, msgRepo *repository.MessageRepository, browser *filesystem.Browser) *WorkspaceHandler {
	return &WorkspaceHandler{repo: repo, memberRepo: memberRepo, msgRepo: msgRepo, browser: browser}
}

// List lists all workspaces
// @Summary List workspaces
// @Description Get all workspaces
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Workspace
// @Failure 500 {object} map[string]string
// @Router /api/workspaces [get]
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

// Get gets a workspace by ID
// @Summary Get workspace
// @Description Get a workspace by ID
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Success 200 {object} models.Workspace
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id} [get]
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

// Create creates a new workspace
// @Summary Create workspace
// @Description Create a new workspace
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.WorkspaceCreate true "Workspace data"
// @Success 201 {object} models.Workspace
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces [post]
func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req models.WorkspaceCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify path exists
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

// ensureWorkspaceOwner creates a single owner member if the workspace has none
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

// Delete deletes a workspace
// @Summary Delete workspace
// @Description Delete a workspace by ID
// @Tags workspaces
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id} [delete]
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Update updates a workspace
// @Summary Update workspace
// @Description Update workspace name or other properties
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param request body models.WorkspaceUpdate true "Workspace update data"
// @Success 200 {object} models.Workspace
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id} [put]
func (h *WorkspaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	ws, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	var req models.WorkspaceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates
	if req.Name != "" {
		ws.Name = req.Name
	}
	if req.Path != "" {
		// Verify new path exists
		exists, err := h.browser.PathExists(req.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path does not exist"})
			return
		}
		ws.Path = req.Path
	}

	if err := h.repo.Update(c.Request.Context(), ws); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ws)
}

// Browse browses files in a workspace
// @Summary Browse workspace files
// @Description Browse files in a workspace directory
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param path query string false "Sub path to browse"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/browse [get]
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

// BrowseRoot browses files from root/home directory
// @Summary Browse root files
// @Description Browse files from home directory or specified path
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param path query string false "Path to browse"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/browse [get]
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

// Search searches messages across all conversations in a workspace
// @Summary Search workspace messages
// @Description Full-text search across all conversations in a workspace
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results" default(50)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/search [get]
func (h *WorkspaceHandler) Search(c *gin.Context) {
	workspaceID := c.Param("id")
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	// Verify workspace exists
	_, err := h.repo.GetByID(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	// Parse limit
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Search messages
	results, err := h.msgRepo.SearchInWorkspace(workspaceID, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}
// ValidatePath validates a path for workspace creation
// @Summary Validate workspace path
// @Description Validate if a path is suitable for a workspace (exists, readable, writable)
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Path to validate"
// @Success 200 {object} filesystem.PathValidationResult
// @Router /api/workspaces/validate-path [post]
func (h *WorkspaceHandler) ValidatePath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	result := h.browser.ValidatePath(req.Path)
	c.JSON(http.StatusOK, result)
}
