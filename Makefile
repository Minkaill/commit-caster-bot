# CommitCaster Makefile

.PHONY: help run build clean test install

help: ## Показать это сообщение
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

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
