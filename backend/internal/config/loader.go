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

	// 从环境变量读取加密密钥
	if key := os.Getenv("ORCHESTRA_ENCRYPTION_KEY"); key != "" {
		cfg.Security.EncryptionKey = key
	}

	// 解析路径中的环境变量
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