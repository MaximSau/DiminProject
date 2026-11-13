package middleware

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

// statusWriter обёртка над http.ResponseWriter для перехвата HTTP статуса и размера тела ответа.
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

// RequestID извлекает или генерирует уникальный идентификатор запроса.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			var b [8]byte
			_, _ = rand.Read(b[:])
			reqID = hex.EncodeToString(b[:])
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SecurityHeaders устанавливает строгие заголовки безопасности HTTP.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self';")
		next.ServeHTTP(w, r)
	})
}

// CacheControl отключает агрессивное кэширование браузера для динамических страниц и свежего CSS.
func CacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize ограничивает размер тела запроса (по умолчанию 1 МБ).
func MaxBodySize(limitBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
				r.Body = http.MaxBytesReader(w, r.Body, limitBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PanicRecovery перехватывает паники, логирует стек и отдает 500 ошибку без краша сервиса.
func PanicRecovery(render500 func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw, ok := w.(*statusWriter)
			if !ok {
				sw = &statusWriter{ResponseWriter: w, status: http.StatusOK}
			}

			defer func() {
				if rec := recover(); rec != nil {
					reqID, _ := r.Context().Value(RequestIDKey).(string)
					slog.Error("перехвачена паника при обработке HTTP запроса",
						"error", rec,
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
						"request_id", reqID,
					)
					if render500 != nil && !sw.wroteHeader {
						render500(sw, r)
					} else if !sw.wroteHeader {
						http.Error(sw, "Внутренняя ошибка сервера", http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(sw, r)
		})
	}
}

// Logging выполняет структурированное логирование HTTP-запросов через slog.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw, ok := w.(*statusWriter)
		if !ok {
			sw = &statusWriter{ResponseWriter: w, status: http.StatusOK}
			w = sw
		}

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		reqID, _ := r.Context().Value(RequestIDKey).(string)
		isStatic := strings.HasPrefix(r.URL.Path, "/static/") || r.URL.Path == "/favicon.ico"

		if !isStatic || sw.status >= 400 {
			lvl := slog.LevelInfo
			if sw.status >= 500 {
				lvl = slog.LevelError
			} else if sw.status >= 400 {
				lvl = slog.LevelWarn
			}
			slog.Log(r.Context(), lvl, "http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", duration.Milliseconds(),
				"bytes", sw.bytesWritten,
				"ip", ClientIP(r),
				"request_id", reqID,
			)
		}
	})
}

// ClientIP определяет реальный IP клиента с учётом прокси.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}

// FullMiddlewareChain собирает стандартный стек мидлварей PulseFit.
func FullMiddlewareChain(render500 func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := next
		handler = Logging(handler)
		handler = PanicRecovery(render500)(handler)
		handler = MaxBodySize(1 << 20)(handler)
		handler = CacheControl(handler)
		handler = SecurityHeaders(handler)
		handler = RequestID(handler)
		return handler
	}
}
