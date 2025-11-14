#!/bin/bash

# CommitCaster VPS Deployment Script
# Автоматическая настройка и деплой на VPS сервер

set -e

echo "🚀 CommitCaster Deployment Script"
echo "=================================="
echo ""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Проверка прав sudo
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}Не запускайте этот скрипт с sudo!${NC}"
    echo "Скрипт сам запросит sudo когда нужно"
    exit 1
fi

# Функция для вывода ошибок
error() {
    echo -e "${RED}❌ Ошибка: $1${NC}"
    exit 1
}

# Функция для вывода успеха
success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Функция для вывода предупреждений
warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Функция для проверки установлен ли пакет
is_installed() {
    command -v "$1" >/dev/null 2>&1
}

echo "📋 Шаг 1: Обновление системы..."
sudo apt update && sudo apt upgrade -y
success "Система обновлена"

echo ""
echo "📦 Шаг 2: Установка зависимостей..."

# Установка Docker
if ! is_installed docker; then
    echo "Установка Docker..."
    sudo apt install -y docker.io
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker $USER
    success "Docker установлен"
else
    success "Docker уже установлен"
fi

# Установка Docker Compose
if ! is_installed docker-compose; then
    echo "Установка Docker Compose..."
    sudo apt install -y docker-compose
    success "Docker Compose установлен"
else
    success "Docker Compose уже установлен"
fi

# Установка Nginx
if ! is_installed nginx; then
    echo "Установка Nginx..."
    sudo apt install -y nginx
    sudo systemctl enable nginx
    sudo systemctl start nginx
    success "Nginx установлен"
else
    success "Nginx уже установлен"
fi

# Установка Certbot для SSL
if ! is_installed certbot; then
    echo "Установка Certbot для SSL..."
    sudo apt install -y certbot python3-certbot-nginx
    success "Certbot установлен"
else
    success "Certbot уже установлен"
fi

echo ""
echo "🔧 Шаг 3: Настройка переменных окружения..."

# Проверяем существует ли .env файл
if [ ! -f .env ]; then
    echo "Создание .env файла..."

    # Запрашиваем данные у пользователя
    read -p "Введите TELEGRAM_BOT_TOKEN: " TELEGRAM_BOT_TOKEN
    read -p "Введите TELEGRAM_CHANNEL_ID (например @yourchannel): " TELEGRAM_CHANNEL_ID
    read -p "Введите GROQ_API_KEY: " GROQ_API_KEY
    read -p "Введите GITHUB_WEBHOOK_SECRET: " GITHUB_WEBHOOK_SECRET

    # Создаём .env файл
    cat > .env << EOF
# Single-user режим (для личного использования)
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
TELEGRAM_CHANNEL_ID=${TELEGRAM_CHANNEL_ID}
GROQ_API_KEY=${GROQ_API_KEY}
GITHUB_WEBHOOK_SECRET=${GITHUB_WEBHOOK_SECRET}
PORT=8080
EOF

    success ".env файл создан"
else
    success ".env файл уже существует"
fi

echo ""
echo "🌐 Шаг 4: Настройка Nginx..."

read -p "Введите ваш домен (например: commitcaster.example.com): " DOMAIN

# Создаём конфигурацию Nginx
sudo tee /etc/nginx/sites-available/commitcaster > /dev/null << EOF
server {
    listen 80;
    server_name ${DOMAIN};

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

# Активируем конфигурацию
sudo ln -sf /etc/nginx/sites-available/commitcaster /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

success "Nginx настроен для домена ${DOMAIN}"

echo ""
echo "🔒 Шаг 5: Настройка SSL сертификата..."

read -p "Хотите установить SSL сертификат? (y/n): " INSTALL_SSL

if [ "$INSTALL_SSL" = "y" ] || [ "$INSTALL_SSL" = "Y" ]; then
    read -p "Введите email для Let's Encrypt: " EMAIL
    sudo certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos -m ${EMAIL}
    success "SSL сертификат установлен"
else
    warning "SSL сертификат не установлен. Вебхук будет работать только по HTTP (не рекомендуется для продакшена)"
fi

echo ""
echo "🐳 Шаг 6: Сборка и запуск Docker контейнера..."

# Останавливаем старый контейнер если есть
docker-compose -f docker-compose.single.yml down 2>/dev/null || true

# Собираем и запускаем новый
docker-compose -f docker-compose.single.yml up -d --build

success "Docker контейнер запущен"

echo ""
echo "⏳ Ожидание запуска приложения..."
sleep 10

# Проверяем здоровье приложения
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    success "Приложение работает!"
else
    error "Приложение не отвечает. Проверьте логи: docker-compose -f docker-compose.single.yml logs"
fi

echo ""
echo "✨ Деплой завершён успешно!"
echo ""
echo "📋 Информация:"
echo "  - Домен: https://${DOMAIN}"
echo "  - Webhook URL: https://${DOMAIN}/webhook/github"
echo "  - Health check: https://${DOMAIN}/health"
echo "  - Swagger UI: https://${DOMAIN}/swagger/index.html"
echo ""
echo "📝 Следующие шаги:"
echo "  1. Настройте GitHub Webhook:"
echo "     - Откройте репозиторий на GitHub"
echo "     - Settings → Webhooks → Add webhook"
echo "     - Payload URL: https://${DOMAIN}/webhook/github"
echo "     - Content type: application/json"
echo "     - Secret: ваш GITHUB_WEBHOOK_SECRET из .env"
echo "     - Events: Just the push event"
echo ""
echo "  2. Добавьте бота как администратора вашего Telegram канала"
echo ""
echo "🔧 Полезные команды:"
echo "  - Просмотр логов: docker-compose -f docker-compose.single.yml logs -f"
echo "  - Перезапуск: docker-compose -f docker-compose.single.yml restart"
echo "  - Остановка: docker-compose -f docker-compose.single.yml down"
echo "  - Обновление: git pull && docker-compose -f docker-compose.single.yml up -d --build"
echo ""
echo "=================================="
echo "🎉 Готово! Бот работает!"
