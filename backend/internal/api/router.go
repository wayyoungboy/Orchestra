package api

import (
	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/api/handlers"
	"github.com/orchestra/backend/internal/api/middleware"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

func SetupRouter(pool *terminal.ProcessPool, gateway *ws.Gateway, db *storage.Database, cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Set trusted proxies (local only)
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Initialize filesystem validator and browser
	validator := filesystem.NewValidator([]string{"~"})
	browser := filesystem.NewBrowser(validator)

	// Initialize repositories
	wsRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	msgRepo := repository.NewMessageRepository(db.DB())
	readRepo := repository.NewConversationReadRepository(db.DB())
	userRepo := repository.NewUserRepository(db.DB())

	// Initialize handlers
	wsHandler := handlers.NewWorkspaceHandler(wsRepo, memberRepo, browser)
	memberHandler := handlers.NewMemberHandler(memberRepo, wsRepo)
	terminalHandler := handlers.NewTerminalHandler(pool, wsRepo)
	convHandler := handlers.NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, pool)

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
		api.GET("/workspaces/:id", wsHandler.Get)
		api.DELETE("/workspaces/:id", wsHandler.Delete)
		api.GET("/workspaces/:id/browse", wsHandler.Browse)
		api.GET("/browse", wsHandler.BrowseRoot)

		// Members
		api.GET("/workspaces/:id/members", memberHandler.List)
		api.POST("/workspaces/:id/members", memberHandler.Create)
		api.PUT("/workspaces/:id/members/:memberId", memberHandler.Update)
		api.DELETE("/workspaces/:id/members/:memberId", memberHandler.Delete)
		api.DELETE("/workspaces/:id/members/:memberId/conversations", convHandler.DeleteConversationsForMember)
		api.GET("/workspaces/:id/members/:memberId/terminal-session", terminalHandler.GetSessionForMember)
		api.GET("/workspaces/:id/terminal-sessions", terminalHandler.ListWorkspaceTerminalSessions)

		// Terminal sessions
		api.POST("/terminals", terminalHandler.CreateSession)
		api.DELETE("/terminals/:sessionId", terminalHandler.DeleteSession)

		// Conversations
		api.GET("/workspaces/:id/conversations", convHandler.List)
		api.POST("/workspaces/:id/conversations", convHandler.Create)
		api.PUT("/workspaces/:id/conversations/:convId", convHandler.UpdateSettings)
		api.DELETE("/workspaces/:id/conversations/:convId", convHandler.Delete)
		api.DELETE("/workspaces/:id/conversations/:convId/messages", convHandler.ClearMessages)
		api.PUT("/workspaces/:id/conversations/:convId/settings", convHandler.UpdateSettings)
		api.GET("/workspaces/:id/conversations/:convId/messages", convHandler.GetMessages)
		api.POST("/workspaces/:id/conversations/:convId/messages", convHandler.SendMessage)
		api.POST("/workspaces/:id/conversations/:convId/read", convHandler.MarkConversationRead)
		api.POST("/workspaces/:id/conversations/read-all", convHandler.MarkAllConversationsRead)
		api.PUT("/workspaces/:id/conversations/:convId/members", convHandler.SetConversationMembers)

		// Internal API for AI assistants
		api.POST("/internal/chat/send", convHandler.InternalChatSend)
	}

	// WebSocket routes with WebSocket-specific auth
	wsAuth := middleware.WebSocketAuth(authConfig)
	r.GET("/ws/terminal/:sessionId", wsAuth, gateway.HandleTerminal)
	r.GET("/ws/chat/:workspaceId", wsAuth, gateway.HandleChat)

	return r
}