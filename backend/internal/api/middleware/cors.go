package middleware

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if isValidOrigin(origin, allowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isValidOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	for _, allowed := range allowedOrigins {
		// Support wildcard subdomain matching (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.ToLower(allowed[2:])
			host := strings.ToLower(originURL.Hostname())
			if host == domain || strings.HasSuffix(host, "."+domain) {
				return true
			}
			continue
		}

		// Exact match
		if origin == allowed {
			return true
		}
	}

	return false
}
