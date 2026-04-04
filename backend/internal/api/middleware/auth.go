package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled bool
	Secret  string
}

// DefaultAuthConfig returns default auth config from environment
func DefaultAuthConfig() AuthConfig {
	token := os.Getenv("ORCHESTRA_AUTH_TOKEN")
	return AuthConfig{
		Enabled: token != "",
		Secret:  token,
	}
}

// Auth returns an authentication middleware
// If auth is disabled, it passes through all requests
func Auth(cfg AuthConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == cfg.Secret {
				c.Next()
				return
			}
		}

		// Check query parameter (for WebSocket connections)
		token := c.Query("token")
		if token != "" && token == cfg.Secret {
			c.Next()
			return
		}

		// Check custom header
		token = c.GetHeader("X-Auth-Token")
		if token != "" && token == cfg.Secret {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}

// WebSocketAuth returns a simpler auth middleware for WebSocket routes
// It only checks query parameter and custom header (not Authorization header)
func WebSocketAuth(cfg AuthConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Check query parameter
		token := c.Query("token")
		if token != "" && token == cfg.Secret {
			c.Next()
			return
		}

		// Check custom header
		token = c.GetHeader("X-Auth-Token")
		if token != "" && token == cfg.Secret {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
	}
}