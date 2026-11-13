package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewares(t *testing.T) {
	// 1. Тест RequestID и заголовков безопасности
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID, ok := r.Context().Value(RequestIDKey).(string)
		if !ok || reqID == "" {
			t.Errorf("missing request ID in context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	stack := FullMiddlewareChain(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "panic handled", http.StatusInternalServerError)
	})(baseHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	stack.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff header")
	}
	if !strings.Contains(rec.Header().Get("Cache-Control"), "no-cache") {
		t.Errorf("expected no-cache Cache-Control header")
	}

	// 2. Тест PanicRecovery
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom!")
	})
	panicStack := FullMiddlewareChain(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "panic recovered", http.StatusInternalServerError)
	})(panicHandler)

	pReq := httptest.NewRequest("GET", "/panic", nil)
	pRec := httptest.NewRecorder()
	panicStack.ServeHTTP(pRec, pReq)

	if pRec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for panic, got %d", pRec.Code)
	}
	if !strings.Contains(pRec.Body.String(), "panic recovered") {
		t.Errorf("expected panic recovered body")
	}
}
