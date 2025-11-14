# Deployment Guide - CommitCaster на VPS

Полное руководство по деплою CommitCaster на VPS сервер.

## Требования

- VPS сервер с Ubuntu 20.04+ или Debian 10+
- Root или sudo доступ
- Доменное имя (для HTTPS)
- Минимум 1GB RAM, 1 CPU core, 10GB диска

## Быстрый старт (Автоматический деплой)

### 1. Подключитесь к VPS

```bash
ssh user@your-server-ip
```

### 2. Клонируйте репозиторий

```bash
cd ~
git clone https://github.com/yourusername/CommitCaster.git
cd CommitCaster
```

### 3. Запустите автоматический деплой

```bash
chmod +x deploy.sh
./deploy.sh
```

Скрипт автоматически:
- Обновит систему
- Установит Docker, Docker Compose, Nginx, Certbot
- Создаст .env файл (запросит ваши токены)
- Настроит Nginx
- Установит SSL сертификат
- Запустит бота в Docker

### 4. Настройте GitHub Webhook

После успешного деплоя, настройте webhook в вашем GitHub репозитории:

1. Откройте репозиторий на GitHub
2. **Settings** → **Webhooks** → **Add webhook**
3. Заполните:
   - **Payload URL**: `https://your-domain.com/webhook/github`
   - **Content type**: `application/json`
   - **Secret**: ваш `GITHUB_WEBHOOK_SECRET` из `.env`
   - **Which events**: Just the push event
4. Нажмите **Add webhook**

### 5. Готово!

Теперь при каждом push в репозиторий, бот будет автоматически публиковать пост в Telegram.

---

## Ручной деплой (Шаг за шагом)

Если вы хотите контролировать каждый шаг:

### Шаг 1: Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
sudo apt install -y docker.io
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo apt install -y docker-compose

# Установка Nginx
sudo apt install -y nginx
sudo systemctl enable nginx

# Установка Certbot для SSL
sudo apt install -y certbot python3-certbot-nginx
```

Перелогиньтесь для применения прав Docker:
```bash
exit
# Подключитесь снова
ssh user@your-server-ip
```

### Шаг 2: Клонирование проекта

```bash
cd ~
git clone https://github.com/yourusername/CommitCaster.git
cd CommitCaster
```

### Шаг 3: Настройка переменных окружения

Создайте файл `.env`:

```bash
nano .env
```

Вставьте:

```env
# Single-user режим
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_CHANNEL_ID=@yourchannel
GROQ_API_KEY=your_groq_api_key_here
GITHUB_WEBHOOK_SECRET=your_secret_here
PORT=8080
```

#### Где получить токены:

**TELEGRAM_BOT_TOKEN:**
1. Откройте [@BotFather](https://t.me/BotFather)
2. Отправьте `/newbot`
3. Следуйте инструкциям
4. Скопируйте токен

**TELEGRAM_CHANNEL_ID:**
1. Создайте публичный канал
2. Добавьте бота как администратора
3. Используйте `@yourchannel` или числовой ID

**GROQ_API_KEY:**
1. Зарегистрируйтесь на [console.groq.com](https://console.groq.com)
2. Перейдите в API Keys
3. Создайте новый ключ (бесплатно!)

**GITHUB_WEBHOOK_SECRET:**
- Придумайте любой случайный ключ, например: `openssl rand -hex 32`

### Шаг 4: Настройка Nginx

Создайте конфигурацию Nginx:

```bash
sudo nano /etc/nginx/sites-available/commitcaster
```

Вставьте (замените `your-domain.com`):

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Активируйте конфигурацию:

```bash
sudo ln -s /etc/nginx/sites-available/commitcaster /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Шаг 5: SSL сертификат

```bash
sudo certbot --nginx -d your-domain.com
```

Следуйте инструкциям certbot.

### Шаг 6: Запуск бота

```bash
docker-compose -f docker-compose.single.yml up -d --build
```

Проверьте статус:

```bash
docker-compose -f docker-compose.single.yml ps
docker-compose -f docker-compose.single.yml logs -f
```

Проверьте health endpoint:

```bash
curl http://localhost:8080/health
# Должен вернуть: {"status":"ok","service":"CommitCaster"}
```

---

## Альтернативный способ: Без Docker (Systemd)

Если вы предпочитаете запускать бота без Docker:

### 1. Установите Go

```bash
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Соберите бинарник

```bash
cd ~/CommitCaster
go build -o commitcaster cmd/bot/main.go
```

### 3. Создайте systemd сервис

```bash
sudo nano /etc/systemd/system/commitcaster.service
```

Вставьте (замените `YOUR_USERNAME` на ваше имя пользователя):

```ini
[Unit]
Description=CommitCaster Bot
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/home/YOUR_USERNAME/CommitCaster
ExecStart=/home/YOUR_USERNAME/CommitCaster/commitcaster
EnvironmentFile=/home/YOUR_USERNAME/CommitCaster/.env
Restart=always

[Install]
WantedBy=multi-user.target
```

### 4. Запустите сервис

```bash
sudo systemctl daemon-reload
sudo systemctl enable commitcaster
sudo systemctl start commitcaster
sudo systemctl status commitcaster
```

---

## Управление ботом

### Docker команды

```bash
# Просмотр логов
docker-compose -f docker-compose.single.yml logs -f

# Перезапуск
docker-compose -f docker-compose.single.yml restart

# Остановка
docker-compose -f docker-compose.single.yml down

# Обновление (после git pull)
docker-compose -f docker-compose.single.yml up -d --build

# Статус контейнера
docker-compose -f docker-compose.single.yml ps
```

### Systemd команды

```bash
# Просмотр логов
sudo journalctl -u commitcaster -f

# Перезапуск
sudo systemctl restart commitcaster

# Остановка
sudo systemctl stop commitcaster

# Статус
sudo systemctl status commitcaster

# Обновление (после пересборки)
sudo systemctl restart commitcaster
```

---

## Обновление бота

### Docker способ

```bash
cd ~/CommitCaster
git pull
docker-compose -f docker-compose.single.yml up -d --build
```

### Systemd способ

```bash
cd ~/CommitCaster
git pull
go build -o commitcaster cmd/bot/main.go
sudo systemctl restart commitcaster
```

---

## Мониторинг и отладка

### Проверка здоровья

```bash
# Health check endpoint
curl https://your-domain.com/health

# Swagger UI (API документация)
# Откройте в браузере: https://your-domain.com/swagger/index.html
```

### Просмотр логов

```bash
# Docker
docker-compose -f docker-compose.single.yml logs -f

# Systemd
sudo journalctl -u commitcaster -f -n 100

# Nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Проверка портов

```bash
# Проверка что порт 8080 слушается
sudo netstat -tulpn | grep 8080

# Проверка Nginx
sudo nginx -t
sudo systemctl status nginx
```

---

## Безопасность

### Firewall (UFW)

```bash
# Установка и настройка firewall
sudo apt install -y ufw

# Разрешаем SSH, HTTP, HTTPS
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'

# Включаем firewall
sudo ufw enable
sudo ufw status
```

### Автоматическое обновление SSL

Certbot автоматически обновляет сертификаты. Проверка:

```bash
# Тест обновления
sudo certbot renew --dry-run

# Автоматическое обновление уже настроено через systemd timer
sudo systemctl status certbot.timer
```

### Ротация логов

```bash
# Настройка ротации логов для Docker
sudo nano /etc/logrotate.d/commitcaster
```

Вставьте:

```
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    missingok
    delaycompress
    copytruncate
}
```

---

## Troubleshooting

### Ошибка: "Invalid signature"

**Проблема:** GitHub webhook возвращает ошибку подписи

**Решение:**
1. Проверьте что `GITHUB_WEBHOOK_SECRET` в `.env` совпадает с секретом в GitHub webhook
2. Перезапустите бота после изменения `.env`

```bash
docker-compose -f docker-compose.single.yml restart
```

### Ошибка: "telegram API error"

**Проблема:** Не удается отправить сообщение в Telegram

**Решение:**
1. Проверьте что бот добавлен как администратор канала
2. Проверьте правильность `TELEGRAM_CHANNEL_ID` (должен начинаться с `@` или быть числовым ID)
3. Проверьте `TELEGRAM_BOT_TOKEN`

### Ошибка: "connection refused"

**Проблема:** Nginx не может подключиться к приложению

**Решение:**
1. Проверьте что контейнер запущен: `docker-compose -f docker-compose.single.yml ps`
2. Проверьте логи: `docker-compose -f docker-compose.single.yml logs`
3. Проверьте что порт 8080 слушается: `curl http://localhost:8080/health`

### AI не отвечает

**Проблема:** Ошибки при генерации постов

**Решение:**
1. Проверьте `GROQ_API_KEY`
2. Проверьте квоты на [console.groq.com](https://console.groq.com)
3. Проверьте интернет соединение на сервере: `ping console.groq.com`

### Бот не получает webhook'и от GitHub

**Проблема:** GitHub отправляет webhook, но бот не реагирует

**Решение:**
1. Проверьте Recent Deliveries в GitHub webhook settings
2. Убедитесь что домен доступен извне: `curl https://your-domain.com/health`
3. Проверьте логи Nginx: `sudo tail -f /var/log/nginx/access.log`
4. Проверьте что SSL сертификат валидный: `curl -I https://your-domain.com`

---

## Полезные ссылки

- [GitHub Webhooks Documentation](https://docs.github.com/en/webhooks)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Groq Console](https://console.groq.com)
- [Docker Documentation](https://docs.docker.com/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt](https://letsencrypt.org/)

---

## Примеры endpoint'ов

После успешного деплоя доступны следующие endpoint'ы:

- `GET /health` - Проверка здоровья бота
- `POST /webhook/github` - GitHub webhook endpoint
- `GET /swagger/index.html` - Swagger UI документация

---

## Поддержка

Если у вас возникли проблемы:

1. Проверьте логи
2. Убедитесь что все переменные окружения установлены
3. Проверьте сетевое подключение
4. Откройте issue на GitHub

---

## Производительность

CommitCaster очень легковесный:

- **RAM**: ~20-50MB
- **CPU**: <1% при idle
- **Диск**: ~50MB (Docker образ)

Один VPS может легко обрабатывать тысячи webhook'ов в день.

---

Создано с использованием Claude Code
