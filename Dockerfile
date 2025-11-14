# Multi-stage build для минимального образа

# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o commitcaster cmd/bot/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированное приложение из builder
COPY --from=builder /app/commitcaster .

# Копируем docs для Swagger (если есть)
COPY --from=builder /app/docs ./docs

# Порт приложения
EXPOSE 8080

# Запуск
CMD ["./commitcaster"]
