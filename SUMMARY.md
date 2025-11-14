# CommitCaster - Итоговая документация

## Что было сделано

Создан **полнофункциональный backend** для CommitCaster с поддержкой двух режимов работы:

### ✅ 1. Single-User Mode (уже работает)
- Готов к использованию
- Работает с вашими токенами
- Webhook: `/webhook/github`

### ✅ 2. SaaS Mode (для публичного сервиса)
- Multi-user с PostgreSQL
- REST API для управления
- JWT аутентификация
- Уникальные webhook URLs
- Готов к интеграции с frontend

---

## Структура проекта

```
CommitCaster/
├── cmd/bot/
│   ├── main.go           # Single-user режим
│   └── main_saas.go      # SaaS режим (multi-user)
│
├── config/
│   └── config.go         # Загрузка .env
│
├── internal/
│   ├── models/
│   │   ├── github.go     # GitHub webhook модели
│   │   └── user.go       # User & UserSettings модели
│   │
│   ├── database/
│   │   └── database.go   # PostgreSQL подключение и миграции
│   │
│   ├── auth/
│   │   └── jwt.go        # JWT токены
│   │
│   ├── middleware/
│   │   └── auth.go       # JWT middleware
│   │
│   ├── handlers/
│   │   ├── webhook.go        # Single-user webhook handler
│   │   ├── webhook_multi.go  # Multi-user webhook handler
│   │   └── api.go            # REST API endpoints
│   │
│   └── services/
│       ├── telegram.go   # Telegram API
│       └── ai.go         # Groq AI
│
├── .env                  # Ваши текущие настройки (single-user)
├── .env.example          # Пример с обоими режимами
├── .env.saas             # Пример для SaaS режима
│
├── README.md             # Single-user документация
├── README_SAAS.md        # SaaS документация
├── API.md                # API reference
├── MODES.md              # Сравнение режимов
└── SUMMARY.md            # Этот файл
```

---

## Backend API готов!

### Endpoints для Frontend:

**Public (без аутентификации):**
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Логин
- `POST /webhook/github/:token` - GitHub webhook

**Protected (требуют JWT):**
- `GET /api/settings` - Получить настройки
- `PUT /api/settings` - Обновить настройки
- `GET /api/webhook` - Получить webhook URL

**См. полную документацию:** [API.md](./API.md)

---

## Что нужно сделать тебе (Frontend)

### Минимальный UI для MVP:

#### 1. Страница регистрации/логина
```javascript
// Login.jsx
POST /api/auth/login
{
  email: "user@example.com",
  password: "password"
}

// Response:
{
  token: "eyJhbGc...",  // Сохрани в localStorage
  webhook_token: "abc123...",
  webhook_url: "https://..."
}
```

#### 2. Dashboard (главная страница после логина)

**Поля для ввода:**
- Telegram Bot Token (текстовое поле)
- Telegram Channel ID (текстовое поле)
- Groq API Key (текстовое поле)
- GitHub Secret (опционально)

**Кнопка "Save Settings":**
```javascript
// SaveSettings()
PUT /api/settings
Headers: { Authorization: "Bearer <token>" }
Body: {
  telegram_bot_token: "...",
  telegram_channel_id: "@channel",
  groq_api_key: "gsk_..."
}
```

#### 3. Webhook URL Display

После сохранения настроек, показать:
```
Ваш Webhook URL:
https://your-domain.com/webhook/github/abc123...

[Copy to Clipboard]

Инструкция:
1. Откройте GitHub → Settings → Webhooks
2. Add webhook
3. Paste this URL
4. Secret: используйте ваш GitHub Secret
5. Events: Just the push event
```

### Технологии (на твой выбор):
- React + Tailwind CSS
- Vue + Element UI
- Plain HTML/CSS/JS
- Next.js

---

## Как запустить SaaS версию локально

### 1. Установи PostgreSQL

```bash
# macOS
brew install postgresql
brew services start postgresql
createdb commitcaster

# Ubuntu
sudo apt install postgresql
sudo -u postgres createdb commitcaster
```

### 2. Настрой .env

```env
DATABASE_URL=postgres://localhost:5432/commitcaster
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=http://localhost:8080
PORT=8080
```

### 3. Запусти backend

```bash
go run cmd/bot/main_saas.go
```

### 4. Протестируй API

```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "password": "password123",
    "name": "Test User"
  }'

# Получишь token в ответе
```

---

## Деплой на Heroku (SaaS)

```bash
# 1. Создай приложение
heroku create commitcaster-saas

# 2. Добавь PostgreSQL
heroku addons:create heroku-postgresql:mini

# 3. Установи переменные
heroku config:set JWT_SECRET=$(openssl rand -hex 32)
heroku config:set BASE_URL=https://commitcaster-saas.herokuapp.com

# 4. Деплой
# Замени main.go на main_saas.go
mv cmd/bot/main.go cmd/bot/main_old.go
mv cmd/bot/main_saas.go cmd/bot/main.go

git add .
git commit -m "Deploy SaaS"
git push heroku main
```

---

## База данных

### Таблицы (создаются автоматически):

**users:**
- id
- email (unique)
- password_hash
- name
- webhook_token (unique) - для GitHub webhook URL

**user_settings:**
- user_id
- telegram_bot_token
- telegram_channel_id
- groq_api_key
- github_secret
- is_active (auto true когда все токены заполнены)
- ai_model
- max_commits
- custom_prompt

---

## Пример интеграции Frontend

### React Dashboard Component

```jsx
import { useState, useEffect } from 'react';
import axios from 'axios';

const Dashboard = () => {
  const [settings, setSettings] = useState({});
  const [webhookUrl, setWebhookUrl] = useState('');
  const token = localStorage.getItem('auth_token');

  const api = axios.create({
    baseURL: 'https://your-domain.com/api',
    headers: { Authorization: `Bearer ${token}` }
  });

  useEffect(() => {
    // Загрузить настройки
    api.get('/settings').then(res => setSettings(res.data));

    // Получить webhook URL
    api.get('/webhook').then(res => setWebhookUrl(res.data.webhook_url));
  }, []);

  const handleSave = async () => {
    await api.put('/settings', settings);
    alert('Saved!');
  };

  return (
    <div>
      <h1>Settings</h1>

      <input
        placeholder="Telegram Bot Token"
        value={settings.telegram_bot_token || ''}
        onChange={e => setSettings({...settings, telegram_bot_token: e.target.value})}
      />

      <input
        placeholder="Telegram Channel ID"
        value={settings.telegram_channel_id || ''}
        onChange={e => setSettings({...settings, telegram_channel_id: e.target.value})}
      />

      <input
        placeholder="Groq API Key"
        value={settings.groq_api_key || ''}
        onChange={e => setSettings({...settings, groq_api_key: e.target.value})}
      />

      <button onClick={handleSave}>Save</button>

      <div>
        <h2>Webhook URL:</h2>
        <code>{webhookUrl}</code>
        <button onClick={() => navigator.clipboard.writeText(webhookUrl)}>
          Copy
        </button>
      </div>
    </div>
  );
};
```

---

## Безопасность

✅ **Реализовано:**
- bcrypt для паролей
- JWT аутентификация
- HMAC-SHA256 для GitHub webhooks
- CORS middleware

⚠️ **Рекомендации для продакшена:**
- Настрой CORS только для твоего домена
- Добавь rate limiting
- Используй HTTPS
- Email верификация (опционально)

---

## Что дальше?

### Backend готов! Теперь ты делаешь:

1. **Frontend** (React/Vue/etc)
   - Регистрация/логин
   - Dashboard с настройками
   - Отображение webhook URL

2. **Deploy Frontend**
   - Vercel / Netlify / GitHub Pages

3. **Свяжи всё вместе**
   - Frontend → Backend API
   - Backend деплоится на Heroku
   - PostgreSQL на Heroku

---

## Документация

- **[README.md](./README.md)** - Single-user режим (текущий рабочий)
- **[README_SAAS.md](./README_SAAS.md)** - SaaS режим (для публичного сервиса)
- **[API.md](./API.md)** - Полная API документация
- **[MODES.md](./MODES.md)** - Сравнение режимов
- **[QUICKSTART.md](./QUICKSTART.md)** - Быстрый старт single-user

---

## Вопросы?

Если что-то непонятно по backend API или нужна помощь с интеграцией - спрашивай!

Backend полностью готов к использованию 🚀
