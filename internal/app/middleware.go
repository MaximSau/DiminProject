package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// statusWriter обёртка над http.ResponseWriter для перехвата статуса ответа и подсчёта байтов.
type statusWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
	wroteHeader  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Middleware объединяет структурированное логирование slog,
// перехват паник (panic recovery), генерацию Request-ID и заголовки безопасности.
func (a *App) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Извлекаем или генерируем уникальный Request ID
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			var b [8]byte
			_, _ = rand.Read(b[:])
			reqID = hex.EncodeToString(b[:])
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		r = r.WithContext(ctx)

		// 2. Строгие заголовки безопасности
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self';")

		// 3. Управление HTTP-кэшированием (без агрессивного кэширования для мгновенного применения правок)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// 4. Ограничиваем размер тела запроса (1 MB max) для защиты от DOS
		if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		}

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		// 5. Panic Recovery: перехват паник и отдача 500 страницы без падения процесса
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("перехвачена паника при обработке запроса",
					"error", rec,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
					"method", r.Method,
					"request_id", reqID,
				)
				if !sw.wroteHeader {
					a.render500(sw, r)
				}
			}
		}()

		next.ServeHTTP(sw, r)

		// 6. Логирование запросов через structured slog
		duration := time.Since(start)
		isStatic := strings.HasPrefix(r.URL.Path, "/static/") || r.URL.Path == "/favicon.ico"
		if !isStatic || sw.status >= 400 {
			lvl := slog.LevelInfo
			if sw.status >= 500 {
				lvl = slog.LevelError
			} else if sw.status >= 400 {
				lvl = slog.LevelWarn
			}
			slog.Log(ctx, lvl, "http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", duration.Milliseconds(),
				"bytes", sw.bytesWritten,
				"ip", clientIP(r),
				"request_id", reqID,
			)
		}
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}
