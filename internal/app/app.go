package app

import (
	"fmt"
	"io/fs"
	"net/http"

	"pulsefit/internal/middleware"
	"pulsefit/internal/repository"
	"pulsefit/internal/routes"
	"pulsefit/internal/service"
	transport "pulsefit/internal/transport/http"
)

// Config конфигурация запуска приложения.
type Config struct {
	DataDir string
	Assets  fs.FS
}

// App основной контроллер веб-приложения PulseFit, объединяющий слои репозитория, сервиса, транспорта и мидлварей.
type App struct {
	repo    *repository.SQLiteRepository
	store   *Store // Псевдоним репозитория для обратной совместимости с тестами
	service *service.Service
	handler *transport.Handler
	mux     *http.ServeMux
}

// New создаёт и связывает все архитектурные слои приложения.
func New(cfg Config) (*App, error) {
	cat, err := service.LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки каталога: %w", err)
	}

	dbPath := fmt.Sprintf("%s/store.db", cfg.DataDir)
	repo, err := repository.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия хранилища: %w", err)
	}

	svc := service.New(repo, cat)
	h, err := transport.NewHandler(svc, cfg.Assets)
	if err != nil {
		_ = repo.Close()
		return nil, fmt.Errorf("ошибка инициализации обработчиков: %w", err)
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, h, cfg.Assets)

	return &App{
		repo:    repo,
		store:   repo,
		service: svc,
		handler: h,
		mux:     mux,
	}, nil
}

// Handler возвращает HTTP-обработчик с зарегистрированными маршрутами.
func (a *App) Handler() http.Handler {
	return a.mux
}

// Middleware оборачивает обработчик стандартным стеком мидлварей (RequestID, Security, CacheControl, MaxBodySize, PanicRecovery, Slog Logging).
func (a *App) Middleware(next http.Handler) http.Handler {
	return middleware.FullMiddlewareChain(func(w http.ResponseWriter, r *http.Request) {
		a.handler.Render500(w, r)
	})(next)
}

// Close закрывает открытые дескрипторы базы данных SQLite.
func (a *App) Close() error {
	return a.repo.Close()
}
