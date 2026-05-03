package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
)

// NotificationHandler handles notification-related HTTP requests.
type NotificationHandler struct {
	notifRepo *repository.NotificationRepository
	chatHub   *ws.ChatHub
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(notifRepo *repository.NotificationRepository, chatHub *ws.ChatHub) *NotificationHandler {
	return &NotificationHandler{
		notifRepo: notifRepo,
		chatHub:   chatHub,
	}
}

// List lists notifications for a user.
// @Summary List notifications
// @Description Get notifications for a user in a workspace
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param userId query string true "User ID"
// @Param limit query int false "Number of notifications to return" default(50)
// @Success 200 {array} repository.Notification
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	notifications, err := h.notifRepo.ListByUser(c.Request.Context(), workspaceID, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// BadgeCounts returns notification badge counts for a user.
// @Summary Get badge counts
// @Description Get total and unread notification counts for a user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param userId query string true "User ID"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/notifications/badge [get]
func (h *NotificationHandler) BadgeCounts(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	userID := c.Query("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}

	total, unread, err := h.notifRepo.BadgeCounts(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  total,
		"unread": unread,
	})
}

// MarkRead marks a single notification as read.
// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param notifId path string true "Notification ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/notifications/{notifId}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	workspaceID := c.Param("id")
	notifID := c.Param("notifId")
	if workspaceID == "" || notifID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id and notification id required"})
		return
	}

	if err := h.notifRepo.MarkRead(c.Request.Context(), notifID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type markAllReadBody struct {
	UserID string `json:"userId"`
}

// MarkAllRead marks all notifications as read for a user.
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for a user in a workspace
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param request body markAllReadBody true "User ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/notifications/read-all [post]
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	var body markAllReadBody
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required in body"})
		return
	}

	if err := h.notifRepo.MarkAllRead(c.Request.Context(), workspaceID, body.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
