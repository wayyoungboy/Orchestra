package middleware

import "testing"

func TestDefaultAuthConfigHonorsExplicitDisabledConfig(t *testing.T) {
	t.Setenv("ORCHESTRA_AUTH_DISABLED", "")
	config := DefaultAuthConfig(false, "configured-but-inactive-secret")
	if config.Enabled || !config.Disabled {
		t.Fatalf("auth config = %#v, want explicitly disabled", config)
	}
}

func TestDefaultAuthConfigEnablesConfiguredAuth(t *testing.T) {
	t.Setenv("ORCHESTRA_AUTH_DISABLED", "")
	config := DefaultAuthConfig(true, "configured-secret")
	if !config.Enabled || config.Disabled || config.JWTConfig == nil {
		t.Fatalf("auth config = %#v, want enabled JWT auth", config)
	}
}
