#!/bin/bash

# Генерация Swagger документации

echo "🔨 Генерация Swagger документации..."

# Установка swag если не установлен
if ! command -v swag &> /dev/null; then
    echo "📦 Установка swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Генерация
swag init -g cmd/bot/main.go -o docs

echo "✅ Swagger документация сгенерирована в ./docs/"
echo "🌐 Запустите сервер и откройте: http://localhost:8080/swagger/index.html"
