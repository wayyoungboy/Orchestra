package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/orchestra/backend/internal/a2a"
)

// ACPTerminalMessage is the WebSocket message format for agent sessions.
type ACPTerminalMessage struct {
	Type string `json:"type"`

	// For user_message
	Content string `json:"content,omitempty"`

	// For tool_result
	ToolUseID string `json:"tool_use_id,omitempty"`
	ToolResult string `json:"tool_result,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`

	// Legacy resize support (not used, kept for UI compatibility)
	Cols int `json:"cols,omitempty"`
	Rows int `json:"rows,omitempty"`
}

// A2ATerminalHandler handles WebSocket connections for A2A sessions.
// It translates between WebSocket messages and A2A protocol via the A2A Session.
type A2ATerminalHandler struct {
	pool *a2a.Pool
}

// NewA2ATerminalHandler creates a new A2A terminal handler.
func NewA2ATerminalHandler(pool *a2a.Pool) *A2ATerminalHandler {
	return &A2ATerminalHandler{pool: pool}
}

// Handle handles a WebSocket connection for an A2A session.
func (h *A2ATerminalHandler) Handle(sessionID string, conn *websocket.Conn) error {
	session := h.pool.Get(sessionID)
	if session == nil {
		_ = h.sendMessage(conn, a2a.ACPTerminalResponse{
			Type:  "error",
			Error: "session not found",
		})
		return nil
	}

	if err := h.sendMessage(conn, a2a.ACPTerminalResponse{
		Type:      "connected",
		SessionID: sessionID,
	}); err != nil {
		return err
	}

	go h.readLoop(conn, session)
	return h.writeLoop(conn, session)
}

// readLoop reads messages from WebSocket and sends to A2A session.
func (h *A2ATerminalHandler) readLoop(conn *websocket.Conn, session *a2a.Session) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[a2a-ws] Read message error: %v", err)
			return
		}

		var msg ACPTerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[a2a-ws] Unmarshal message error: %v", err)
			continue
		}

		switch msg.Type {
		case "user_message", "input":
			if err := session.SendUserMessage(msg.Content); err != nil {
				log.Printf("[a2a-ws] Failed to send user message: %v", err)
			}
		case "tool_result":
			if err := session.SendToolResult(msg.ToolUseID, msg.ToolResult, msg.IsError); err != nil {
				log.Printf("[a2a-ws] Failed to send tool result: %v", err)
			}
		case "close":
			h.pool.Release(session.ID)
			return
		}
	}
}

// writeLoop writes A2A messages (converted to ACP format) to WebSocket.
func (h *A2ATerminalHandler) writeLoop(conn *websocket.Conn, session *a2a.Session) error {
	streamCh := make(chan []byte, 256)
	session.SetChatStreamSink(streamCh)
	defer session.SetChatStreamSink(nil)

	for {
		select {
		case msg, ok := <-session.OutputChan:
			if !ok {
				return nil
			}
			response := a2a.ConvertACPToWS(msg)
			if response == nil {
				continue
			}
			if err := h.sendMessage(conn, *response); err != nil {
				return err
			}

		case streamJSON := <-streamCh:
			if err := conn.WriteMessage(websocket.TextMessage, streamJSON); err != nil {
				return err
			}

		case err := <-session.ErrorChan:
			h.sendMessage(conn, a2a.ACPTerminalResponse{
				Type:  "error",
				Error: err.Error(),
			})
			return err

		case <-session.DoneChan:
			h.sendMessage(conn, a2a.ACPTerminalResponse{
				Type: "exit",
				Code: 0,
			})
			return nil
		}
	}
}

func (h *A2ATerminalHandler) sendMessage(conn *websocket.Conn, msg a2a.ACPTerminalResponse) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
