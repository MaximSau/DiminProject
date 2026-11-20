.PHONY: all build run test coverage vet docker-build docker-up docker-down clean help build-frontend

APP_NAME = pulsefit
BIN_DIR = bin

# Определение установленной утилиты контейнеризации (docker или podman)
DOCKER_CMD := $(shell which docker 2>/dev/null || which podman 2>/dev/null)
COMPOSE_CMD := $(shell which docker-compose 2>/dev/null || (docker compose version >/dev/null 2>&1 && echo "docker compose") || which podman-compose 2>/dev/null)

all: vet test build

## help: Вывести список доступных команд
help:
	@echo "Доступные команды:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## run: Быстрый локальный запуск на Go (сборка React SPA фронтенда + go run)
run: build-frontend
	go run main.go

## build-frontend: Сборка React TypeScript SPA (Vite + Tailwind) в web/dist
build-frontend:
	@if which npm >/dev/null 2>&1; then \
		echo "⚡ Сборка React SPA фронтенда..."; \
		cd frontend && npm run build; \
	else \
		echo "💡 Node.js/npm не обнаружен, используется предварительно скомпилированный web/dist"; \
	fi

## build: Сборка статического бинарника в папку bin/ (включая SPA фронтенд)
build: build-frontend
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) main.go
	@echo "Бинарник успешно собран: $(BIN_DIR)/$(APP_NAME)"

## start: Сборка и прямой запуск скомпилированного бинарника
start: build
	./$(BIN_DIR)/$(APP_NAME)

## test: Запуск всех модульных и интеграционных тестов с проверкой гонок (race)
test:
	go test -v -race ./...

## coverage: Генерация отчёта о покрытии кода тестами
coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo "Детальный HTML отчёт можно открыть командой: go tool cover -html=coverage.out"

## vet: Запуск статического анализа кода go vet
vet:
	go vet ./...

## docker-build: Сборка оптимизированного Docker-образа
docker-build:
	@if [ -z "$(DOCKER_CMD)" ]; then \
		echo "❌ Ошибка: Docker не установлен в системе."; \
		echo "💡 Для запуска без Docker используйте команду: make run"; \
		echo "💡 Инструкция по установке Docker приведена в README.md"; \
		exit 1; \
	fi
	$(DOCKER_CMD) build -t $(APP_NAME):latest .

## docker-up: Запуск контейнеров в фоновом режиме через docker compose
docker-up:
	@if [ -z "$(COMPOSE_CMD)" ]; then \
		echo "❌ Ошибка: docker compose не найден в системе."; \
		echo "💡 Вы можете запустить PulseFit прямо сейчас БЕЗ Docker командой:"; \
		echo "     make run"; \
		echo "   или:"; \
		echo "     make start"; \
		echo "💡 Чтобы установить Docker в Fedora Linux, выполните:"; \
		echo "     sudo dnf install docker-ce docker-compose-plugin"; \
		echo "     sudo systemctl enable --now docker"; \
		exit 1; \
	fi
	$(COMPOSE_CMD) up -d

## docker-down: Остановка контейнеров docker compose
docker-down:
	@if [ -n "$(COMPOSE_CMD)" ]; then \
		$(COMPOSE_CMD) down; \
	else \
		echo "docker compose не найден."; \
	fi

## clean: Очистка артефактов сборки и временных файлов
clean:
	rm -rf $(BIN_DIR) test_* cookies.txt coverage.out
	@echo "Артефакты очищены."
