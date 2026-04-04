package config

import "time"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Terminal TerminalConfig `yaml:"terminal"`
	Security SecurityConfig `yaml:"security"`
	Storage  StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
	WSAddr   string `yaml:"ws_addr"`
}

type TerminalConfig struct {
	MaxSessions int           `yaml:"max_sessions"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type SecurityConfig struct {
	EncryptionKey    string   `yaml:"encryption_key"`
	AllowedCommands  []string `yaml:"allowed_commands"`
	AllowedPaths     []string `yaml:"allowed_paths"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
}

type StorageConfig struct {
	Database   string `yaml:"database"`
	Workspaces string `yaml:"workspaces"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPAddr: ":8080",
			WSAddr:   ":8081",
		},
		Terminal: TerminalConfig{
			MaxSessions: 10,
			IdleTimeout: 30 * time.Minute,
		},
		Security: SecurityConfig{
			EncryptionKey:   "",
			AllowedCommands: []string{"claude", "gemini", "codex", "qwen"},
			AllowedPaths:    []string{},
			AllowedOrigins:  []string{"http://localhost:3000", "http://localhost:5173"},
		},
		Storage: StorageConfig{
			Database:   "./data/orchestra.db",
			Workspaces: "./workspaces",
		},
	}
}