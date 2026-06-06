package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/outbox"
)

// OutboxHandler exposes dispatch delivery diagnostics to the web client.
type OutboxHandler struct {
	worker *outbox.Worker
}

func NewOutboxHandler(worker *outbox.Worker) *OutboxHandler {
	return &OutboxHandler{worker: worker}
}

func (h *OutboxHandler) ListWorkspace(c *gin.Context) {
	if h.worker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "outbox worker unavailable"})
		return
	}

	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := h.worker.ListWorkspace(c.Request.Context(), workspaceID, outbox.ListFilter{
		Status:         c.Query("status"),
		ConversationID: c.Query("conversationId"),
		Limit:          limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []*outbox.Item{}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
