package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/agent"
)

// ACPTerminalMessage is the WebSocket message format for agent sessions.
type ACPTerminalMessage struct {
	Type string `json:"type"`

	// For user_message
	Content string `json:"content,omitempty"`

	// For tool_result
	ToolUseID  string `json:"tool_use_id,omitempty"`
	ToolResult string `json:"tool_result,omitempty"`
	IsError    bool   `json:"is_error,omitempty"`

	// For resize
	Cols int `json:"cols,omitempty"`
	Rows int `json:"rows,omitempty"`
}

// A2ATerminalHandler handles WebSocket connections for agent terminal sessions.
type A2ATerminalHandler struct {
	registry *agent.Registry
}

// NewA2ATerminalHandler creates a new terminal handler.
func NewA2ATerminalHandler(registry *agent.Registry) *A2ATerminalHandler {
	return &A2ATerminalHandler{registry: registry}
}

// Handle handles a WebSocket connection for an agent session.
func (h *A2ATerminalHandler) Handle(sessionID string, conn *websocket.Conn) error {
	sess := h.registry.GetByID(sessionID)
	if sess == nil {
		_ = h.sendMessage(conn, a2a.ACPTerminalResponse{
			Type:  "error",
			Error: "session not found",
		})
		return nil
	}

	transport := sess.Transport()
	if transport == nil {
		_ = h.sendMessage(conn, a2a.ACPTerminalResponse{
			Type:  "error",
			Error: "transport not configured",
		})
		return nil
	}

	if err := h.sendMessage(conn, a2a.ACPTerminalResponse{
		Type:      "connected",
		SessionID: sessionID,
	}); err != nil {
		return err
	}

	go h.readLoop(conn, transport)
	return h.writeLoop(conn, transport)
}

// readLoop reads messages from WebSocket and sends to agent session.
func (h *A2ATerminalHandler) readLoop(conn *websocket.Conn, session *a2a.Session) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[agent-ws] Read message error: %v", err)
			return
		}

		var msg ACPTerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[agent-ws] Unmarshal message error: %v", err)
			continue
		}

		switch msg.Type {
		case "user_message", "input":
			if err := session.SendUserMessage(msg.Content); err != nil {
				log.Printf("[agent-ws] Failed to send user message: %v", err)
			}
		case "raw_input":
			if err := session.SendRawInput(msg.Content); err != nil {
				log.Printf("[agent-ws] Failed to send raw input: %v", err)
			}
		case "tool_result":
			if err := session.SendToolResult(msg.ToolUseID, msg.ToolResult, msg.IsError); err != nil {
				log.Printf("[agent-ws] Failed to send tool result: %v", err)
			}
		case "resize":
			if err := session.Resize(msg.Cols, msg.Rows); err != nil {
				log.Printf("[agent-ws] Failed to resize terminal: %v", err)
			}
		case "close":
			// Find the AgentSession by its transport ID
			h.registry.ReleaseByTransport(session.ID)
			return
		}
	}
}

// writeLoop writes agent output to WebSocket.
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
