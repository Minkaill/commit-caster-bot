# CommitCaster Makefile

.PHONY: help run build clean test install deploy update logs

help: ## Показать это сообщение
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Установить зависимости
	go mod download
	go mod tidy

run: ## Запустить бота
	go run cmd/bot/main.go

build: ## Собрать бинарник
	go build -o bin/commitcaster cmd/bot/main.go

build-windows: ## Собрать для Windows
	GOOS=windows GOARCH=amd64 go build -o bin/commitcaster.exe cmd/bot/main.go

build-linux: ## Собрать для Linux
	GOOS=linux GOARCH=amd64 go build -o bin/commitcaster cmd/bot/main.go

clean: ## Удалить собранные файлы
	rm -rf bin/

test: ## Запустить тесты
	go test -v ./...

dev: ## Режим разработки (с hot reload, требует установки air)
	air

swagger: ## Генерация Swagger документации
	swag init -g cmd/bot/main.go -o docs
	@echo "📖 Swagger UI: http://localhost:8080/swagger/index.html"

# === Docker команды ===

docker-build: ## Собрать Docker образ
	docker-compose -f docker-compose.single.yml build

docker-up: ## Запустить в Docker
	docker-compose -f docker-compose.single.yml up -d

docker-down: ## Остановить Docker контейнер
	docker-compose -f docker-compose.single.yml down

docker-restart: ## Перезапустить Docker контейнер
	docker-compose -f docker-compose.single.yml restart

docker-logs: ## Показать логи Docker контейнера
	docker-compose -f docker-compose.single.yml logs -f

docker-ps: ## Статус Docker контейнера
	docker-compose -f docker-compose.single.yml ps

# === Deployment команды ===

deploy: ## Запустить автоматический деплой на VPS
	@chmod +x deploy.sh
	./deploy.sh

update: ## Обновить бота на VPS
	@chmod +x update.sh
	./update.sh

logs: ## Показать логи (Docker или systemd)
	@if [ -f "docker-compose.single.yml" ] && docker-compose -f docker-compose.single.yml ps -q 2>/dev/null | grep -q .; then \
		docker-compose -f docker-compose.single.yml logs -f; \
	else \
		sudo journalctl -u commitcaster -f -n 100; \
	fi

health: ## Проверить здоровье бота
	@curl -f http://localhost:8080/health || echo "Бот не отвечает"

# === Утилиты ===

check-env: ## Проверить переменные окружения
	@echo "Проверка .env файла..."
	@if [ ! -f .env ]; then \
		echo "❌ .env файл не найден"; \
		exit 1; \
	fi
	@echo "✅ .env файл существует"
	@echo ""
	@echo "Проверка обязательных переменных:"
	@grep -q "TELEGRAM_BOT_TOKEN=" .env && echo "✅ TELEGRAM_BOT_TOKEN" || echo "❌ TELEGRAM_BOT_TOKEN отсутствует"
	@grep -q "TELEGRAM_CHANNEL_ID=" .env && echo "✅ TELEGRAM_CHANNEL_ID" || echo "❌ TELEGRAM_CHANNEL_ID отсутствует"
	@grep -q "GROQ_API_KEY=" .env && echo "✅ GROQ_API_KEY" || echo "❌ GROQ_API_KEY отсутствует"
	@grep -q "GITHUB_WEBHOOK_SECRET=" .env && echo "✅ GITHUB_WEBHOOK_SECRET" || echo "❌ GITHUB_WEBHOOK_SECRET отсутствует"

setup-env: ## Создать .env из примера
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ .env файл создан из .env.example"; \
		echo "⚠️  Отредактируйте .env и добавьте ваши токены"; \
	else \
		echo "⚠️  .env файл уже существует"; \
	fi
