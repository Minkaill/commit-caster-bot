# CommitCaster SaaS - Multi-User Guide

Это документация для развёртывания CommitCaster как публичного SaaS сервиса.

## Что изменилось в SaaS версии?

### Новые возможности:

1. **Multi-user поддержка** - множество пользователей с отдельными настройками
2. **REST API** для управления через веб-интерфейс
3. **JWT аутентификация** для безопасного доступа
4. **PostgreSQL** для хранения пользователей и настроек
5. **Уникальные webhook URLs** для каждого пользователя
6. **Кастомизация промптов** и настроек AI

### Архитектура:

```
┌─────────────────┐
│   Frontend      │  (Ваш веб-интерфейс)
│  (React/Vue)    │
└────────┬────────┘
         │ REST API
         ↓
┌─────────────────┐
│  CommitCaster   │  (Backend на Golang)
│     Backend     │
└────────┬────────┘
         │
    ┌────┴─────┐
    ↓          ↓
┌────────┐  ┌──────────┐
│Postgre │  │ External │
│   SQL  │  │   APIs   │
└────────┘  └──────────┘
             - Telegram
             - Groq
             - GitHub
```

---

## Быстрый старт (SaaS)

### 1. Установите PostgreSQL

**Local (для разработки):**
```bash
# macOS
brew install postgresql
brew services start postgresql

# Ubuntu/Debian
sudo apt-get install postgresql
sudo systemctl start postgresql

# Windows
# Скачайте с https://www.postgresql.org/download/
```

Создайте базу данных:
```bash
psql postgres
CREATE DATABASE commitcaster;
\q
```

### 2. Настройте переменные окружения

Создайте `.env`:
```env
# PostgreSQL (обязательно для SaaS)
DATABASE_URL=postgres://user:password@localhost:5432/commitcaster

# JWT Secret (обязательно для SaaS)
JWT_SECRET=your_random_secret_key_here

# Base URL вашего приложения
BASE_URL=https://your-domain.com

# Port
PORT=8080
```

Сгенерируйте JWT_SECRET:
```bash
openssl rand -hex 32
```

### 3. Запустите backend

```bash
go run cmd/bot/main_saas.go
```

Вывод:
```
🌐 Starting in SaaS mode (multi-user)
✅ Connected to PostgreSQL
Running database migrations...
✅ Database migrations completed
📋 API Endpoints:
  POST /api/auth/register - Register new user
  POST /api/auth/login - Login
  ...
🚀 CommitCaster запущен на порту 8080
```

### 4. Протестируйте API

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Получите токен из ответа и используйте его:
TOKEN="your_jwt_token"

# Обновите настройки
curl -X PUT http://localhost:8080/api/settings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "telegram_bot_token": "...",
    "telegram_channel_id": "@mychannel",
    "groq_api_key": "gsk_..."
  }'
```

---

## Деплой на Heroku (SaaS)

### 1. Создайте приложение

```bash
heroku create commitcaster-saas
```

### 2. Добавьте PostgreSQL

```bash
heroku addons:create heroku-postgresql:mini
```

Heroku автоматически установит `DATABASE_URL`.

### 3. Установите переменные

```bash
# Сгенерируйте JWT secret
JWT_SECRET=$(openssl rand -hex 32)
heroku config:set JWT_SECRET=$JWT_SECRET

# Base URL
heroku config:set BASE_URL=https://commitcaster-saas.herokuapp.com

# Port (Heroku установит автоматически, но можно задать)
heroku config:set PORT=8080
```

### 4. Задеплойте

```bash
# Замените main.go на main_saas.go
mv cmd/bot/main.go cmd/bot/main_old.go
mv cmd/bot/main_saas.go cmd/bot/main.go

git add .
git commit -m "Deploy SaaS version"
git push heroku main
```

### 5. Проверьте

```bash
heroku logs --tail

# Откройте в браузере
open https://commitcaster-saas.herokuapp.com/health
```

---

## Frontend интеграция

Теперь вам нужно создать веб-интерфейс. Вот что он должен делать:

### Минимальный UI (MVP):

1. **Страница регистрации/логина**
   - Email, пароль, имя
   - Кнопки "Sign Up" и "Login"

2. **Dashboard (после логина)**
   - Поля для ввода токенов:
     - Telegram Bot Token
     - Telegram Channel ID
     - Groq API Key
     - GitHub Secret (optional)
   - Кнопка "Save Settings"

3. **Webhook URL display**
   - Показать уникальный webhook URL
   - Копировать в буфер обмена
   - Инструкция как добавить в GitHub

### Пример React компонента:

```jsx
// Dashboard.jsx
import { useState, useEffect } from 'react';
import axios from 'axios';

const API_URL = 'https://your-domain.com/api';

function Dashboard() {
  const [settings, setSettings] = useState({});
  const [webhookUrl, setWebhookUrl] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('auth_token');

    // Загрузить настройки
    axios.get(`${API_URL}/settings`, {
      headers: { Authorization: `Bearer ${token}` }
    }).then(res => setSettings(res.data));

    // Получить webhook URL
    axios.get(`${API_URL}/webhook`, {
      headers: { Authorization: `Bearer ${token}` }
    }).then(res => setWebhookUrl(res.data.webhook_url));
  }, []);

  const handleSave = async () => {
    const token = localStorage.getItem('auth_token');

    await axios.put(`${API_URL}/settings`, settings, {
      headers: { Authorization: `Bearer ${token}` }
    });

    alert('Settings saved!');
  };

  return (
    <div>
      <h1>CommitCaster Dashboard</h1>

      <div>
        <label>Telegram Bot Token</label>
        <input
          value={settings.telegram_bot_token || ''}
          onChange={e => setSettings({...settings, telegram_bot_token: e.target.value})}
        />
      </div>

      <div>
        <label>Telegram Channel ID</label>
        <input
          value={settings.telegram_channel_id || ''}
          onChange={e => setSettings({...settings, telegram_channel_id: e.target.value})}
        />
      </div>

      <div>
        <label>Groq API Key</label>
        <input
          value={settings.groq_api_key || ''}
          onChange={e => setSettings({...settings, groq_api_key: e.target.value})}
        />
      </div>

      <button onClick={handleSave}>Save Settings</button>

      <div>
        <h2>Your Webhook URL:</h2>
        <code>{webhookUrl}</code>
        <button onClick={() => navigator.clipboard.writeText(webhookUrl)}>
          Copy
        </button>
      </div>
    </div>
  );
}
```

---

## API Reference

См. полную документацию в [API.md](./API.md)

**Основные endpoints:**
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Логин
- `GET /api/settings` - Получить настройки (protected)
- `PUT /api/settings` - Обновить настройки (protected)
- `GET /api/webhook` - Получить webhook URL (protected)
- `POST /webhook/github/:token` - GitHub webhook

---

## Database Schema

### Table: users
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  email VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  name VARCHAR,
  webhook_token VARCHAR UNIQUE NOT NULL
);
```

### Table: user_settings
```sql
CREATE TABLE user_settings (
  id SERIAL PRIMARY KEY,
  user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  telegram_bot_token VARCHAR,
  telegram_channel_id VARCHAR,
  groq_api_key VARCHAR,
  github_secret VARCHAR,
  is_active BOOLEAN DEFAULT false,
  ai_model VARCHAR DEFAULT 'llama-3.3-70b-versatile',
  post_language VARCHAR DEFAULT 'ru',
  max_commits INTEGER DEFAULT 5,
  custom_prompt TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

Миграции выполняются автоматически при запуске (GORM AutoMigrate).

---

## Безопасность

1. **Пароли** - хешируются с bcrypt
2. **JWT токены** - срок действия 7 дней
3. **GitHub webhooks** - проверяются HMAC-SHA256 подписью
4. **CORS** - настройте в `main_saas.go` для вашего домена

**Рекомендации:**
- Используйте HTTPS в продакшене
- Настройте rate limiting
- Добавьте email верификацию
- Логируйте подозрительную активность

---

## Мониторинг

### Health Check

```bash
curl https://your-domain.com/health
```

Response:
```json
{
  "status": "ok",
  "mode": "saas"
}
```

### Logs

```bash
# Heroku
heroku logs --tail

# Local
# Логи выводятся в консоль
```

---

## Troubleshooting

**Database connection failed**
```
Проверьте DATABASE_URL. Формат: postgres://user:password@host:5432/database
```

**JWT_SECRET not set**
```
Обязательно установите JWT_SECRET для SaaS режима
```

**Миграции не применяются**
```
Проверьте права пользователя БД. Должен иметь CREATE TABLE права.
```

**CORS ошибки**
```
Настройте CORS в main_saas.go для вашего frontend домена
```

---

## Масштабирование

### Redis для сессий (опционально)

Добавьте Redis для хранения сессий вместо JWT:

```bash
heroku addons:create heroku-redis:mini
```

### Load Balancing

Используйте несколько инстансов:

```bash
heroku ps:scale web=3
```

### CDN

Для статики фронтенда используйте CloudFlare или аналог.

---

## Roadmap

Будущие улучшения:
- [ ] Email верификация
- [ ] Rate limiting
- [ ] Webhook retry mechanism
- [ ] Аналитика (сколько постов, статистика)
- [ ] Поддержка нескольких каналов
- [ ] Webhook для других событий GitHub (PR, Issues)
- [ ] Web UI (готовый фронтенд)

---

## Лицензия

MIT

## Поддержка

Создавайте issues в GitHub репозитории.
