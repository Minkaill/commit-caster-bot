#!/bin/bash

# CommitCaster Update Script
# Быстрое обновление бота на VPS

set -e

echo "🔄 CommitCaster Update Script"
echo "============================="
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Проверяем что мы в правильной директории
if [ ! -f "docker-compose.single.yml" ]; then
    echo "❌ Ошибка: docker-compose.single.yml не найден"
    echo "Запустите этот скрипт из директории CommitCaster"
    exit 1
fi

echo "📥 Шаг 1: Получение обновлений из Git..."
git fetch origin

# Проверяем есть ли обновления
LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u})

if [ $LOCAL = $REMOTE ]; then
    success "У вас уже последняя версия"
    read -p "Хотите пересобрать контейнер? (y/n): " REBUILD
    if [ "$REBUILD" != "y" ] && [ "$REBUILD" != "Y" ]; then
        echo "Выход без изменений"
        exit 0
    fi
else
    success "Найдены обновления"
    echo "Обновляем код..."
    git pull origin main
    success "Код обновлён"
fi

echo ""
echo "🐳 Шаг 2: Пересборка Docker контейнера..."

# Сохраняем старый контейнер на случай проблем
OLD_CONTAINER=$(docker ps -q -f name=commitcaster-app)

if [ ! -z "$OLD_CONTAINER" ]; then
    warning "Останавливаем старый контейнер..."
    docker-compose -f docker-compose.single.yml down
fi

# Собираем новый образ
echo "Сборка нового образа..."
docker-compose -f docker-compose.single.yml build --no-cache

# Запускаем обновлённый контейнер
echo "Запуск обновлённого контейнера..."
docker-compose -f docker-compose.single.yml up -d

success "Контейнер обновлён и запущен"

echo ""
echo "⏳ Ожидание запуска приложения..."
sleep 10

# Проверяем здоровье
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    success "Приложение работает!"
else
    echo "❌ Приложение не отвечает"
    echo "Проверьте логи: docker-compose -f docker-compose.single.yml logs"
    exit 1
fi

echo ""
echo "🧹 Шаг 3: Очистка старых образов..."
docker image prune -f
success "Старые образы удалены"

echo ""
echo "✨ Обновление завершено успешно!"
echo ""
echo "📋 Информация:"
echo "  - Версия: $(git rev-parse --short HEAD)"
echo "  - Ветка: $(git branch --show-current)"
echo "  - Статус: $(docker-compose -f docker-compose.single.yml ps --services)"
echo ""
echo "🔧 Полезные команды:"
echo "  - Просмотр логов: docker-compose -f docker-compose.single.yml logs -f"
echo "  - Перезапуск: docker-compose -f docker-compose.single.yml restart"
echo ""
