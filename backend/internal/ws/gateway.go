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
	a2aTerminal    *A2ATerminalHandler
	chat           *ChatHandler
	allowedOrigins []string
	upgrader       websocket.Upgrader
}

func NewGateway(a2aTerminal *A2ATerminalHandler, allowedOrigins []string) *Gateway {
	g := &Gateway{
		a2aTerminal:    a2aTerminal,
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
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:]
			if strings.HasSuffix(originURL.Host, domain) || originURL.Host == domain[1:] {
				return true
			}
			continue
		}

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

	if g.a2aTerminal != nil {
		if err := g.a2aTerminal.Handle(sessionID, conn); err != nil {
			log.Printf("A2A terminal handler error: %v", err)
		}
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
