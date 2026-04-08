// Package main provides the Orchestra API server
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/orchestra/backend/docs" // swagger docs
	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/api"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/models"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/ws"
	"github.com/orchestra/backend/pkg/utils"
)

// @title Orchestra API
// @version 1.0
// @description Multi-Agent Collaboration Platform API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate("internal/storage/migrations"); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	// Initialize security whitelist
	security.NewWhitelist(
		cfg.Security.AllowedCommands,
		cfg.Security.AllowedPaths,
	)

	var encryptor *security.KeyEncryptor
	if cfg.Security.EncryptionKey != "" {
		encryptor, err = security.NewKeyEncryptor(cfg.Security.EncryptionKey)
		if err != nil {
			log.Fatalf("init encryptor: %v", err)
		}
	}

	// Create default user if auth is enabled and no users exist
	if cfg.Auth.Enabled && cfg.Auth.JWTSecret != "" {
		ensureDefaultUserAndWorkspace(db)
	}

	// Initialize A2A pool for agent communication
	a2aRegistry := a2a.NewAgentRegistry()
	workspacePath := cfg.Storage.Workspaces
	a2aPool := a2a.NewPool(
		cfg.Terminal.IdleTimeout,
		a2aRegistry,
		workspacePath,
	)

	// Initialize WebSocket gateway with A2A terminal handler
	a2aTerminalHandler := ws.NewA2ATerminalHandler(a2aPool)
	gateway := ws.NewGateway(a2aTerminalHandler, cfg.Security.AllowedOrigins)

	// Setup router
	router := api.SetupRouter(a2aPool, gateway, db, cfg)

	log.Printf("Orchestra starting on %s", cfg.Server.HTTPAddr)
	log.Printf("Authentication: enabled=%v", cfg.Auth.Enabled)

	// Start HTTP server
	go func() {
		if err := router.Run(cfg.Server.HTTPAddr); err != nil {
			log.Fatalf("start server: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	_ = encryptor // TODO: use encryptor for API key encryption
}

// ensureDefaultUserAndWorkspace creates a default admin user, workspace, and member if none exist
func ensureDefaultUserAndWorkspace(db *storage.Database) {
	userRepo := repository.NewUserRepository(db.DB())
	wsRepo := repository.NewWorkspaceRepository(db.DB())
	memberRepo := repository.NewMemberRepository(db.DB())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := userRepo.List(ctx)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return
	}

	if len(users) > 0 {
		return // Already has users, nothing to do
	}

	// Create default user
	hash, err := security.HashPassword("orchestra")
	if err != nil {
		log.Printf("Failed to hash default password: %v", err)
		return
	}

	user := &models.User{
		ID:           utils.GenerateID(),
		Username:     "orchestra",
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		log.Printf("Failed to create default user: %v", err)
		return
	}
	log.Println("Created default user: orchestra (password: orchestra)")

	// Create default workspace
	ws := &models.Workspace{
		ID:           utils.GenerateID(),
		Name:         "My First Workspace",
		Path:         ".",
		LastOpenedAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	if err := wsRepo.Create(ctx, ws); err != nil {
		log.Printf("Failed to create default workspace: %v", err)
		return
	}
	log.Printf("Created default workspace: %s (path: .)", ws.Name)

	// Create owner member for the default workspace
	member := &models.Member{
		ID:                utils.GenerateID(),
		WorkspaceID:       ws.ID,
		Name:              "orchestra",
		RoleType:          models.RoleOwner,
		TerminalType:      "native",
		TerminalCommand:   "/bin/bash",
		AutoStartTerminal: true,
		Status:            "online",
		CreatedAt:         time.Now(),
	}

	if err := memberRepo.Create(ctx, member); err != nil {
		log.Printf("Failed to create default member: %v", err)
		return
	}
	log.Printf("Created default member: %s (owner) in workspace %s", member.Name, ws.Name)
}
