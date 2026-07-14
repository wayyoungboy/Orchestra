// Package main provides the Orchestra API server
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
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
	cfg, err := config.Load(serverConfigPath())
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := validateStartupConfig(cfg); err != nil {
		log.Fatalf("invalid security configuration: %v", err)
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

	// This policy is passed to the session pool and enforced immediately before
	// tmux starts an agent process.
	whitelist := security.NewWhitelist(
		cfg.Security.AllowedCommands,
		cfg.Security.AllowedPaths,
	)

	// A deployment with authentication must explicitly provide the first admin;
	// do not create a known default password.
	if cfg.Auth.Enabled {
		if err := ensureBootstrapAdmin(db); err != nil {
			log.Fatalf("bootstrap administrator: %v", err)
		}
	}

	// Initialize A2A pool for agent communication (tmux-backed)
	workspacePath := cfg.Storage.Workspaces
	a2aPool := a2a.NewPool(
		cfg.Terminal.IdleTimeout,
		workspacePath,
	)
	a2aPool.SetExecutionPolicy(whitelist)
	a2aPool.SetMaxSessions(cfg.Terminal.MaxSessions)

	// Recover existing tmux sessions on startup
	recoverCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := a2aPool.RecoverSessions(recoverCtx); err != nil {
		log.Printf("Session recovery error: %v", err)
	}
	cancel()

	// Initialize WebSocket gateway with A2A terminal handler
	a2aTerminalHandler := ws.NewA2ATerminalHandler(a2aPool)
	gateway := ws.NewGateway(a2aTerminalHandler, cfg.Security.AllowedOrigins)

	// Setup router
	router, toolHandler := api.SetupRouter(a2aPool, gateway, db, cfg)

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

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if toolHandler != nil {
		if err := toolHandler.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down tool handler: %v", err)
		}
	}

}

func serverConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("ORCHESTRA_CONFIG")); path != "" {
		return path
	}
	return "configs/config.yaml"
}

func validateStartupConfig(cfg *config.Config) error {
	if cfg.Auth.Enabled && cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("authentication is enabled but no JWT secret is configured")
	}
	if cfg.Auth.AllowRegistration {
		return fmt.Errorf("self-registration is unavailable until workspace-level user authorization exists")
	}
	if !cfg.Auth.Enabled && !isLoopbackAddress(cfg.Server.HTTPAddr) {
		return fmt.Errorf("authentication must be enabled when listening on a non-loopback address")
	}
	return nil
}

func isLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// ensureBootstrapAdmin creates the first administrator from explicit
// deployment secrets. It deliberately never creates a predictable account.
func ensureBootstrapAdmin(db *storage.Database) error {
	userRepo := repository.NewUserRepository(db.DB())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := userRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	if len(users) == 1 {
		return nil
	}
	if len(users) > 1 {
		return fmt.Errorf("authenticated mode currently supports one administrator; migrate users only after workspace-level authorization is available")
	}

	username := strings.TrimSpace(os.Getenv("ORCHESTRA_ADMIN_USERNAME"))
	password := os.Getenv("ORCHESTRA_ADMIN_PASSWORD")
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("ORCHESTRA_ADMIN_USERNAME must be 3-50 characters for the first authenticated startup")
	}
	if len(password) < 12 {
		return fmt.Errorf("ORCHESTRA_ADMIN_PASSWORD must be at least 12 characters for the first authenticated startup")
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash administrator password: %w", err)
	}

	user := &models.User{
		ID:           utils.GenerateID(),
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create administrator: %w", err)
	}
	log.Printf("Created bootstrap administrator %q", user.Username)
	return nil
}
