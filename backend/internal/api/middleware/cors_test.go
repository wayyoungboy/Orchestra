package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsValidOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "exact match",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{"http://localhost:3000", "http://example.com"},
			expected:       true,
		},
		{
			name:           "no match",
			origin:         "http://malicious.com",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       false,
		},
		{
			name:           "wildcard subdomain match",
			origin:         "http://sub.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard root domain",
			origin:         "http://example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard no match",
			origin:         "http://other.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       false,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       false,
		},
		{
			name:           "invalid origin URL",
			origin:         "not a valid url",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       false,
		},
		{
			name:           "empty allowed origins",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{},
			expected:       false,
		},
		{
			name:           "https vs http",
			origin:         "https://localhost:3000",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       false,
		},
		{
			name:           "port different",
			origin:         "http://localhost:3001",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidOrigin(tt.origin, tt.allowedOrigins)
			if result != tt.expected {
				t.Errorf("isValidOrigin(%q, %v) = %v, expected %v", tt.origin, tt.allowedOrigins, result, tt.expected)
			}
		})
	}
}

func TestCORS_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectHeader   string
		method         string
	}{
		{
			name:           "allowed origin sets header",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expectHeader:   "http://localhost:3000",
			method:         "GET",
		},
		{
			name:           "disallowed origin no header",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://malicious.com",
			expectHeader:   "",
			method:         "GET",
		},
		{
			name:           "OPTIONS request handled",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expectHeader:   "http://localhost:3000",
			method:         "OPTIONS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(CORS(tt.allowedOrigins))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			originHeader := w.Header().Get("Access-Control-Allow-Origin")
			if originHeader != tt.expectHeader {
				t.Errorf("Access-Control-Allow-Origin = %q, expected %q", originHeader, tt.expectHeader)
			}

			if tt.method == "OPTIONS" && w.Code != 204 {
				t.Errorf("OPTIONS request should return 204, got %d", w.Code)
			}
		})
	}
}

func TestCORS_CredentialsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	credentials := w.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "true" {
		t.Errorf("Access-Control-Allow-Credentials should be 'true' for allowed origins, got %q", credentials)
	}
}