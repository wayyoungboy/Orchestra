package api

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/api/handlers"
	"github.com/orchestra/backend/internal/api/middleware"
	"github.com/orchestra/backend/internal/chatbridge"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
)

func SetupRouter(a2aPool *a2a.Pool, gateway *ws.Gateway, db *storage.Database, cfg *config.Config) (*gin.Engine, *a2a.ToolHandler) {
	r := gin.New()

	// Set trusted proxies (local only)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Initialize filesystem validator and browser
	allowedPaths := cfg.Security.AllowedPaths
	if len(allowedPaths) == 0 {
		allowedPaths = []string{"~"}
	}
	validator := filesystem.NewValidator(allowedPaths)
	browser := filesystem.NewBrowser(validator)

	// Initialize repositories
	wsRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	msgRepo := repository.NewMessageRepository(db.DB())
	readRepo := repository.NewConversationReadRepository(db.DB())
	userRepo := repository.NewUserRepository(db.DB())
	attachRepo := repository.NewAttachmentRepository(db.DB())
	taskRepo := repository.NewTaskRepository(db.DB())
	apiKeyRepo := repository.NewAPIKeyRepository(db.DB())

	// Read configured base URL from environment
	baseURL := os.Getenv("ORCHESTRA_BASE_URL")

	// Initialize handlers
	wsHandler := handlers.NewWorkspaceHandler(wsRepo, memberRepo, msgRepo, browser)
	memberHandler := handlers.NewMemberHandler(memberRepo, wsRepo, ws.GlobalChatHub)
	terminalHandler := handlers.NewTerminalHandler(a2aPool, wsRepo)
	convHandler := handlers.NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, wsRepo, a2aPool, ws.GlobalChatHub, cfg.Server.HTTPAddr, cfg.Auth.Enabled, baseURL)
	attachmentHandler := handlers.NewAttachmentHandler(msgRepo, convRepo, attachRepo, cfg.Server.UploadDir)
	taskHandler := handlers.NewTaskHandler(taskRepo, memberRepo)
	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKeyRepo, cfg)
	if err != nil {
		panic("failed to create API key handler: " + err.Error())
	}

	// Initialize JWT config
	var jwtConfig *security.JWTConfig
	if cfg.Auth.JWTSecret != "" {
		jwtConfig = security.NewJWTConfig(cfg.Auth.JWTSecret)
	}

	// Auth handler
	authHandler := handlers.NewAuthHandler(userRepo, jwtConfig, cfg.Auth.Enabled)

	// Middleware
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(gin.Recovery())

	// Auth middleware config
	authConfig := middleware.DefaultAuthConfig(cfg.Auth.JWTSecret)

	// Health check (no auth required)
	r.GET("/health", handlers.HealthCheck)

	// Swagger API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (no auth middleware)
	authGroup := r.Group("/api/auth")
	{
		authGroup.GET("/config", authHandler.GetAuthConfig)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/validate", authHandler.ValidateToken)
		authGroup.GET("/me", middleware.Auth(authConfig), authHandler.GetCurrentUser)

		// Registration route (optional)
		if cfg.Auth.AllowRegistration {
			authGroup.POST("/register", authHandler.Register)
		}
	}

	// API routes with auth
	api := r.Group("/api")
	api.Use(middleware.Auth(authConfig))
	{
		// Workspaces
		api.GET("/workspaces", wsHandler.List)
		api.POST("/workspaces", wsHandler.Create)
		api.POST("/workspaces/validate-path", wsHandler.ValidatePath)
		api.GET("/workspaces/:id", wsHandler.Get)
		api.PUT("/workspaces/:id", wsHandler.Update)
		api.DELETE("/workspaces/:id", wsHandler.Delete)
		api.GET("/workspaces/:id/browse", wsHandler.Browse)
		api.GET("/workspaces/:id/search", wsHandler.Search)
		api.GET("/browse", wsHandler.BrowseRoot)

		// Members
		api.GET("/workspaces/:id/members", memberHandler.List)
		api.GET("/workspaces/:id/members/:memberId", memberHandler.Get)
		api.POST("/workspaces/:id/members", memberHandler.Create)
		api.PUT("/workspaces/:id/members/:memberId", memberHandler.Update)
		api.DELETE("/workspaces/:id/members/:memberId", memberHandler.Delete)
		api.DELETE("/workspaces/:id/members/:memberId/conversations", convHandler.DeleteConversationsForMember)
		api.GET("/workspaces/:id/members/:memberId/terminal-session", terminalHandler.GetSessionForMember)
		api.POST("/workspaces/:id/members/:memberId/terminal-session", terminalHandler.GetOrCreateSessionForMember)
		api.GET("/workspaces/:id/terminal-sessions", terminalHandler.ListWorkspaceTerminalSessions)

		// Member presence
		api.POST("/workspaces/:id/members/:memberId/presence", memberHandler.UpdatePresence)

		// Terminal sessions
		api.POST("/terminals", terminalHandler.CreateSession)
		api.DELETE("/terminals/:sessionId", terminalHandler.DeleteSession)

		// Conversations
		api.GET("/workspaces/:id/conversations", convHandler.List)
		api.GET("/workspaces/:id/conversations/:convId", convHandler.GetConversation)
		api.POST("/workspaces/:id/conversations", convHandler.Create)
		api.PUT("/workspaces/:id/conversations/:convId", convHandler.UpdateSettings)
		api.DELETE("/workspaces/:id/conversations/:convId", convHandler.Delete)
		api.DELETE("/workspaces/:id/conversations/:convId/messages", convHandler.ClearMessages)
		api.DELETE("/workspaces/:id/conversations/:convId/messages/:messageId", convHandler.DeleteMessage)
		api.GET("/workspaces/:id/conversations/:convId/messages", convHandler.GetMessages)
		api.POST("/workspaces/:id/conversations/:convId/messages", convHandler.SendMessage)
		api.POST("/workspaces/:id/conversations/:convId/read", convHandler.MarkConversationRead)
		api.POST("/workspaces/:id/conversations/read-all", convHandler.MarkAllConversationsRead)
		api.PUT("/workspaces/:id/conversations/:convId/members", convHandler.SetConversationMembers)

		// Internal API for AI assistants
		api.POST("/internal/chat/send", convHandler.InternalChatSend)
		api.POST("/internal/agent-status", convHandler.UpdateAgentStatus)

		// Task API for secretary coordination
		api.POST("/internal/tasks/create", taskHandler.CreateTask)
		api.POST("/internal/tasks/assign", taskHandler.AssignTask)
		api.POST("/internal/tasks/start", taskHandler.StartTask)
		api.POST("/internal/tasks/complete", taskHandler.CompleteTask)
		api.POST("/internal/tasks/fail", taskHandler.FailTask)
		api.POST("/internal/tasks/cancel", taskHandler.CancelTask)
		api.GET("/internal/workloads/list", taskHandler.ListWorkloads)

		// Task management (for frontend)
		api.GET("/workspaces/:id/tasks", taskHandler.ListTasks)
		api.GET("/workspaces/:id/tasks/:taskId", taskHandler.GetTask)
		api.GET("/workspaces/:id/tasks/my-tasks", taskHandler.GetMyTasks)
		api.POST("/workspaces/:id/tasks/:taskId/cancel", taskHandler.CancelTask)

		// Attachments
		api.GET("/workspaces/:id/attachments", attachmentHandler.ListAttachments)
		api.POST("/workspaces/:id/conversations/:convId/attachments", attachmentHandler.UploadAttachment)
		api.GET("/workspaces/:id/attachments/:attachmentId", attachmentHandler.DownloadAttachment)
		api.GET("/workspaces/:id/attachments/:attachmentId/info", attachmentHandler.GetAttachmentInfo)
		api.DELETE("/workspaces/:id/attachments/:attachmentId", attachmentHandler.DeleteAttachment)

		// API Keys
		api.GET("/api-keys", apiKeyHandler.List)
		api.GET("/api-keys/provider/:provider", apiKeyHandler.GetByProvider)
		api.POST("/api-keys", apiKeyHandler.Create)
		api.DELETE("/api-keys/:id", apiKeyHandler.Delete)
		api.POST("/api-keys/test", apiKeyHandler.Test)
	}

	// WebSocket routes with WebSocket-specific auth
	wsAuth := middleware.WebSocketAuth(authConfig)
	r.GET("/ws/terminal/:sessionId", wsAuth, gateway.HandleTerminal)
	r.GET("/ws/chat/:workspaceId", wsAuth, gateway.HandleChat)

	// Wire up AgentBridge: A2A output → database chat messages + WebSocket broadcasts
	bridge := chatbridge.NewAgentBridge(msgRepo, ws.GlobalChatHub)
	a2aPool.SetOutputHook(func(sess *a2a.Session, msg *a2a.ACPMessage) {
		bridge.OnMessage(sess, msg)
	})

	// Wire up ToolHandler: agent tool calls → Orchestra operations
	// Wrap ChatHub to satisfy ChatBroadcaster interface (interface{} → ChatEvent)
	chatBroadcaster := &chatEventAdapter{hub: ws.GlobalChatHub}
	toolHandler := a2a.NewToolHandler(msgRepo, taskRepo, memberRepo, convRepo, chatBroadcaster, browser, validator)
	a2aPool.SetToolHandler(toolHandler)

	// Wire up Pool to ToolHandler for task dispatch
	toolHandler.SetPool(a2aPool)
	toolHandler.SetWorkspaceRepo(wsRepo)

	return r, toolHandler
}

// chatEventAdapter wraps ws.ChatHub to satisfy a2a.ChatBroadcaster interface.
// The ToolHandler passes raw maps which this adapter converts to ChatEvent structs.
type chatEventAdapter struct {
	hub *ws.ChatHub
}

func (a *chatEventAdapter) BroadcastToWorkspace(workspaceID string, event interface{}) {
	// If it's already a ChatEvent, pass it through directly
	if ce, ok := event.(ws.ChatEvent); ok {
		a.hub.BroadcastToWorkspace(workspaceID, ce)
		return
	}
	// Otherwise serialize the raw map and re-emit as raw JSON to the workspace
	if raw, ok := event.(map[string]interface{}); ok {
		// Add "type" field as ChatEventType for frontend parsing
		if eventType, hasType := raw["type"].(string); hasType {
			raw["type"] = ws.ChatEventType(eventType)
		}
		jsonBytes, err := json.Marshal(raw)
		if err != nil {
			return
		}
		a.hub.BroadcastRawToWorkspace(workspaceID, jsonBytes)
	}
}