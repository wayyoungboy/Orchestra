package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
)

type ConversationHandler struct {
	convRepo   *repository.ConversationRepository
	msgRepo    *repository.MessageRepository
	readRepo   *repository.ConversationReadRepository
	memberRepo repository.MemberRepository
	a2aPool    *a2a.Pool
	chatHub    *ws.ChatHub
}

func NewConversationHandler(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	readRepo *repository.ConversationReadRepository,
	memberRepo repository.MemberRepository,
	a2aPool *a2a.Pool,
	chatHub *ws.ChatHub,
) *ConversationHandler {
	return &ConversationHandler{
		convRepo:   convRepo,
		msgRepo:    msgRepo,
		readRepo:   readRepo,
		memberRepo: memberRepo,
		a2aPool:    a2aPool,
		chatHub:    chatHub,
	}
}

type ConversationListResponse struct {
	Pinned          []ConversationDTO `json:"pinned"`
	Timeline        []ConversationDTO `json:"timeline"`
	DefaultChannelID string           `json:"defaultChannelId,omitempty"`
	TotalUnreadCount int              `json:"totalUnreadCount"`
}

type ConversationDTO struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	TargetID    string   `json:"targetId,omitempty"`
	MemberIDs   []string `json:"memberIds"`
	CustomName  string   `json:"customName,omitempty"`
	Pinned      bool     `json:"pinned"`
	Muted       bool     `json:"muted"`
	IsDefault   bool     `json:"isDefault,omitempty"`
	UnreadCount int      `json:"unreadCount"`
}

type MessageDTO struct {
	ID        string                 `json:"id"`
	SenderID  string                 `json:"senderId"`
	Content   map[string]interface{} `json:"content"`
	CreatedAt int64                  `json:"createdAt"`
	IsAI      bool                   `json:"isAi"`
	Status    string                 `json:"status"`
}

type CreateConversationRequest struct {
	Type      string   `json:"type"`
	MemberIDs []string `json:"memberIds"`
	Name      string   `json:"name,omitempty"`
	TargetID  string   `json:"targetId,omitempty"`
}

// SendMessageRequest is the JSON body for POST .../messages.
// Extra fields (conversationType, mentions, timestamp) may be sent by the web client for forward compatibility; only the fields below are persisted today.
// ClientTraceID is optional idempotency / correlation (e.g. UUID); not stored in DB until a migration adds a column.
type SendMessageRequest struct {
	Text           string `json:"text"`
	SenderID       string `json:"senderId"`
	SenderName     string `json:"senderName"`
	ClientTraceID  string `json:"clientTraceId,omitempty"`
}

// List lists all conversations in a workspace
// @Summary List workspace conversations
// @Description Get all conversations (channels and DMs) in a workspace
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param userId query string false "User ID for unread counts"
// @Success 200 {object} ConversationListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations [get]
func (h *ConversationHandler) List(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	conversations, err := h.convRepo.ListByWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-create default "general" channel if no conversations exist
	if len(conversations) == 0 {
		// Create a default channel with empty member IDs (will be joined later)
		defaultConv, err := h.convRepo.Create(workspaceID, repository.ConversationCreate{
			Type:      repository.ConversationTypeChannel,
			MemberIDs: []string{},
			Name:      "general",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		conversations = []repository.Conversation{*defaultConv}
	}

	var pinned, timeline []ConversationDTO
	var defaultChannelID string

	// Initialize empty slices to avoid null in JSON
	pinned = make([]ConversationDTO, 0)
	timeline = make([]ConversationDTO, 0)

	viewerID := c.Query("userId")
	var totalUnread int

	for _, conv := range conversations {
		dto := ConversationDTO{
			ID:        conv.ID,
			Type:      string(conv.Type),
			TargetID:  conv.TargetID,
			MemberIDs: conv.MemberIDs,
			CustomName: conv.Name,
			Pinned:    conv.Pinned,
			Muted:     conv.Muted,
		}

		if viewerID != "" && h.readRepo != nil {
			lastRead, err := h.readRepo.GetLastRead(conv.ID, viewerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			unread, err := h.msgRepo.CountUnreadForViewer(conv.ID, viewerID, lastRead)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			dto.UnreadCount = unread
			totalUnread += unread
		}

		if conv.Name == "general" && conv.Type == repository.ConversationTypeChannel {
			dto.IsDefault = true
			defaultChannelID = conv.ID
		}

		if conv.Pinned {
			pinned = append(pinned, dto)
		} else {
			timeline = append(timeline, dto)
		}
	}

	c.JSON(http.StatusOK, ConversationListResponse{
		Pinned:           pinned,
		Timeline:         timeline,
		DefaultChannelID: defaultChannelID,
		TotalUnreadCount: totalUnread,
	})
}

// Create creates a new conversation
// @Summary Create conversation
// @Description Create a new channel or DM conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param request body CreateConversationRequest true "Conversation data"
// @Success 201 {object} ConversationDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations [post]
func (h *ConversationHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv, err := h.convRepo.Create(workspaceID, repository.ConversationCreate{
		Type:      repository.ConversationType(req.Type),
		MemberIDs: req.MemberIDs,
		Name:      req.Name,
		TargetID:  req.TargetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ConversationDTO{
		ID:        conv.ID,
		Type:      string(conv.Type),
		TargetID:  conv.TargetID,
		MemberIDs: conv.MemberIDs,
		CustomName: conv.Name,
		Pinned:    conv.Pinned,
		Muted:     conv.Muted,
	})
}

// GetMessages gets messages from a conversation
// @Summary Get conversation messages
// @Description Get messages from a conversation with pagination
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param limit query int false "Number of messages to return" default(200)
// @Param before query string false "Get messages before this message ID"
// @Success 200 {array} MessageDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/messages [get]
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	limit := 200
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	beforeID := c.Query("beforeId")

	messages, err := h.msgRepo.ListByConversation(convID, limit, beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		dtos[i] = MessageDTO{
			ID:       msg.ID,
			SenderID: msg.SenderID,
			Content: map[string]interface{}{
				"type": msg.Content.Type,
				"text": msg.Content.Text,
			},
			CreatedAt: msg.CreatedAt,
			IsAI:      msg.IsAI,
			Status:    string(msg.Status),
		}
	}

	c.JSON(http.StatusOK, dtos)
}

// SendMessage sends a message to a conversation
// @Summary Send message
// @Description Send a message to a conversation (broadcasts to WebSocket)
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param request body SendMessageRequest true "Message data"
// @Success 201 {object} MessageDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/messages [post]
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.msgRepo.Create(repository.MessageCreate{
		ConversationID: convID,
		SenderID:       req.SenderID,
		Content: repository.MessageContent{
			Type: "text",
			Text: req.Text,
		},
		IsAI: false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workspaceID := c.Param("id")

	// Broadcast message to all WebSocket clients in the workspace
	ws.GlobalChatHub.BroadcastToWorkspace(workspaceID, ws.ChatEvent{
		Type:           ws.EventNewMessage,
		WorkspaceID:    workspaceID,
		ConversationID: convID,
		MessageID:      msg.ID,
		SenderID:       msg.SenderID,
		SenderName:     req.SenderName,
		Content:        req.Text,
		CreatedAt:      msg.CreatedAt,
		IsAI:           msg.IsAI,
	})

	if workspaceID != "" && h.a2aPool != nil && h.memberRepo != nil {
		h.forwardUserTextToAgent(c, workspaceID, convID, req.Text)
	}

	c.JSON(http.StatusCreated, MessageDTO{
		ID:       msg.ID,
		SenderID: msg.SenderID,
		Content: map[string]interface{}{
			"type": msg.Content.Type,
			"text": msg.Content.Text,
		},
		CreatedAt: msg.CreatedAt,
		IsAI:      msg.IsAI,
		Status:    string(msg.Status),
	})
}

// UpdateSettings updates conversation settings
// @Summary Update conversation settings
// @Description Update pinned, muted, or other conversation settings
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param request body map[string]interface{} true "Settings to update"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId} [put]
func (h *ConversationHandler) UpdateSettings(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.convRepo.Update(convID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a conversation
// @Summary Delete conversation
// @Description Delete a conversation by ID
// @Tags conversations
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId} [delete]
func (h *ConversationHandler) Delete(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	// Delete messages first
	h.msgRepo.DeleteByConversation(convID)

	if err := h.convRepo.Delete(convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ClearMessages clears all messages in a conversation
// @Summary Clear conversation messages
// @Description Delete all messages in a conversation
// @Tags conversations
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/messages [delete]
func (h *ConversationHandler) ClearMessages(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	if err := h.msgRepo.DeleteByConversation(convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// EnsureDefaultConversation creates a default conversation if none exists
func (h *ConversationHandler) EnsureDefaultConversation(workspaceID string, memberIDs []string) (*repository.Conversation, error) {
	return h.convRepo.GetOrCreateDefaultChannel(workspaceID, memberIDs)
}

type markReadBody struct {
	UserID string `json:"userId"`
}

// MarkConversationRead marks a conversation as read
// @Summary Mark conversation as read
// @Description Mark a conversation as read for a user
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param request body map[string]string true "User ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/read [post]
func (h *ConversationHandler) MarkConversationRead(c *gin.Context) {
	convID := c.Param("convId")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}
	var body markReadBody
	if err := c.ShouldBindJSON(&body); err != nil {
		body.UserID = c.Query("userId")
	}
	userID := body.UserID
	if userID == "" {
		userID = c.Query("userId")
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}
	ts, err := h.msgRepo.LatestMessageTime(convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	if err := h.readRepo.Upsert(convID, userID, ts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllConversationsRead marks all conversations in a workspace as read
// @Summary Mark all conversations as read
// @Description Mark all conversations in a workspace as read for a user
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param request body map[string]string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/read-all [post]
func (h *ConversationHandler) MarkAllConversationsRead(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}
	var body markReadBody
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId required"})
		return
	}
	conversations, err := h.convRepo.ListByWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, conv := range conversations {
		ts, err := h.msgRepo.LatestMessageTime(conv.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ts == 0 {
			ts = time.Now().UnixMilli()
		}
		if err := h.readRepo.Upsert(conv.ID, body.UserID, ts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type setConversationMembersBody struct {
	MemberIDs []string `json:"memberIds"`
}

// SetConversationMembers updates channel membership
// @Summary Set conversation members
// @Description Update the members of a conversation (full replace)
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param request body map[string][]string true "Member IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/members [put]
func (h *ConversationHandler) SetConversationMembers(c *gin.Context) {
	workspaceID := c.Param("id")
	convID := c.Param("convId")
	if workspaceID == "" || convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace and conversation id required"})
		return
	}
	conv, err := h.convRepo.GetByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if conv.WorkspaceID != workspaceID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation not in workspace"})
		return
	}
	if conv.Type != repository.ConversationTypeChannel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only channel conversations support member list"})
		return
	}
	var body setConversationMembersBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	members, err := h.memberRepo.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	valid := make(map[string]struct{}, len(members))
	for _, m := range members {
		valid[m.ID] = struct{}{}
	}
	for _, id := range body.MemberIDs {
		if _, ok := valid[id]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown member id: " + id})
			return
		}
	}
	if err := h.convRepo.SetMemberIDs(convID, body.MemberIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.convRepo.GetByID(convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ConversationDTO{
		ID:         updated.ID,
		Type:       string(updated.Type),
		TargetID:   updated.TargetID,
		MemberIDs:  updated.MemberIDs,
		CustomName: updated.Name,
		Pinned:     updated.Pinned,
		Muted:      updated.Muted,
	})
}

func memberSliceContains(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func memberSliceWithout(ids []string, id string) []string {
	out := make([]string, 0, len(ids))
	for _, x := range ids {
		if x != id {
			out = append(out, x)
		}
	}
	return out
}

// DeleteConversationsForMember deletes conversations for a member
// @Summary Delete member conversations
// @Description Remove DM threads and channel membership for a member
// @Tags conversations
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/members/{memberId}/conversations [delete]
func (h *ConversationHandler) DeleteConversationsForMember(c *gin.Context) {
	workspaceID := c.Param("id")
	memberID := c.Param("memberId")
	if workspaceID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace and member id required"})
		return
	}
	conversations, err := h.convRepo.ListByWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, conv := range conversations {
		switch conv.Type {
		case repository.ConversationTypeDM:
			if conv.TargetID == memberID || memberSliceContains(conv.MemberIDs, memberID) {
				_ = h.msgRepo.DeleteByConversation(conv.ID)
				if err := h.convRepo.Delete(conv.ID); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		case repository.ConversationTypeChannel:
			if !memberSliceContains(conv.MemberIDs, memberID) {
				continue
			}
			next := memberSliceWithout(conv.MemberIDs, memberID)
			if err := h.convRepo.SetMemberIDs(conv.ID, next); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ConversationHandler) forwardUserTextToAgent(c *gin.Context, workspaceID, convID, text string) {
	if text == "" {
		return
	}
	conv, err := h.convRepo.GetByID(convID)
	if err != nil || conv == nil {
		return
	}
	members, err := h.memberRepo.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		return
	}
	// Forward user messages to assistants and secretaries (ACP-enabled members)
	acpRecipientIDs := make(map[string]struct{})
	for _, m := range members {
		if m.RoleType == models.RoleAssistant || m.RoleType == models.RoleSecretary {
			acpRecipientIDs[m.ID] = struct{}{}
		}
	}
	if len(acpRecipientIDs) == 0 {
		return
	}

	// Parse @mentions from text
	mentionedIDs := parseMentions(text, members)

	targets := make([]string, 0)
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := acpRecipientIDs[id]; !ok {
			return
		}
		for _, existing := range targets {
			if existing == id {
				return
			}
		}
		targets = append(targets, id)
	}

	// If there are @mentions, only send to mentioned members
	if len(mentionedIDs) > 0 {
		for _, id := range mentionedIDs {
			add(id)
		}
	} else {
		// No @mentions - use original logic based on conversation type
		switch conv.Type {
		case repository.ConversationTypeDM:
			add(conv.TargetID)
			for _, mid := range conv.MemberIDs {
				add(mid)
			}
		case repository.ConversationTypeChannel:
			for _, mid := range conv.MemberIDs {
				add(mid)
			}
			// Channel without @mention: don't forward (align with reference behavior)
		}
	}

	// Strip @mentions from text before sending
	cleanText := stripMentions(text, members)

	for _, memberID := range targets {
		sess := h.a2aPool.SessionForWorkspaceMember(workspaceID, memberID)
		if sess == nil {
			continue
		}
		sess.SetLastChatTargetConversation(convID)

		// Find member for role check
		var member *models.Member
		for _, m := range members {
			if m.ID == memberID {
				member = m
				break
			}
		}

		memberName := sess.MemberName
		if memberName == "" {
			memberName = memberID
		}

		// Build prompt based on role
		var fullPrompt string
		if member != nil && member.RoleType == models.RoleSecretary {
			// Secretary prompt - coordinator role with task management
			fullPrompt = fmt.Sprintf(`#conversationId{%s}#senderId{%s}[user]: %s

你是团队的协调者（秘书）。你的职责：
1. 分析用户需求，拆解为可执行的任务
2. 查询助手负载，智能分配任务
3. 追踪任务进度，协调多个助手协作

【可用工具】

使用 orchestra_workload_list 查询助手负载。
使用 orchestra_task_create 创建任务分配给助手。
使用 orchestra_chat_send 回复用户。

规则：先查询负载，再分配任务。任务完成后助手会汇报，你审核后回复用户。`,
				convID,
				memberID,
				cleanText,
			)
		} else {
			// Regular assistant prompt with task status reporting
			fullPrompt = fmt.Sprintf(`#conversationId{%s}#senderId{%s}[user]: %s

【可用工具】

使用 orchestra_task_start 开始执行任务。
使用 orchestra_task_complete 完成任务汇报。
使用 orchestra_task_fail 报告任务失败。
使用 orchestra_chat_send 回复用户。

规则：完成任务要求。若收到秘书分配的任务(taskId)，开始执行时调用start，完成后调用complete，失败时调用fail。`,
				convID,
				memberID,
				cleanText,
			)
		}

		// Send via ACP protocol
		_ = sess.SendUserMessage(fullPrompt)
	}
}

// parseMentions extracts member IDs from @mentions in text
func parseMentions(text string, members []*models.Member) []string {
	var mentionIDs []string
	for _, m := range members {
		if m == nil {
			continue
		}
		mention := "@" + m.Name
		if idx := findMentionIndex(text, mention); idx != -1 {
			mentionIDs = append(mentionIDs, m.ID)
		}
	}
	return mentionIDs
}

// stripMentions removes @memberName mentions from text before sending to terminal.
func stripMentions(text string, members []*models.Member) string {
	result := text
	for _, m := range members {
		if m == nil {
			continue
		}
		mention := "@" + m.Name
		// Remove all occurrences of the mention
		for {
			idx := findMentionIndex(result, mention)
			if idx == -1 {
				break
			}
			// Remove the mention and any trailing space
			end := idx + len(mention)
			if end < len(result) && result[end] == ' ' {
				end++
			}
			result = result[:idx] + result[end:]
		}
	}
	return strings.TrimSpace(result)
}

func findMentionIndex(text, mention string) int {
	for i := 0; i <= len(text)-len(mention); i++ {
		if text[i:i+len(mention)] == mention {
			// Check if it's a standalone mention (not part of another word)
			if i > 0 && text[i-1] != ' ' && text[i-1] != '\n' {
				continue
			}
			return i
		}
	}
	return -1
}

// InternalChatSendRequest is the request body for internal chat send API.
type InternalChatSendRequest struct {
	ConversationID string `json:"conversationId"`
	WorkspaceID    string `json:"workspaceId"`
	SenderID       string `json:"senderId"`
	SenderName     string `json:"senderName"`
	Text           string `json:"text"`
}

// InternalChatSend provides a simplified API for AI assistants to send messages.
// InternalChatSend handles AI responses from terminal
// @Summary Internal chat send
// @Description Receive AI responses from terminal PTY (internal API)
// @Tags internal
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body InternalChatSendRequest true "AI response data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/internal/chat/send [post]
func (h *ConversationHandler) InternalChatSend(c *gin.Context) {
	var req InternalChatSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ConversationID == "" || req.WorkspaceID == "" || req.SenderID == "" || req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversationId, workspaceId, senderId, and text are required"})
		return
	}

	// Create message
	msg, err := h.msgRepo.Create(repository.MessageCreate{
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		Content: repository.MessageContent{
			Type: "text",
			Text: req.Text,
		},
		IsAI: true, // Mark as AI message
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to all WebSocket clients via ChatHub
	ws.GlobalChatHub.BroadcastToWorkspace(req.WorkspaceID, ws.ChatEvent{
		Type:           ws.EventNewMessage,
		WorkspaceID:    req.WorkspaceID,
		ConversationID: req.ConversationID,
		MessageID:      msg.ID,
		SenderID:       req.SenderID,
		SenderName:     req.SenderName,
		Content:        req.Text,
		CreatedAt:      msg.CreatedAt,
		IsAI:           true,
	})

	// Broadcast to WebSocket clients via terminal session
	if h.a2aPool != nil {
		sess := h.a2aPool.SessionForWorkspaceMember(req.WorkspaceID, req.SenderID)
		if sess != nil {
			// Build terminal_chat_stream payload for WebSocket broadcast
			payload := map[string]interface{}{
				"type":           "terminal_chat_stream",
				"terminalId":     sess.ID,
				"memberId":       req.SenderID,
				"workspaceId":    req.WorkspaceID,
				"conversationId": req.ConversationID,
				"seq":            sess.NextStreamSeq(),
				"timestamp":      msg.CreatedAt,
				"content":        req.Text,
				"source":         "ai",
				"mode":           "final",
				"messageId":      msg.ID,
				"isAi":           true,
			}
			if jsonBytes, err := json.Marshal(payload); err == nil {
				sess.TrySendChatStream(jsonBytes)
			}
		}
	}

	// Check if sender is secretary - auto-forward @mentions to assistants
	senderMember, err := h.memberRepo.GetByID(c.Request.Context(), req.SenderID)
	if err == nil && senderMember != nil && senderMember.RoleType == models.RoleSecretary {
		// Get all workspace members for mention parsing
		allMembers, err := h.memberRepo.ListByWorkspace(c.Request.Context(), req.WorkspaceID)
		if err == nil {
			// Parse @mentions from secretary's response
			mentionedIDs := parseMentions(req.Text, allMembers)
			if len(mentionedIDs) > 0 {
				// Strip @mentions from forwarded text
				cleanText := stripMentions(req.Text, allMembers)

				// Forward to mentioned assistants
				for _, targetID := range mentionedIDs {
					// Find target member
					var targetMember *models.Member
					for _, m := range allMembers {
						if m.ID == targetID {
							targetMember = m
							break
						}
					}

					// Only forward to assistants (not other secretaries or members)
					if targetMember != nil && targetMember.RoleType == models.RoleAssistant {
						targetSess := h.a2aPool.SessionForWorkspaceMember(req.WorkspaceID, targetID)
						if targetSess != nil {
							targetName := targetSess.MemberName
							if targetName == "" {
								targetName = targetID
							}

							// Build forwarded prompt
							forwardPrompt := fmt.Sprintf(`#conversationId{%s}#senderId{%s}[秘书分配任务]: %s

规则：完成秘书分配的任务。完成后可 @秘书 汇报结果。
回复时使用 curl 调用内部 API：
curl -X POST http://127.0.0.1:8080/api/internal/chat/send \
  -H "Content-Type: application/json" \
  -d '{"workspaceId":"%s","conversationId":"%s","senderId":"%s","senderName":"%s","text":"你的回复内容"}'`,
								req.ConversationID,
								targetID,
								cleanText,
								req.WorkspaceID,
								req.ConversationID,
								targetID,
								targetName,
							)

							_ = targetSess.SendUserMessage(forwardPrompt)
						}
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"messageId": msg.ID,
	})
}

// GetConversation gets a single conversation by ID
// @Summary Get conversation
// @Description Get a single conversation by ID
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Success 200 {object} ConversationDTO
// @Failure 404 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId} [get]
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	convID := c.Param("convId")

	conv, err := h.convRepo.GetByID(convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	dto := ConversationDTO{
		ID:         conv.ID,
		Type:       string(conv.Type),
		TargetID:   conv.TargetID,
		MemberIDs:  conv.MemberIDs,
		CustomName: conv.Name,
		Pinned:     conv.Pinned,
		Muted:      conv.Muted,
	}

	c.JSON(http.StatusOK, dto)
}

// DeleteMessage deletes a single message
// @Summary Delete message
// @Description Delete a single message from a conversation
// @Tags conversations
// @Security BearerAuth
// @Param id path string true "Workspace ID"
// @Param convId path string true "Conversation ID"
// @Param messageId path string true "Message ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/workspaces/{id}/conversations/{convId}/messages/{messageId} [delete]
func (h *ConversationHandler) DeleteMessage(c *gin.Context) {
	convID := c.Param("convId")
	msgID := c.Param("messageId")

	// Verify message belongs to this conversation
	msg, err := h.msgRepo.GetByID(msgID)
	if err != nil || msg.ConversationID != convID {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}

	if err := h.msgRepo.Delete(msgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
// UpdateAgentStatus updates an AI agent's activity status
// @Summary Update agent status
// @Description Update an AI agent's current activity status (thinking, reading file, etc.)
// @Tags internal
// @Accept json
// @Produce json
// @Param request body models.AgentStatusUpdate true "Agent status data"
// @Success 200 {object} models.AgentStatus
// @Failure 400 {object} map[string]string
// @Router /api/internal/agent-status [post]
func (h *ConversationHandler) UpdateAgentStatus(c *gin.Context) {
	var req models.AgentStatusUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := models.AgentStatus{
		MemberID:       req.MemberID,
		WorkspaceID:    req.WorkspaceID,
		ConversationID: req.ConversationID,
		Status:         req.Status,
		Message:        req.Message,
		Progress:       req.Progress,
		Timestamp:      time.Now(),
	}

	// TODO: Broadcast to WebSocket clients via chat gateway
	// This would require passing the gateway to the handler

	c.JSON(http.StatusOK, status)
}
