@echo off
REM Генерация Swagger документации для Windows

echo 🔨 Генерация Swagger документации...

REM Проверка установки swag
where swag >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 📦 Установка swag...
    go install github.com/swaggo/swag/cmd/swag@latest
)

REM Генерация
swag init -g cmd/bot/main.go -o docs

echo ✅ Swagger документация сгенерирована в ./docs/
echo 🌐 Запустите сервер и откройте: http://localhost:8080/swagger/index.html
pause
