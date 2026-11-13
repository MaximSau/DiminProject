package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pulsefit/internal/repository"
	"pulsefit/internal/service"
	transport "pulsefit/internal/transport/http"
)

func TestRegisterRoutes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_routes_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := repository.Open(tmpDir + "/test_routes.db")
	if err != nil {
		t.Fatalf("repo open: %v", err)
	}
	defer repo.Close()

	cat, err := service.LoadCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	svc := service.New(repo, cat)
	assets := os.DirFS("../../web")
	h, err := transport.NewHandler(svc, assets)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, h, assets)

	// Проверяем работу зарегистрированного роута /healthz через mux
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 from registered /healthz route, got %d", rec.Code)
	}
}
