package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orchestra/backend/internal/a2a"
	"github.com/orchestra/backend/internal/config"
	"github.com/orchestra/backend/internal/storage"
	"github.com/orchestra/backend/internal/ws"
)

func testDB(t *testing.T) *storage.Database {
	t.Helper()
	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	// Apply migrations if available
	_ = db.Migrate("internal/storage/migrations")
	return db
}

func TestSetupRouterReturnsEngine(t *testing.T) {
	db := testDB(t)
	cfg := config.Default()
	cfg.Auth.Enabled = false

	a2aPool := a2a.NewPool(0, "")
	gateway := ws.NewGateway(nil, nil)

	r, toolHandler := SetupRouter(a2aPool, gateway, db, cfg)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	if toolHandler == nil {
		t.Fatal("expected non-nil tool handler")
	}
}

func TestHealthEndpoint(t *testing.T) {
	db := testDB(t)
	cfg := config.Default()
	cfg.Auth.Enabled = false

	a2aPool := a2a.NewPool(0, "")
	gateway := ws.NewGateway(nil, nil)

	r, _ := SetupRouter(a2aPool, gateway, db, cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	db := testDB(t)
	cfg := config.Default()
	cfg.Auth.Enabled = false

	a2aPool := a2a.NewPool(0, "")
	gateway := ws.NewGateway(nil, nil)

	r, _ := SetupRouter(a2aPool, gateway, db, cfg)

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return something (not 404)
	if w.Code == http.StatusNotFound {
		t.Error("expected swagger endpoint to exist")
	}
}

func TestDependenciesStruct(t *testing.T) {
	deps := &Dependencies{
		Cfg: config.Default(),
	}
	if deps.Cfg == nil {
		t.Error("expected cfg to be set")
	}
}
