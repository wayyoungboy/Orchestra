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
	"github.com/orchestra/backend/pkg/utils"
)

type MemberHandler struct {
	repo   repository.MemberRepository
	wsRepo repository.WorkspaceRepository
}

func NewMemberHandler(repo repository.MemberRepository, wsRepo repository.WorkspaceRepository) *MemberHandler {
	return &MemberHandler{repo: repo, wsRepo: wsRepo}
}

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

	// JSON patch: only keys present are applied (matches frontend Partial<Member>).
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