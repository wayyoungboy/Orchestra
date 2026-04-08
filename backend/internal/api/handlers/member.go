package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
	"github.com/orchestra/backend/pkg/utils"
)

type MemberHandler struct {
	repo    repository.MemberRepository
	wsRepo  repository.WorkspaceRepository
	chatHub *ws.ChatHub
}

func NewMemberHandler(repo repository.MemberRepository, wsRepo repository.WorkspaceRepository, chatHub *ws.ChatHub) *MemberHandler {
	return &MemberHandler{repo: repo, wsRepo: wsRepo, chatHub: chatHub}
}

// List lists all members in a workspace
// @Summary List workspace members
// @Description Get all members in a workspace
// @Tags members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Success 200 {array} models.Member
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members [get]
func (h *MemberHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	ws, err := h.wsRepo.GetByID(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = ensureWorkspaceOwner(c.Request.Context(), h.repo, ws, "")

	members, err := h.repo.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

// Get gets a single member by ID
// @Summary Get workspace member
// @Description Get a single member by ID
// @Tags members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} models.Member
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId} [get]
func (h *MemberHandler) Get(c *gin.Context) {
	memberID := c.Param("memberId")

	member, err := h.repo.GetByID(c.Request.Context(), memberID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// Create creates a new member in a workspace
// @Summary Create workspace member
// @Description Create a new member (AI assistant or team member) in a workspace
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param request body models.MemberCreate true "Member data"
// @Success 201 {object} models.Member
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members [post]
func (h *MemberHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")

	if _, err := h.wsRepo.GetByID(c.Request.Context(), workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.MemberCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	switch req.RoleType {
	case models.RoleOwner, models.RoleAdmin, models.RoleSecretary, models.RoleAssistant, models.RoleMember:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roleType"})
		return
	}

	m := &models.Member{
		ID:                utils.GenerateID(),
		WorkspaceID:       workspaceID,
		Name:              req.Name,
		RoleType:          req.RoleType,
		TerminalType:      req.TerminalType,
		TerminalCommand:   req.TerminalCommand,
		AutoStartTerminal: true,
		Status:            "online",
		CreatedAt:         time.Now(),
		ACPEnabled:        req.ACPEnabled,
		ACPCommand:        req.ACPCommand,
		ACPArgs:           req.ACPArgs,
	}

	if err := h.repo.Create(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

// Update updates a member
// @Summary Update workspace member
// @Description Update a member's properties (partial update supported)
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Param request body map[string]interface{} true "Member patch data"
// @Success 200 {object} models.Member
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId} [put]
func (h *MemberHandler) Update(c *gin.Context) {
	id := c.Param("memberId")

	m, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		return
	}

	// JSON patch: only keys present are applied
	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if v, ok := patch["name"].(string); ok {
		m.Name = v
	}
	if v, ok := patch["roleType"].(string); ok {
		m.RoleType = models.MemberRole(v)
	}
	if v, ok := patch["avatar"].(string); ok {
		m.Avatar = v
	}
	if v, ok := patch["terminalType"].(string); ok {
		m.TerminalType = v
	}
	if v, ok := patch["terminalCommand"].(string); ok {
		m.TerminalCommand = v
	}
	if v, ok := patch["terminalPath"].(string); ok {
		m.TerminalPath = v
	}
	if v, ok := patch["autoStartTerminal"].(bool); ok {
		m.AutoStartTerminal = v
	}
	if v, ok := patch["status"].(string); ok {
		m.Status = v
	} else if v, ok := patch["manualStatus"].(string); ok {
		m.Status = v
	}
	// ACP fields
	if v, ok := patch["acpEnabled"].(bool); ok {
		m.ACPEnabled = v
	}
	if v, ok := patch["acpCommand"].(string); ok {
		m.ACPCommand = v
	}
	if v, ok := patch["acpArgs"].([]interface{}); ok {
		args := make([]string, 0, len(v))
		for _, a := range v {
			if s, ok := a.(string); ok {
				args = append(args, s)
			}
		}
		m.ACPArgs = args
	}
	// A2A fields
	if v, ok := patch["a2aEnabled"].(bool); ok {
		m.A2AEnabled = v
	}
	if v, ok := patch["a2aAgentURL"].(string); ok {
		if v != "" {
			ptr := v
			m.A2AAgentURL = &ptr
		} else {
			m.A2AAgentURL = nil
		}
	}
	if v, ok := patch["a2aAuthType"].(string); ok {
		if v != "" {
			ptr := v
			m.A2AAuthType = &ptr
		} else {
			m.A2AAuthType = nil
		}
	}
	if v, ok := patch["a2aAuthToken"].(string); ok {
		if v != "" {
			ptr := v
			m.A2AAuthToken = &ptr
		} else {
			m.A2AAuthToken = nil
		}
	}

	if err := h.repo.Update(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

// Delete deletes a member
// @Summary Delete workspace member
// @Description Delete a member from a workspace
// @Tags members
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId} [delete]
func (h *MemberHandler) Delete(c *gin.Context) {
	id := c.Param("memberId")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
// UpdatePresence updates a member's real-time presence status
// @Summary Update member presence
// @Description Update a member's real-time activity status (typing, viewing, etc.)
// @Tags members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Param request body models.PresenceUpdate true "Presence data"
// @Success 200 {object} models.Presence
// @Router /api/workspaces/{id}/members/{memberId}/presence [post]
func (h *MemberHandler) UpdatePresence(c *gin.Context) {
	memberID := c.Param("memberId")
	workspaceID := c.Param("id")

	var req models.PresenceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	presence := models.Presence{
		MemberID:    memberID,
		WorkspaceID: workspaceID,
		Activity:    req.Activity,
		TargetID:    req.TargetID,
		TargetType:  req.TargetType,
		Timestamp:   time.Now().UnixMilli(),
	}

	// Broadcast to WebSocket clients
	if h.chatHub != nil {
		h.chatHub.BroadcastToWorkspace(workspaceID, ws.ChatEvent{
			Type:        ws.EventNewMessage, // Reuse new_message type with presence metadata
			WorkspaceID: workspaceID,
			SenderID:    memberID,
			Content:     req.Activity,
			Status:      req.TargetType + ":" + req.TargetID,
			CreatedAt:   presence.Timestamp,
		})
	}

	c.JSON(http.StatusOK, presence)
}
