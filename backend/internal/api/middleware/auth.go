package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/security"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled     bool
	Disabled    bool // Explicit disable flag
	JWTConfig   *security.JWTConfig
	LegacyToken string // For backward compatibility with ORCHESTRA_AUTH_TOKEN
}

// DefaultAuthConfig returns default auth config from environment
func DefaultAuthConfig(jwtSecret string) AuthConfig {
	// Check if auth is explicitly disabled
	authDisabled := os.Getenv("ORCHESTRA_AUTH_DISABLED") == "true"

	// Legacy token support
	legacyToken := os.Getenv("ORCHESTRA_AUTH_TOKEN")

	// Auth is enabled if:
	// 1. Not explicitly disabled
	// 2. Either legacy token is set OR JWT secret is set
	authEnabled := !authDisabled && (legacyToken != "" || jwtSecret != "")

	var jwtConfig *security.JWTConfig
	if jwtSecret != "" {
		jwtConfig = security.NewJWTConfig(jwtSecret)
	}

	return AuthConfig{
		Enabled:     authEnabled,
		Disabled:    authDisabled,
		JWTConfig:   jwtConfig,
		LegacyToken: legacyToken,
	}
}

// Auth returns an authentication middleware that supports both JWT and legacy token
func Auth(cfg AuthConfig) gin.HandlerFunc {
	// If auth is disabled, pass through
	if cfg.Disabled {
		return func(c *gin.Context) {
			c.Set("authDisabled", true)
			c.Next()
		}
	}

	// If auth is not configured, pass through
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Try JWT authentication first
		if cfg.JWTConfig != nil {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := cfg.JWTConfig.ValidateToken(token)
				if err == nil {
					c.Set("userId", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("authMethod", "jwt")
					c.Next()
					return
				}
			}
		}

		// Fall back to legacy token authentication
		if cfg.LegacyToken != "" {
			// Check Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token == cfg.LegacyToken {
					c.Set("authMethod", "legacy")
					c.Next()
					return
				}
			}

			// Check query parameter (for WebSocket connections)
			token := c.Query("token")
			if token != "" && token == cfg.LegacyToken {
				c.Set("authMethod", "legacy")
				c.Next()
				return
			}

			// Check custom header
			token = c.GetHeader("X-Auth-Token")
			if token != "" && token == cfg.LegacyToken {
				c.Set("authMethod", "legacy")
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}

// WebSocketAuth returns auth middleware optimized for WebSocket routes
func WebSocketAuth(cfg AuthConfig) gin.HandlerFunc {
	if cfg.Disabled || !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// For WebSocket, prioritize query parameter (since headers may not be available)
		token := c.Query("token")

		// Try JWT first
		if cfg.JWTConfig != nil && token != "" {
			claims, err := cfg.JWTConfig.ValidateToken(token)
			if err == nil {
				c.Set("userId", claims.UserID)
				c.Set("username", claims.Username)
				c.Next()
				return
			}
		}

		// Fall back to legacy token
		if cfg.LegacyToken != "" && token == cfg.LegacyToken {
			c.Next()
			return
		}

		// Also check custom header for non-browser WebSocket clients
		headerToken := c.GetHeader("X-Auth-Token")
		if cfg.JWTConfig != nil && headerToken != "" {
			claims, err := cfg.JWTConfig.ValidateToken(headerToken)
			if err == nil {
				c.Set("userId", claims.UserID)
				c.Set("username", claims.Username)
				c.Next()
				return
			}
		}

		if cfg.LegacyToken != "" && headerToken == cfg.LegacyToken {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}