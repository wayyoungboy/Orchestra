package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

type ConversationHandler struct {
	convRepo   *repository.ConversationRepository
	msgRepo    *repository.MessageRepository
	readRepo   *repository.ConversationReadRepository
	memberRepo repository.MemberRepository
	pool       *terminal.ProcessPool
}

func NewConversationHandler(
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	readRepo *repository.ConversationReadRepository,
	memberRepo repository.MemberRepository,
	pool *terminal.ProcessPool,
) *ConversationHandler {
	return &ConversationHandler{
		convRepo:   convRepo,
		msgRepo:    msgRepo,
		readRepo:   readRepo,
		memberRepo: memberRepo,
		pool:       pool,
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

	if workspaceID != "" && h.pool != nil && h.memberRepo != nil {
		h.forwardUserTextToAssistantPTY(c, workspaceID, convID, req.Text)
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

// MarkConversationRead persists last_read_at to the latest message time in the conversation.
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

// MarkAllConversationsRead marks every conversation in the workspace read for the given member.
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

// SetConversationMembers updates channel membership (full replace).
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

// DeleteConversationsForMember removes DM threads involving the member and drops them from channel member lists.
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

func (h *ConversationHandler) forwardUserTextToAssistantPTY(c *gin.Context, workspaceID, convID, text string) {
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
	// 用户消息转发到带 PTY 的助手与秘书（秘书 = Web 侧监工/协调角色）
	ptyRecipientIDs := make(map[string]struct{})
	for _, m := range members {
		if m.RoleType == models.RoleAssistant || m.RoleType == models.RoleSecretary {
			ptyRecipientIDs[m.ID] = struct{}{}
		}
	}
	if len(ptyRecipientIDs) == 0 {
		return
	}

	targets := make([]string, 0)
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := ptyRecipientIDs[id]; !ok {
			return
		}
		for _, existing := range targets {
			if existing == id {
				return
			}
		}
		targets = append(targets, id)
	}

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
		if len(targets) == 0 && len(conv.MemberIDs) == 0 {
			for id := range ptyRecipientIDs {
				targets = append(targets, id)
			}
		}
	}

	// Strip @mentions from text before sending to terminal (matches golutra behavior)
	cleanText := stripMentions(text, members)

	spanID := ulid.Make().String()
	for _, memberID := range targets {
		sess := h.pool.SessionForWorkspaceMember(workspaceID, memberID)
		if sess == nil {
			continue
		}
		sess.SetLastChatTargetConversation(convID)
		sess.SetStreamSpanID(spanID)
		sess.NoteChatInjectedLine(cleanText)

		// Build prompt with reply instructions
		memberName := sess.MemberName
		if memberName == "" {
			memberName = memberID
		}

		// Format the message with instructions for AI
		// AI should use curl to send reply via internal API
		fullPrompt := fmt.Sprintf(`#conversationId{%s}#senderId{%s}[user]: %s

规则：完成用户的要求。回复时使用 curl 调用内部 API：
curl -X POST http://127.0.0.1:8080/api/internal/chat/send \
  -H "Content-Type: application/json" \
  -d '{"workspaceId":"%s","conversationId":"%s","senderId":"%s","senderName":"%s","text":"你的回复内容"}'`,
			convID,
			memberID,
			cleanText,
			workspaceID,
			convID,
			memberID,
			memberName,
		)

		// Send the message text directly (no ESC - it interferes with TUI input mode)
		_, _ = sess.Write([]byte(fullPrompt))

		// Wait 100ms before sending Enter (matches golutra's COMMAND_CONFIRM_DELAY_MS)
		time.Sleep(100 * time.Millisecond)

		// Send Enter separately to submit the query
		_, _ = sess.Write([]byte{'\r'})
	}
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
// This is used by AI running in PTY to respond to user messages.
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
	if h.pool != nil {
		sess := h.pool.SessionForWorkspaceMember(req.WorkspaceID, req.SenderID)
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

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"messageId": msg.ID,
	})
}