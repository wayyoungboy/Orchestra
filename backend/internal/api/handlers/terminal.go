package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
)

type TerminalHandler struct {
	pool   *terminal.ProcessPool
	wsRepo repository.WorkspaceRepository
}

func NewTerminalHandler(pool *terminal.ProcessPool, wsRepo repository.WorkspaceRepository) *TerminalHandler {
	return &TerminalHandler{pool: pool, wsRepo: wsRepo}
}

type CreateSessionRequest struct {
	Command      string   `json:"command"`
	Args         []string `json:"args"`
	WorkspaceID  string   `json:"workspaceId"`
	MemberID     string   `json:"memberId"`
	TerminalType string   `json:"terminalType"`
	MemberName   string   `json:"memberName"` // For introduction prompt
}

type CreateSessionResponse struct {
	SessionID string `json:"sessionId"`
	PID       int    `json:"pid"`
}

func (h *TerminalHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default command if not specified
	command := req.Command
	if command == "" {
		command = "/bin/zsh"
	}

	// Get workspace path if provided
	workspacePath := ""
	if req.WorkspaceID != "" {
		ws, err := h.wsRepo.GetByID(c.Request.Context(), req.WorkspaceID)
		if err == nil {
			workspacePath = ws.Path
		}
	}

	// Set environment variables for AI assistants
	env := []string{
		fmt.Sprintf("ORCHESTRA_WORKSPACE_ID=%s", req.WorkspaceID),
		fmt.Sprintf("ORCHESTRA_MEMBER_ID=%s", req.MemberID),
		fmt.Sprintf("ORCHESTRA_MEMBER_NAME=%s", req.MemberName),
	}

	config := terminal.ProcessConfig{
		Command:      command,
		Args:         req.Args,
		Workspace:    workspacePath,
		WorkspaceID:  req.WorkspaceID,
		MemberID:     req.MemberID,
		MemberName:   req.MemberName,
		TerminalType: req.TerminalType,
		Env:          env,
	}

	session, err := h.pool.Acquire(c.Request.Context(), config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send introduction prompt for AI assistants after a short delay
	// Disabled: the prompt causes AI to output system help instead of waiting for user input
	// The member name is already set via the launch context, no need to send a prompt
	_ = req.MemberName // Silence unused variable warning

	c.JSON(http.StatusCreated, CreateSessionResponse{
		SessionID: session.ID,
		PID:       session.PID,
	})
}

func (h *TerminalHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	h.pool.Release(sessionID)
	c.JSON(http.StatusNoContent, nil)
}

// ListWorkspaceTerminalSessions lists active server-side PTY sessions for a workspace (polling / REQ-303).
func (h *TerminalHandler) ListWorkspaceTerminalSessions(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workspace id"})
		return
	}
	sessions := h.pool.ListSessionsForWorkspace(workspaceID)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// GetSessionForMember returns an existing in-pool terminal for workspace+member (terminal_attach 等价：按成员恢复会话 id)。
func (h *TerminalHandler) GetSessionForMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")
	if workspaceID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing workspace or member"})
		return
	}
	s := h.pool.SessionForWorkspaceMember(workspaceID, memberID)
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active terminal session"})
		return
	}
	c.JSON(http.StatusOK, CreateSessionResponse{
		SessionID: s.ID,
		PID:       s.PID,
	})
}
