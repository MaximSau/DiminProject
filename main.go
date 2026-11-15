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
//go:embed web/templates web/static web/dist
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

	go func() {
		slog.Info("сервер PulseFit запущен", "address", addr, "env_addr", os.Getenv("PULSEFIT_ADDR"))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("критическая ошибка HTTP-сервера", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("получен сигнал завершения, выполняется graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при завершении сервера", "error", err)
	} else {
		slog.Info("сервер PulseFit корректно остановлен")
	}
}

func envOr(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func setupLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}
