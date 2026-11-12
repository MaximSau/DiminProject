# --- СТАДИЯ 1: Сборка статического бинарника на Go ---
FROM golang:alpine AS builder

WORKDIR /build

# Установка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка чистого Go бинарника без зависимостей CGO (pure-Go SQLite)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /pulsefit main.go

# --- СТАДИЯ 2: Финальный легковесный образ для продакшена ---
FROM alpine:3.21

# Установка корневых сертификатов для внешних HTTPS-запросов и часовых поясов
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Создание каталога для персистентной базы данных SQLite
RUN mkdir -p /data && chmod 777 /data

# Копирование скомпилированного бинарника
COPY --from=builder /pulsefit /app/pulsefit

# Переменные окружения по умолчанию
ENV PULSEFIT_ADDR="0.0.0.0:8099"
ENV PULSEFIT_DATA="/data"

EXPOSE 8099

# Проверка работоспособности сервиса (Healthcheck)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://127.0.0.1:8099/ || exit 1

VOLUME ["/data"]

ENTRYPOINT ["/app/pulsefit"]
