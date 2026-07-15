package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orchestra/backend/internal/config"
)

func TestAPIKeyStorageRequiresEncryptionKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, err := NewAPIKeyHandler(nil, config.Default())
	if err != nil {
		t.Fatalf("NewAPIKeyHandler() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/api-keys", nil)
	handler.List(ctx)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}
