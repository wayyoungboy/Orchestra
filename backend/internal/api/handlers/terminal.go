package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/agent"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
)

type TerminalHandler struct {
	registry *agent.Registry
	wsRepo   repository.WorkspaceRepository
}

func NewTerminalHandler(registry *agent.Registry, wsRepo repository.WorkspaceRepository) *TerminalHandler {
	return &TerminalHandler{registry: registry, wsRepo: wsRepo}
}

// CreateSessionRequest represents the request body for creating a terminal session
type CreateSessionRequest struct {
	Command      string   `json:"command"`
	Args         []string `json:"args"`
	WorkspaceID  string   `json:"workspaceId"`
	MemberID     string   `json:"memberId"`
	MemberName   string   `json:"memberName"`
	TerminalType string   `json:"terminalType"`
}

// CreateSessionResponse represents the response for creating a terminal session
type CreateSessionResponse struct {
	SessionID string `json:"sessionId"`
}

// CreateSession creates a new agent session
// @Summary Create terminal session
// @Description Create a new agent session for an AI assistant
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateSessionRequest true "Session configuration"
// @Success 201 {object} CreateSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/terminals [post]
func (h *TerminalHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get workspace path if needed
	workspacePath := ""
	if req.WorkspaceID != "" {
		ws, err := h.wsRepo.GetByID(c.Request.Context(), req.WorkspaceID)
		if err == nil {
			workspacePath = ws.Path
		}
	}

	member := &models.Member{
		WorkspaceID:  req.WorkspaceID,
		ID:           req.MemberID,
		Name:         req.MemberName,
		TerminalType: req.TerminalType,
		ACPEnabled:   req.Command != "",
		ACPCommand:   req.Command,
		ACPArgs:      req.Args,
	}

	session, err := h.registry.AcquireOrCreate(c.Request.Context(), member, workspacePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent not configured for this member"})
		return
	}

	c.JSON(http.StatusCreated, CreateSessionResponse{
		SessionID: session.ID,
	})
}

// DeleteSession deletes a terminal session
// @Summary Delete terminal session
// @Description Terminate and remove a terminal session
// @Tags terminals
// @Security BearerAuth
// @Param sessionId path string true "Session ID"
// @Success 204 "No Content"
// @Router /api/terminals/{sessionId} [delete]
func (h *TerminalHandler) DeleteSession(c *gin.Context) {
	sess := h.registry.GetByID(c.Param("sessionId"))
	if sess != nil {
		sess.Kill()
		h.registry.Unregister(sess.ID)
	}
	c.JSON(http.StatusNoContent, nil)
}

// ListWorkspaceTerminalSessions lists active sessions for a workspace
// @Summary List workspace terminal sessions
// @Description Get all active agent sessions for a workspace
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/workspaces/{id}/terminal-sessions [get]
func (h *TerminalHandler) ListWorkspaceTerminalSessions(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workspace id"})
		return
	}
	sessions := h.registry.ListSessionsForWorkspace(workspaceID)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// GetSessionForMember gets the session for a specific member
// @Summary Get member's terminal session
// @Description Get the active agent session for a workspace member
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} CreateSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId}/terminal-session [get]
func (h *TerminalHandler) GetSessionForMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")
	if workspaceID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workspace or member"})
		return
	}
	s := h.registry.GetByMember(workspaceID, memberID)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active session"})
		return
	}
	c.JSON(http.StatusOK, CreateSessionResponse{
		SessionID: s.ID,
	})
}

// GetOrCreateSessionForMember gets or creates a session for a member
// @Summary Get or create member's terminal session
// @Description Get existing agent session or create a new one for the member
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Param request body CreateSessionRequest false "Optional session configuration"
// @Success 200 {object} CreateSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId}/terminal-session [post]
func (h *TerminalHandler) GetOrCreateSessionForMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")
	if workspaceID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workspace or member"})
		return
	}

	// Check if session already exists
	s := h.registry.GetByMember(workspaceID, memberID)
	if s != nil {
		c.JSON(http.StatusOK, CreateSessionResponse{
			SessionID: s.ID,
		})
		return
	}

	// Parse optional request body
	var req CreateSessionRequest
	c.ShouldBindJSON(&req) // Ignore error - body is optional

	workspacePath := ""
	if ws, err := h.wsRepo.GetByID(c.Request.Context(), workspaceID); err == nil {
		workspacePath = ws.Path
	}

	member := &models.Member{
		WorkspaceID:  workspaceID,
		ID:           memberID,
		Name:         req.MemberName,
		TerminalType: req.TerminalType,
		ACPEnabled:   req.Command != "",
		ACPCommand:   req.Command,
		ACPArgs:      req.Args,
	}

	session, err := h.registry.AcquireOrCreate(c.Request.Context(), member, workspacePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent not configured for this member"})
		return
	}

	c.JSON(http.StatusCreated, CreateSessionResponse{
		SessionID: session.ID,
	})
}
