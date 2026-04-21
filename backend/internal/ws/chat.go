package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ChatEventType defines the type of chat event
type ChatEventType string

const (
	EventNewMessage    ChatEventType = "new_message"
	EventMessageStatus ChatEventType = "message_status"
	EventUnreadSync    ChatEventType = "unread_sync"
	EventTaskStatus    ChatEventType = "task_status"
)

// ChatEvent represents a chat event to be broadcast
type ChatEvent struct {
	Type           ChatEventType `json:"type"`
	WorkspaceID    string        `json:"workspaceId"`
	ConversationID string        `json:"conversationId,omitempty"`
	MessageID      string        `json:"messageId,omitempty"`
	SenderID       string        `json:"senderId,omitempty"`
	SenderName     string        `json:"senderName,omitempty"`
	Content        string        `json:"content,omitempty"`
	CreatedAt      int64         `json:"createdAt,omitempty"`
	IsAI           bool          `json:"isAi,omitempty"`
	Status         string        `json:"status,omitempty"`
	UnreadCount    int           `json:"unreadCount,omitempty"`
}

// ChatClient represents a connected chat WebSocket client
type ChatClient struct {
	ID          string
	WorkspaceID string
	Conn        *websocket.Conn
	Send        chan []byte
	Quit        chan struct{}
}

// ChatHub manages all chat WebSocket connections and broadcasts messages
type ChatHub struct {
	mu            sync.RWMutex
	clients       map[string]*ChatClient       // clientID -> client
	workspaceSubs map[string]map[string]struct{} // workspaceID -> clientIDs
}

// GlobalChatHub is the global chat hub instance
var GlobalChatHub = &ChatHub{
	clients:       make(map[string]*ChatClient),
	workspaceSubs: make(map[string]map[string]struct{}),
}

// Register registers a new client to the hub
func (h *ChatHub) Register(client *ChatClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	if h.workspaceSubs[client.WorkspaceID] == nil {
		h.workspaceSubs[client.WorkspaceID] = make(map[string]struct{})
	}
	h.workspaceSubs[client.WorkspaceID][client.ID] = struct{}{}
	log.Printf("[ChatHub] client %s registered for workspace %s", client.ID, client.WorkspaceID)
}

// Unregister removes a client from the hub
func (h *ChatHub) Unregister(clientID, workspaceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, clientID)
	if subs, ok := h.workspaceSubs[workspaceID]; ok {
		delete(subs, clientID)
		if len(subs) == 0 {
			delete(h.workspaceSubs, workspaceID)
		}
	}
	log.Printf("[ChatHub] client %s unregistered", clientID)
}

// BroadcastToWorkspace broadcasts an event to all clients in a workspace
func (h *ChatHub) BroadcastToWorkspace(workspaceID string, event ChatEvent) {
	h.mu.RLock()
	subs := h.workspaceSubs[workspaceID]
	h.mu.RUnlock()

	if len(subs) == 0 {
		return
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("[ChatHub] failed to marshal event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for cid := range subs {
		if c, ok := h.clients[cid]; ok {
			select {
			case c.Send <- jsonBytes:
			default:
				log.Printf("[ChatHub] client %s channel full, dropped message", cid)
			}
		}
	}
}

// BroadcastRawToWorkspace broadcasts pre-serialized JSON bytes to all clients in a workspace.
func (h *ChatHub) BroadcastRawToWorkspace(workspaceID string, rawJSON []byte) {
	h.mu.RLock()
	subs := h.workspaceSubs[workspaceID]
	h.mu.RUnlock()

	if len(subs) == 0 {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for cid := range subs {
		if c, ok := h.clients[cid]; ok {
			select {
			case c.Send <- rawJSON:
			default:
				log.Printf("[ChatHub] client %s channel full, dropped message", cid)
			}
		}
	}
}

// BroadcastToConversation broadcasts an event to all clients subscribed to a conversation
func (h *ChatHub) BroadcastToConversation(workspaceID, conversationID string, event ChatEvent) {
	event.ConversationID = conversationID
	h.BroadcastToWorkspace(workspaceID, event)
}

// GetClientCount returns the number of connected clients
func (h *ChatHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetWorkspaceClientCount returns the number of clients for a workspace
func (h *ChatHub) GetWorkspaceClientCount(workspaceID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if subs, ok := h.workspaceSubs[workspaceID]; ok {
		return len(subs)
	}
	return 0
}

// ChatHandler handles chat WebSocket connections
type ChatHandler struct {
	Hub *ChatHub
}

// NewChatHandler creates a new chat handler
func NewChatHandler(hub *ChatHub) *ChatHandler {
	return &ChatHandler{Hub: hub}
}

const (
	pingInterval    = 30 * time.Second
	pongWaitTimeout = 60 * time.Second
	writeWait       = 10 * time.Second
)

// Handle handles a chat WebSocket connection
func (h *ChatHandler) Handle(workspaceID string, conn *websocket.Conn) error {
	clientID := generateClientID()
	client := &ChatClient{
		ID:          clientID,
		WorkspaceID: workspaceID,
		Conn:        conn,
		Send:        make(chan []byte, 64),
		Quit:        make(chan struct{}),
	}

	h.Hub.Register(client)
	defer h.Hub.Unregister(clientID, workspaceID)
	defer close(client.Quit)

	// Set read deadline and pong handler
	conn.SetReadDeadline(time.Now().Add(pongWaitTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWaitTimeout))
	})

	// Start ping ticker
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	// Read goroutine - handles incoming messages and keeps connection alive
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				// Signal quit via channel close; only the defer should close it,
				// so just return here to avoid double-close.
				return
			}
		}
	}()

	// Write loop - sends messages and pings
	for {
		select {
		case msg := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return err
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		case <-client.Quit:
			return nil
		}
	}
}

func generateClientID() string {
	return "chat-" + time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}