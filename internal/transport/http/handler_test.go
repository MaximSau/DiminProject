package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pulsefit/internal/repository"
	"pulsefit/internal/service"
)

func TestHandlerHealthzAndManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pulsefit_http_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := repository.Open(tmpDir + "/test_http.db")
	if err != nil {
		t.Fatalf("repo open: %v", err)
	}
	defer repo.Close()

	cat, err := service.LoadCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	svc := service.New(repo, cat)
	assets := os.DirFS("../../../web")
	h, err := NewHandler(svc, assets)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	// 1. Healthz
	reqHealth := httptest.NewRequest("GET", "/healthz", nil)
	recHealth := httptest.NewRecorder()
	h.HandleHealthz(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recHealth.Code)
	}

	// 2. Manifest
	reqMan := httptest.NewRequest("GET", "/manifest.json", nil)
	recMan := httptest.NewRecorder()
	h.HandleManifest(recMan, reqMan)
	if recMan.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recMan.Code)
	}

	// 3. Service Worker
	reqSW := httptest.NewRequest("GET", "/sw.js", nil)
	recSW := httptest.NewRecorder()
	h.HandleServiceWorker(recSW, reqSW)
	if recSW.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recSW.Code)
	}
}
