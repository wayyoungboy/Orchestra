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

// Dependencies holds shared dependencies for route registration.
type Dependencies struct {
	DB         *storage.Database
	Cfg        *config.Config
	Validator  *filesystem.Validator
	Browser    *filesystem.Browser
	Gateway    *ws.Gateway
	A2APool    *a2a.Pool
	AuthConfig middleware.AuthConfig
	JWTConfig  *security.JWTConfig

	// Repositories
	WorkspaceRepo  repository.WorkspaceRepository
	MemberRepo     repository.MemberRepository
	ConvRepo       *repository.ConversationRepository
	MsgRepo        *repository.MessageRepository
	ReadRepo       *repository.ConversationReadRepository
	UserRepo       repository.UserRepository
	AttachRepo     *repository.AttachmentRepository
	TaskRepo       *repository.TaskRepo
	APIKeyRepo     repository.APIKeyRepository

	// Handlers (created by registerRepositories)
	WsHandler         *handlers.WorkspaceHandler
	MemberHandler     *handlers.MemberHandler
	TerminalHandler   *handlers.TerminalHandler
	ConvHandler       *handlers.ConversationHandler
	AttachmentHandler *handlers.AttachmentHandler
	TaskHandler       *handlers.TaskHandler
	APIKeyHandler     *handlers.APIKeyHandler
	AuthHandler       *handlers.AuthHandler
}

// SetupRouter creates and configures the Gin engine with all routes.
func SetupRouter(a2aPool *a2a.Pool, gateway *ws.Gateway, db *storage.Database, cfg *config.Config) (*gin.Engine, *a2a.ToolHandler) {
	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	deps := registerRepositories(db, cfg, r, a2aPool, gateway)
	registerAuthRoutes(r, deps)
	registerWorkspaceRoutes(r, deps)
	registerMemberRoutes(r, deps)
	registerTerminalRoutes(r, deps)
	registerConversationRoutes(r, deps)
	registerTaskRoutes(r, deps)
	registerAttachmentRoutes(r, deps)
	registerAPIKeyRoutes(r, deps)
	registerWebSocketRoutes(r, deps)
	wireUpIntegrations(a2aPool, deps)

	return r, deps.wireUpToolHandler(a2aPool)
}

func registerRepositories(db *storage.Database, cfg *config.Config, r *gin.Engine, a2aPool *a2a.Pool, gateway *ws.Gateway) *Dependencies {
	allowedPaths := cfg.Security.AllowedPaths
	if len(allowedPaths) == 0 {
		allowedPaths = []string{"~"}
	}
	validator := filesystem.NewValidator(allowedPaths)
	browser := filesystem.NewBrowser(validator)

	var jwtConfig *security.JWTConfig
	if cfg.Auth.JWTSecret != "" {
		jwtConfig = security.NewJWTConfig(cfg.Auth.JWTSecret)
	}

	wsRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())
	convRepo := repository.NewConversationRepository(db.DB())
	msgRepo := repository.NewMessageRepository(db.DB())
	readRepo := repository.NewConversationReadRepository(db.DB())
	userRepo := repository.NewUserRepository(db.DB())
	attachRepo := repository.NewAttachmentRepository(db.DB())
	taskRepo := repository.NewTaskRepository(db.DB())
	apiKeyRepo := repository.NewAPIKeyRepository(db.DB())

	baseURL := os.Getenv("ORCHESTRA_BASE_URL")

	wsHandler := handlers.NewWorkspaceHandler(wsRepo, memberRepo, msgRepo, browser)
	memberHandler := handlers.NewMemberHandler(memberRepo, wsRepo, ws.GlobalChatHub)
	terminalHandler := handlers.NewTerminalHandler(a2aPool, wsRepo)
	convHandler := handlers.NewConversationHandler(convRepo, msgRepo, readRepo, memberRepo, wsRepo, a2aPool, ws.GlobalChatHub, cfg.Server.HTTPAddr, cfg.Auth.Enabled, baseURL)
	attachmentHandler := handlers.NewAttachmentHandler(msgRepo, convRepo, attachRepo, cfg.Server.UploadDir)
	taskHandler := handlers.NewTaskHandler(taskRepo, memberRepo, ws.GlobalChatHub)
	apiKeyHandler, err := handlers.NewAPIKeyHandler(apiKeyRepo, cfg)
	if err != nil {
		panic("failed to create API key handler: " + err.Error())
	}
	authHandler := handlers.NewAuthHandler(userRepo, jwtConfig, cfg.Auth.Enabled)

	authConfig := middleware.DefaultAuthConfig(cfg.Auth.JWTSecret)

	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(gin.Recovery())

	return &Dependencies{
		DB:         db,
		Cfg:        cfg,
		Validator:  validator,
		Browser:    browser,
		Gateway:    gateway,
		A2APool:    a2aPool,
		AuthConfig: authConfig,
		JWTConfig:  jwtConfig,

		WorkspaceRepo: wsRepo,
		MemberRepo:    memberRepo,
		ConvRepo:      convRepo,
		MsgRepo:       msgRepo,
		ReadRepo:      readRepo,
		UserRepo:      userRepo,
		AttachRepo:    attachRepo,
		TaskRepo:      taskRepo,
		APIKeyRepo:    apiKeyRepo,

		WsHandler:         wsHandler,
		MemberHandler:     memberHandler,
		TerminalHandler:   terminalHandler,
		ConvHandler:       convHandler,
		AttachmentHandler: attachmentHandler,
		TaskHandler:       taskHandler,
		APIKeyHandler:     apiKeyHandler,
		AuthHandler:       authHandler,
	}
}

func registerHealthRoute(r *gin.Engine) {
	r.GET("/health", handlers.HealthCheck)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func registerAuthRoutes(r *gin.Engine, deps *Dependencies) {
	registerHealthRoute(r)

	authGroup := r.Group("/api/auth")
	{
		authGroup.GET("/config", deps.AuthHandler.GetAuthConfig)
		authGroup.POST("/login", deps.AuthHandler.Login)
		authGroup.POST("/validate", deps.AuthHandler.ValidateToken)
		authGroup.GET("/me", middleware.Auth(deps.AuthConfig), deps.AuthHandler.GetCurrentUser)

		if deps.Cfg.Auth.AllowRegistration {
			authGroup.POST("/register", deps.AuthHandler.Register)
		}
	}
}

func registerWorkspaceRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/workspaces", deps.WsHandler.List)
	api.POST("/workspaces", deps.WsHandler.Create)
	api.POST("/workspaces/validate-path", deps.WsHandler.ValidatePath)
	api.GET("/workspaces/:id", deps.WsHandler.Get)
	api.PUT("/workspaces/:id", deps.WsHandler.Update)
	api.DELETE("/workspaces/:id", deps.WsHandler.Delete)
	api.GET("/workspaces/:id/browse", deps.WsHandler.Browse)
	api.GET("/workspaces/:id/search", deps.WsHandler.Search)
	api.GET("/browse", deps.WsHandler.BrowseRoot)
}

func registerMemberRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/workspaces/:id/members", deps.MemberHandler.List)
	api.GET("/workspaces/:id/members/:memberId", deps.MemberHandler.Get)
	api.POST("/workspaces/:id/members", deps.MemberHandler.Create)
	api.PUT("/workspaces/:id/members/:memberId", deps.MemberHandler.Update)
	api.DELETE("/workspaces/:id/members/:memberId", deps.MemberHandler.Delete)
	api.DELETE("/workspaces/:id/members/:memberId/conversations", deps.ConvHandler.DeleteConversationsForMember)
	api.POST("/workspaces/:id/members/:memberId/presence", deps.MemberHandler.UpdatePresence)
}

func registerTerminalRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/workspaces/:id/members/:memberId/terminal-session", deps.TerminalHandler.GetSessionForMember)
	api.POST("/workspaces/:id/members/:memberId/terminal-session", deps.TerminalHandler.GetOrCreateSessionForMember)
	api.GET("/workspaces/:id/terminal-sessions", deps.TerminalHandler.ListWorkspaceTerminalSessions)
	api.POST("/terminals", deps.TerminalHandler.CreateSession)
	api.DELETE("/terminals/:sessionId", deps.TerminalHandler.DeleteSession)
}

func registerConversationRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/workspaces/:id/conversations", deps.ConvHandler.List)
	api.GET("/workspaces/:id/conversations/:convId", deps.ConvHandler.GetConversation)
	api.POST("/workspaces/:id/conversations", deps.ConvHandler.Create)
	api.PUT("/workspaces/:id/conversations/:convId", deps.ConvHandler.UpdateSettings)
	api.DELETE("/workspaces/:id/conversations/:convId", deps.ConvHandler.Delete)
	api.DELETE("/workspaces/:id/conversations/:convId/messages", deps.ConvHandler.ClearMessages)
	api.DELETE("/workspaces/:id/conversations/:convId/messages/:messageId", deps.ConvHandler.DeleteMessage)
	api.GET("/workspaces/:id/conversations/:convId/messages", deps.ConvHandler.GetMessages)
	api.POST("/workspaces/:id/conversations/:convId/messages", deps.ConvHandler.SendMessage)
	api.POST("/workspaces/:id/conversations/:convId/read", deps.ConvHandler.MarkConversationRead)
	api.POST("/workspaces/:id/conversations/read-all", deps.ConvHandler.MarkAllConversationsRead)
	api.PUT("/workspaces/:id/conversations/:convId/members", deps.ConvHandler.SetConversationMembers)

	// Internal API for AI assistants
	api.POST("/internal/chat/send", deps.ConvHandler.InternalChatSend)
	api.POST("/internal/agent-status", deps.ConvHandler.UpdateAgentStatus)
}

func registerTaskRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	// Internal task API
	api.POST("/internal/tasks/create", deps.TaskHandler.CreateTask)
	api.POST("/internal/tasks/assign", deps.TaskHandler.AssignTask)
	api.POST("/internal/tasks/start", deps.TaskHandler.StartTask)
	api.POST("/internal/tasks/complete", deps.TaskHandler.CompleteTask)
	api.POST("/internal/tasks/fail", deps.TaskHandler.FailTask)
	api.POST("/internal/tasks/cancel", deps.TaskHandler.CancelTask)
	api.GET("/internal/workloads/list", deps.TaskHandler.ListWorkloads)

	// Task management (frontend)
	api.GET("/workspaces/:id/tasks", deps.TaskHandler.ListTasks)
	api.GET("/workspaces/:id/tasks/:taskId", deps.TaskHandler.GetTask)
	api.GET("/workspaces/:id/tasks/my-tasks", deps.TaskHandler.GetMyTasks)
	api.POST("/workspaces/:id/tasks/:taskId/cancel", deps.TaskHandler.CancelTask)
}

func registerAttachmentRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/workspaces/:id/attachments", deps.AttachmentHandler.ListAttachments)
	api.POST("/workspaces/:id/conversations/:convId/attachments", deps.AttachmentHandler.UploadAttachment)
	api.GET("/workspaces/:id/attachments/:attachmentId", deps.AttachmentHandler.DownloadAttachment)
	api.GET("/workspaces/:id/attachments/:attachmentId/info", deps.AttachmentHandler.GetAttachmentInfo)
	api.DELETE("/workspaces/:id/attachments/:attachmentId", deps.AttachmentHandler.DeleteAttachment)
}

func registerAPIKeyRoutes(r *gin.Engine, deps *Dependencies) {
	api := r.Group("/api")
	api.Use(middleware.Auth(deps.AuthConfig))

	api.GET("/api-keys", deps.APIKeyHandler.List)
	api.GET("/api-keys/provider/:provider", deps.APIKeyHandler.GetByProvider)
	api.POST("/api-keys", deps.APIKeyHandler.Create)
	api.DELETE("/api-keys/:id", deps.APIKeyHandler.Delete)
	api.POST("/api-keys/test", deps.APIKeyHandler.Test)
}

func registerWebSocketRoutes(r *gin.Engine, deps *Dependencies) {
	wsAuth := middleware.WebSocketAuth(deps.AuthConfig)
	r.GET("/ws/terminal/:sessionId", wsAuth, deps.Gateway.HandleTerminal)
	r.GET("/ws/chat/:workspaceId", wsAuth, deps.Gateway.HandleChat)
}

func wireUpIntegrations(a2aPool *a2a.Pool, deps *Dependencies) {
	bridge := chatbridge.NewAgentBridge(deps.MsgRepo, ws.GlobalChatHub)
	a2aPool.SetOutputHook(func(sess *a2a.Session, msg *a2a.ACPMessage) {
		bridge.OnMessage(sess, msg)
	})
}

func (deps *Dependencies) wireUpToolHandler(a2aPool *a2a.Pool) *a2a.ToolHandler {
	chatBroadcaster := &chatEventAdapter{hub: ws.GlobalChatHub}
	toolHandler := a2a.NewToolHandler(deps.MsgRepo, deps.TaskRepo, deps.MemberRepo, deps.ConvRepo, chatBroadcaster, deps.Browser, deps.Validator)
	a2aPool.SetToolHandler(toolHandler)
	toolHandler.SetPool(a2aPool)
	toolHandler.SetWorkspaceRepo(deps.WorkspaceRepo)
	return toolHandler
}

// chatEventAdapter wraps ws.ChatHub to satisfy a2a.ChatBroadcaster interface.
// The ToolHandler passes raw maps which this adapter converts to ChatEvent structs.
type chatEventAdapter struct {
	hub *ws.ChatHub
}

func (a *chatEventAdapter) BroadcastToWorkspace(workspaceID string, event interface{}) {
	if ce, ok := event.(ws.ChatEvent); ok {
		a.hub.BroadcastToWorkspace(workspaceID, ce)
		return
	}
	if raw, ok := event.(map[string]interface{}); ok {
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
