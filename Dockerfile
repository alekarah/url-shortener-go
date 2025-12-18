# Этап 1: Сборка
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Этап 2: Production образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/main .

# Копируем статические файлы
COPY --from=builder /app/static ./static

# Копируем миграции (если нужны)
COPY --from=builder /app/migrations ./migrations

# Expose порт (Render использует переменную PORT)
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
