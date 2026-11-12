// PulseFit — современное веб-приложение для тренировок, питания и контроля прогресса.
package main

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"pulsefit/internal/app"
)

// Шаблоны и статика вшиты в бинарник: приложение полностью автономно.
//
//go:embed web/templates web/static
var webFiles embed.FS

func main() {
	setupLogger()

	addr := envOr("PULSEFIT_ADDR", "0.0.0.0:8099")
	dataDir := envOr("PULSEFIT_DATA", "data")

	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		slog.Error("не удалось подготовить встроенные файлы", "error", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		slog.Error("не удалось создать каталог данных", "error", err)
		os.Exit(1)
	}

	srv, err := app.New(app.Config{DataDir: dataDir, Assets: assets})
	if err != nil {
		slog.Error("запуск приложения не удался", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := srv.Close(); err != nil {
			slog.Error("ошибка закрытия ресурсов приложения", "error", err)
		}
	}()

	// Middleware безопасности, Request-ID, паник-рекавери и структурированного логирования
	handler := srv.Middleware(srv.Handler())

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("🚀 PulseFit успешно запущен", "addr", "http://"+addr, "storage", dataDir+"/store.db")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("сервер остановлен с ошибкой", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Завершение работы PulseFit (graceful shutdown)...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при остановке сервера", "error", err)
	}
	slog.Info("Сервер успешно остановлен.")
}

func setupLogger() {
	var handler slog.Handler
	logFormat := strings.ToLower(envOr("PULSEFIT_LOG_FORMAT", envOr("LOG_FORMAT", "")))
	env := strings.ToLower(envOr("ENV", envOr("APP_ENV", "")))

	level := slog.LevelInfo
	if strings.ToLower(envOr("LOG_LEVEL", "")) == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	if logFormat == "json" || env == "production" || env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
