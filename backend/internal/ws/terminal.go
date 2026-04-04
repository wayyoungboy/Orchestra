package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/orchestra/backend/internal/terminal"
)

const ptyReplayChunkBytes = 32768

func chunkPTYForReplay(b []byte, max int) [][]byte {
	if max <= 0 {
		max = ptyReplayChunkBytes
	}
	if len(b) == 0 {
		return nil
	}
	var parts [][]byte
	for offset := 0; offset < len(b); offset += max {
		end := offset + max
		if end > len(b) {
			end = len(b)
		}
		part := make([]byte, end-offset)
		copy(part, b[offset:end])
		parts = append(parts, part)
	}
	return parts
}

type TerminalHandler struct {
	pool *terminal.ProcessPool
}

func NewTerminalHandler(pool *terminal.ProcessPool) *TerminalHandler {
	return &TerminalHandler{pool: pool}
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type TerminalResponse struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	Code      int    `json:"code,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

func (h *TerminalHandler) Handle(sessionID string, conn *websocket.Conn) error {
	session, err := h.pool.Get(sessionID)
	if err != nil {
		_ = h.sendMessage(conn, TerminalResponse{
			Type:    "error",
			Message: err.Error(),
		})
		return err
	}

	if err := h.sendMessage(conn, TerminalResponse{
		Type:      "connected",
		SessionID: sessionID,
	}); err != nil {
		return err
	}

	// Replay scrollback so late WebSocket attaches see prior PTY output (not only new chunks).
	for _, part := range chunkPTYForReplay(session.SnapshotTerminalScrollback(), ptyReplayChunkBytes) {
		if err := h.sendMessage(conn, TerminalResponse{
			Type: "output",
			Data: string(part),
		}); err != nil {
			return err
		}
	}

	go h.readLoop(conn, session)
	return h.writeLoop(conn, session)
}

func (h *TerminalHandler) readLoop(conn *websocket.Conn, session *terminal.ProcessSession) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read message error: %v", err)
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Unmarshal message error: %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			session.Write([]byte(msg.Data))
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				session.Resize(msg.Cols, msg.Rows)
			}
		case "close":
			h.pool.Release(session.ID)
			return
		}
	}
}

func (h *TerminalHandler) writeLoop(conn *websocket.Conn, session *terminal.ProcessSession) error {
	streamCh := make(chan []byte, 256)
	session.SetChatStreamSink(streamCh)
	defer session.SetChatStreamSink(nil)

	for {
		select {
		case data := <-session.OutputChan:
			if err := h.sendMessage(conn, TerminalResponse{
				Type: "output",
				Data: string(data),
			}); err != nil {
				return err
			}
		case streamJSON := <-streamCh:
			if err := conn.WriteMessage(websocket.TextMessage, streamJSON); err != nil {
				return err
			}
		case err := <-session.ErrorChan:
			h.sendMessage(conn, TerminalResponse{
				Type:    "error",
				Message: err.Error(),
			})
			return err
		case <-session.DoneChan:
			h.sendMessage(conn, TerminalResponse{
				Type: "exit",
				Code: 0,
			})
			return nil
		}
	}
}

func (h *TerminalHandler) sendMessage(conn *websocket.Conn, msg TerminalResponse) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}