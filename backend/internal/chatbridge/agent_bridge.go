package chatbridge

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
)

// AgentBridge processes A2A/ACP messages and creates chat messages.
// It bridges agent output to database chat messages + WebSocket broadcasts.
type AgentBridge struct {
	mu      sync.Mutex
	msgRepo *repository.MessageRepository
	chatHub *ws.ChatHub
}

// NewAgentBridge creates a new agent bridge.
func NewAgentBridge(msgRepo *repository.MessageRepository, chatHub *ws.ChatHub) *AgentBridge {
	return &AgentBridge{
		msgRepo: msgRepo,
		chatHub: chatHub,
	}
}

// OnMessage handles an incoming message from a session.
func (b *AgentBridge) OnMessage(session SessionInterface, msg *a2a.ACPMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch msg.Type {
	case a2a.TypeAssistantMessage:
		b.handleAssistantMessage(session, msg)
	case a2a.TypeResult:
		b.handleResult(session, msg)
	case a2a.TypeError:
		b.handleError(session, msg)
	}
}

// SessionInterface defines the interface for sessions.
type SessionInterface interface {
	LastChatTargetConversation() string
	GetWorkspaceID() string
	GetMemberID() string
	GetMemberName() string
	TrySendChatStream([]byte)
	NextStreamSeq() uint64
	StreamSpanID() string
}

// handleAssistantMessage processes an assistant message and creates a chat message.
func (b *AgentBridge) handleAssistantMessage(session SessionInterface, msg *a2a.ACPMessage) {
	parsed, err := msg.ParseAssistantMessage()
	if err != nil {
		log.Printf("[agent-bridge] Failed to parse assistant message: %v", err)
		return
	}

	convID := session.LastChatTargetConversation()
	if convID == "" {
		log.Printf("[agent-bridge] No conversation bound for assistant message")
		return
	}

	// Create message in database
	chatMsg, err := b.msgRepo.Create(repository.MessageCreate{
		ConversationID: convID,
		SenderID:       session.GetMemberID(),
		Content:        repository.MessageContent{Type: "text", Text: parsed.Content},
		IsAI:           true,
	})
	if err != nil {
		log.Printf("[agent-bridge] Failed to create message: %v", err)
		return
	}

	// Broadcast to WebSocket clients
	if b.chatHub != nil {
		b.chatHub.BroadcastToWorkspace(session.GetWorkspaceID(), ws.ChatEvent{
			Type:           ws.EventNewMessage,
			WorkspaceID:    session.GetWorkspaceID(),
			ConversationID: convID,
			MessageID:      chatMsg.ID,
			SenderID:       session.GetMemberID(),
			SenderName:     session.GetMemberName(),
			Content:        parsed.Content,
			IsAI:           true,
			CreatedAt:      chatMsg.CreatedAt,
		})
	}

	// Also send via terminal chat stream for real-time UI update
	spanID := session.StreamSpanID()
	if spanID != "" {
		streamPayload := map[string]interface{}{
			"type":           "terminal_chat_stream",
			"conversationId": convID,
			"senderId":       session.GetMemberID(),
			"senderName":     session.GetMemberName(),
			"text":           parsed.Content,
			"spanId":         spanID,
			"seq":            session.NextStreamSeq(),
			"isFinal":        false,
		}
		if data, err := json.Marshal(streamPayload); err == nil {
			session.TrySendChatStream(data)
		}
	}

	log.Printf("[agent-bridge] Created chat message %s from agent %s", chatMsg.ID, session.GetMemberName())
}

// handleResult processes a result message (completion notification).
func (b *AgentBridge) handleResult(session SessionInterface, msg *a2a.ACPMessage) {
	parsed, err := msg.ParseResultMessage()
	if err != nil {
		log.Printf("[agent-bridge] Failed to parse result message: %v", err)
		return
	}

	convID := session.LastChatTargetConversation()
	if convID == "" {
		return
	}

	// Send final stream update
	spanID := session.StreamSpanID()
	if spanID != "" {
		streamPayload := map[string]interface{}{
			"type":           "terminal_chat_stream",
			"conversationId": convID,
			"senderId":       session.GetMemberID(),
			"senderName":     session.GetMemberName(),
			"text":           "",
			"spanId":         spanID,
			"seq":            session.NextStreamSeq(),
			"isFinal":        true,
			"message":        parsed.Message,
			"costUsd":        parsed.CostUSD,
		}
		if data, err := json.Marshal(streamPayload); err == nil {
			session.TrySendChatStream(data)
		}
	}

	log.Printf("[agent-bridge] Session %s completed: %s", session.GetMemberName(), parsed.Message)
}

// handleError processes an error message.
func (b *AgentBridge) handleError(session SessionInterface, msg *a2a.ACPMessage) {
	parsed, err := msg.ParseErrorMessage()
	if err != nil {
		log.Printf("[agent-bridge] Failed to parse error message: %v", err)
		return
	}

	log.Printf("[agent-bridge] Error from session %s: %s", session.GetMemberName(), parsed.Error)

	// Send error notification to terminal
	errorPayload := map[string]interface{}{
		"type":  "error",
		"error": parsed.Error,
	}
	if data, err := json.Marshal(errorPayload); err == nil {
		session.TrySendChatStream(data)
	}
}