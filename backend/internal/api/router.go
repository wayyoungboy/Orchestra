package api

import (
	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/api/handlers"
	"github.com/orchestra/backend/internal/api/middleware"
	"github.com/orchestra/backend/internal/filesystem"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

func SetupRouter(pool *terminal.ProcessPool, gateway *ws.Gateway, db *storage.Database, allowedOrigins []string) *gin.Engine {
	r := gin.New()

	// 设置信任的代理（仅信任本地）
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// 初始化文件系统验证器和浏览器（允许访问用户主目录）
	validator := filesystem.NewValidator([]string{"~"})
	browser := filesystem.NewBrowser(validator)

	// 初始化仓库
	wsRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	msgRepo := repository.NewMessageRepository(db.DB())
	readRepo := repository.NewConversationReadRepository(db.DB())

	// 初始化 handlers
	wsHandler := handlers.NewWorkspaceHandler(wsRepo, memberRepo, browser)
	memberHandler := handlers.NewMemberHandler(memberRepo, wsRepo)
	terminalHandler := handlers.NewTerminalHandler(pool, wsRepo)
	convHandler := handlers.NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, pool)

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(gin.Recovery())

	// 认证配置（可选）
	authConfig := middleware.DefaultAuthConfig()

	// 健康检查（无需认证）
	r.GET("/health", handlers.HealthCheck)

	// API 路由组
	api := r.Group("/api")
	api.Use(middleware.Auth(authConfig))
	{
		// 工作区
		api.GET("/workspaces", wsHandler.List)
		api.POST("/workspaces", wsHandler.Create)
		api.GET("/workspaces/:id", wsHandler.Get)
		api.DELETE("/workspaces/:id", wsHandler.Delete)
		api.GET("/workspaces/:id/browse", wsHandler.Browse)
		api.GET("/browse", wsHandler.BrowseRoot)

		// 成员
		api.GET("/workspaces/:id/members", memberHandler.List)
		api.POST("/workspaces/:id/members", memberHandler.Create)
		api.PUT("/workspaces/:id/members/:memberId", memberHandler.Update)
		api.DELETE("/workspaces/:id/members/:memberId", memberHandler.Delete)
		api.DELETE("/workspaces/:id/members/:memberId/conversations", convHandler.DeleteConversationsForMember)
		api.GET("/workspaces/:id/members/:memberId/terminal-session", terminalHandler.GetSessionForMember)
		api.GET("/workspaces/:id/terminal-sessions", terminalHandler.ListWorkspaceTerminalSessions)

		// 终端会话
		api.POST("/terminals", terminalHandler.CreateSession)
		api.DELETE("/terminals/:sessionId", terminalHandler.DeleteSession)

		// 会话
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

	// WebSocket 路由（使用简化的认证中间件）
	wsAuth := middleware.WebSocketAuth(authConfig)
	r.GET("/ws/terminal/:sessionId", wsAuth, gateway.HandleTerminal)
	r.GET("/ws/chat/:workspaceId", wsAuth, gateway.HandleChat)

	return r
}