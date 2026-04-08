package ws

import (
	"net/http"
	"testing"
	"time"

	"github.com/orchestra/backend/internal/a2a"
)

func TestA2ATerminalHandler_Handle(t *testing.T) {
	registry := a2a.NewAgentRegistry()
	pool := a2a.NewPool(30*time.Minute, registry)
	handler := NewA2ATerminalHandler(pool)

	if handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestGateway_NewGateway(t *testing.T) {
	registry := a2a.NewAgentRegistry()
	pool := a2a.NewPool(30*time.Minute, registry)
	handler := NewA2ATerminalHandler(pool)
	allowedOrigins := []string{"http://localhost:3000", "http://example.com"}

	gateway := NewGateway(handler, allowedOrigins)

	if gateway == nil {
		t.Error("gateway should not be nil")
	}
	if len(gateway.allowedOrigins) != 2 {
		t.Errorf("expected 2 allowed origins, got %d", len(gateway.allowedOrigins))
	}
}

func TestGateway_checkOrigin(t *testing.T) {
	registry := a2a.NewAgentRegistry()
	pool := a2a.NewPool(30*time.Minute, registry)
	handler := NewA2ATerminalHandler(pool)

	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		expected       bool
	}{
		{
			name:           "exact match",
			allowedOrigins: []string{"http://localhost:3000", "http://example.com"},
			origin:         "http://localhost:3000",
			expected:       true,
		},
		{
			name:           "no match",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://malicious.com",
			expected:       false,
		},
		{
			name:           "wildcard subdomain match",
			allowedOrigins: []string{"*.example.com"},
			origin:         "http://sub.example.com",
			expected:       true,
		},
		{
			name:           "wildcard subdomain no match",
			allowedOrigins: []string{"*.example.com"},
			origin:         "http://other.com",
			expected:       false,
		},
		{
			name:           "empty origin",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "",
			expected:       false,
		},
		{
			name:           "invalid origin URL",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "not a valid url",
			expected:       false,
		},
		{
			name:           "empty allowed origins",
			allowedOrigins: []string{},
			origin:         "http://localhost:3000",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := NewGateway(handler, tt.allowedOrigins)

			r := &http.Request{}
			if tt.origin != "" {
				r.Header = http.Header{}
				r.Header.Set("Origin", tt.origin)
			}

			result := gateway.checkOrigin(r)
			if result != tt.expected {
				t.Errorf("checkOrigin() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
