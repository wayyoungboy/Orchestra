package ws

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Gateway struct {
	mu             sync.RWMutex
	terminal       *TerminalHandler
	chat           *ChatHandler
	allowedOrigins []string
	upgrader       websocket.Upgrader
}

func NewGateway(terminal *TerminalHandler, allowedOrigins []string) *Gateway {
	g := &Gateway{
		terminal:       terminal,
		chat:           NewChatHandler(GlobalChatHub),
		allowedOrigins: allowedOrigins,
	}
	g.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     g.checkOrigin,
	}
	return g
}

func (g *Gateway) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	for _, allowed := range g.allowedOrigins {
		// Support wildcard subdomain matching (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(originURL.Host, domain) || originURL.Host == domain[1:] {
				return true
			}
			continue
		}

		// Exact match
		if origin == allowed {
			return true
		}
	}

	return false
}

func (g *Gateway) HandleTerminal(c *gin.Context) {
	sessionID := c.Param("sessionId")

	conn, err := g.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if err := g.terminal.Handle(sessionID, conn); err != nil {
		log.Printf("Terminal handler error: %v", err)
	}
}

// HandleChat handles chat WebSocket connections
func (g *Gateway) HandleChat(c *gin.Context) {
	workspaceID := c.Param("workspaceId")
	if workspaceID == "" {
		workspaceID = c.Query("workspace")
	}
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace id required"})
		return
	}

	conn, err := g.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Chat WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if err := g.chat.Handle(workspaceID, conn); err != nil {
		log.Printf("Chat handler error: %v", err)
	}
}