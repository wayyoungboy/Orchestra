package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		path = os.Getenv("ORCHESTRA_CONFIG")
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	// Read encryption key from environment
	if key := os.Getenv("ORCHESTRA_ENCRYPTION_KEY"); key != "" {
		cfg.Security.EncryptionKey = key
	}

	// Auth configuration from environment
	if secret := os.Getenv("ORCHESTRA_JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
		cfg.Auth.Enabled = true
	}

	if os.Getenv("ORCHESTRA_AUTH_DISABLED") == "true" {
		cfg.Auth.Enabled = false
	}

	if os.Getenv("ORCHESTRA_ALLOW_REGISTRATION") == "true" {
		cfg.Auth.AllowRegistration = true
	}

	// Expand paths
	cfg.Storage.Database = expandPath(cfg.Storage.Database)
	cfg.Storage.Workspaces = expandPath(cfg.Storage.Workspaces)

	return cfg, nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}