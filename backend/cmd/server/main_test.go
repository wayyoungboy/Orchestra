package main

import (
	"strings"
	"testing"

	"github.com/orchestra/backend/internal/config"
)

func TestValidateStartupConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*config.Config)
		wantErr string
	}{
		{
			name:   "default local development",
			mutate: func(*config.Config) {},
		},
		{
			name: "public listener requires authentication",
			mutate: func(cfg *config.Config) {
				cfg.Server.HTTPAddr = ":8080"
			},
			wantErr: "authentication must be enabled",
		},
		{
			name: "enabled auth requires secret",
			mutate: func(cfg *config.Config) {
				cfg.Auth.Enabled = true
			},
			wantErr: "no JWT secret",
		},
		{
			name: "registration requires authorization model",
			mutate: func(cfg *config.Config) {
				cfg.Auth.Enabled = true
				cfg.Auth.JWTSecret = "test-secret"
				cfg.Auth.AllowRegistration = true
			},
			wantErr: "self-registration",
		},
		{
			name: "authenticated public listener",
			mutate: func(cfg *config.Config) {
				cfg.Server.HTTPAddr = ":8080"
				cfg.Auth.Enabled = true
				cfg.Auth.JWTSecret = "test-secret"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			tt.mutate(cfg)
			err := validateStartupConfig(cfg)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateStartupConfig() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("validateStartupConfig() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestServerConfigPath(t *testing.T) {
	t.Setenv("ORCHESTRA_CONFIG", "/tmp/orchestra.yaml")
	if got := serverConfigPath(); got != "/tmp/orchestra.yaml" {
		t.Fatalf("serverConfigPath() = %q", got)
	}

	t.Setenv("ORCHESTRA_CONFIG", "")
	if got := serverConfigPath(); got != "configs/config.yaml" {
		t.Fatalf("serverConfigPath() = %q", got)
	}
}
