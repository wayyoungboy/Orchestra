package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/orchestra/backend/internal/api"
	"github.com/orchestra/backend/internal/chatbridge"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/security"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/storage/repository"
	"github.com/orchestra/backend/internal/terminal"
	"github.com/orchestra/backend/internal/ws"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	db, err := storage.NewDatabase(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	defer db.Close()

	// 执行迁移
	if err := db.Migrate("internal/storage/migrations"); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	// 初始化安全模块
	whitelist := security.NewWhitelist(
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

	// 初始化进程池
	pool := terminal.NewProcessPool(
		cfg.Terminal.MaxSessions,
		cfg.Terminal.IdleTimeout,
	)
	pool.SetValidator(whitelist)

	msgRepo := repository.NewMessageRepository(db.DB())
	chatBridge := chatbridge.New(msgRepo)
	pool.SetOutputHook(chatBridge.OnTerminalOutput)

	// 初始化 WebSocket 网关
	terminalHandler := ws.NewTerminalHandler(pool)
	gateway := ws.NewGateway(terminalHandler, cfg.Security.AllowedOrigins)

	// 启动 HTTP 服务器
	router := api.SetupRouter(pool, gateway, db, cfg.Security.AllowedOrigins)

	log.Printf("Orchestra starting on %s", cfg.Server.HTTPAddr)
	log.Printf("Allowed commands: %v", whitelist.AllowedCommands())
	log.Printf("Allowed paths: %v", whitelist.AllowedPaths())

	go func() {
		if err := router.Run(cfg.Server.HTTPAddr); err != nil {
			log.Fatalf("start server: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	_ = encryptor // TODO: 使用加密器
	fmt.Println("Goodbye!")
}