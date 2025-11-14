# CommitCaster API Documentation

## Режимы работы

CommitCaster поддерживает два режима:

1. **Single-User Mode** - персональное использование (без БД)
2. **SaaS Mode** - публичный сервис для множества пользователей (с PostgreSQL)

Режим определяется автоматически по наличию переменной `DATABASE_URL`.

---

## SaaS Mode API

Для использования в SaaS режиме:

### Base URL
```
https://your-domain.com
```

### Authentication

Используется JWT токен. Получите токен через `/api/auth/register` или `/api/auth/login`.

Все защищённые endpoints требуют заголовок:
```
Authorization: Bearer <your_jwt_token>
```

---

## Endpoints

### 1. Регистрация пользователя

**POST** `/api/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword",
  "name": "Your Name"
}
```

**Response:** `201 Created`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "webhook_token": "a1b2c3d4e5f6...",
  "webhook_url": "https://your-domain.com/webhook/github/a1b2c3d4e5f6..."
}
```

**Errors:**
- `400` - Invalid request data
- `409` - User already exists

---

### 2. Логин

**POST** `/api/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "webhook_token": "a1b2c3d4e5f6...",
  "webhook_url": "https://your-domain.com/webhook/github/a1b2c3d4e5f6..."
}
```

**Errors:**
- `400` - Invalid request data
- `401` - Invalid credentials

---

### 3. Получить настройки (Protected)

**GET** `/api/settings`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 1,
  "telegram_bot_token": "1234567890:ABC...",
  "telegram_channel_id": "@mychannel",
  "groq_api_key": "gsk_...",
  "github_secret": "my_secret",
  "is_active": true,
  "ai_model": "llama-3.3-70b-versatile",
  "post_language": "ru",
  "max_commits": 5,
  "custom_prompt": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `401` - Unauthorized
- `404` - Settings not found

---

### 4. Обновить настройки (Protected)

**PUT** `/api/settings`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Request Body:**
```json
{
  "telegram_bot_token": "1234567890:ABC...",
  "telegram_channel_id": "@mychannel",
  "groq_api_key": "gsk_...",
  "github_secret": "my_webhook_secret",
  "ai_model": "llama-3.3-70b-versatile",
  "post_language": "ru",
  "max_commits": 5,
  "custom_prompt": "Custom prompt with %s for commits and %s for repo name"
}
```

Все поля опциональные. Отправляйте только те, которые хотите обновить.

**Response:** `200 OK`
```json
{
  "id": 1,
  "user_id": 1,
  "telegram_bot_token": "1234567890:ABC...",
  "telegram_channel_id": "@mychannel",
  "groq_api_key": "gsk_...",
  "github_secret": "my_secret",
  "is_active": true,
  "ai_model": "llama-3.3-70b-versatile",
  "post_language": "ru",
  "max_commits": 5,
  "custom_prompt": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Note:** Бот автоматически становится активным (`is_active: true`) когда все обязательные токены заполнены.

**Errors:**
- `400` - Invalid request data
- `401` - Unauthorized
- `404` - Settings not found

---

### 5. Получить информацию о webhook (Protected)

**GET** `/api/webhook`

**Headers:**
```
Authorization: Bearer <your_jwt_token>
```

**Response:** `200 OK`
```json
{
  "webhook_token": "a1b2c3d4e5f6...",
  "webhook_url": "https://your-domain.com/webhook/github/a1b2c3d4e5f6..."
}
```

**Errors:**
- `401` - Unauthorized
- `404` - User not found

---

### 6. GitHub Webhook Endpoint (Public)

**POST** `/webhook/github/:token`

Этот endpoint вызывается GitHub'ом автоматически при push событиях.

**Path Parameters:**
- `token` - уникальный webhook токен пользователя

**Headers от GitHub:**
```
X-GitHub-Event: push
X-Hub-Signature-256: sha256=...
Content-Type: application/json
```

**Request Body:** GitHub Push Event Payload

**Response:** `200 OK`
```json
{
  "message": "Webhook received"
}
```

**Errors:**
- `400` - Invalid request
- `401` - Invalid signature
- `403` - Bot not active (tokens not configured)
- `404` - Invalid webhook token

---

## Workflow для Frontend

### 1. Регистрация/Логин

```javascript
// Регистрация
const registerResponse = await fetch('https://your-domain.com/api/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123',
    name: 'John Doe'
  })
});

const { token, webhook_url } = await registerResponse.json();

// Сохраните token в localStorage
localStorage.setItem('auth_token', token);
```

### 2. Настройка токенов

```javascript
const token = localStorage.getItem('auth_token');

await fetch('https://your-domain.com/api/settings', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    telegram_bot_token: '...',
    telegram_channel_id: '@mychannel',
    groq_api_key: 'gsk_...',
    github_secret: 'my_secret'
  })
});
```

### 3. Получение webhook URL

```javascript
const webhookInfo = await fetch('https://your-domain.com/api/webhook', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const { webhook_url } = await webhookInfo.json();
console.log('Use this URL in GitHub webhook settings:', webhook_url);
```

### 4. Настройка в GitHub

1. Откройте репозиторий → Settings → Webhooks → Add webhook
2. Payload URL: `webhook_url` из предыдущего шага
3. Content type: `application/json`
4. Secret: значение из поля `github_secret`
5. Events: Just the push event

---

## Настройки (UserSettings)

### Обязательные поля:
- `telegram_bot_token` - токен Telegram бота
- `telegram_channel_id` - ID или @username канала
- `groq_api_key` - API ключ Groq

Бот станет активным только после заполнения всех трёх полей.

### Опциональные поля:

**github_secret** (string)
- Секретный ключ для проверки подписи GitHub webhook
- Рекомендуется установить для безопасности

**ai_model** (string, default: `"llama-3.3-70b-versatile"`)
- Модель AI для генерации постов
- Доступные модели на Groq:
  - `llama-3.3-70b-versatile`
  - `llama-3.1-70b-versatile`
  - `mixtral-8x7b-32768`

**post_language** (string, default: `"ru"`)
- Язык постов (пока не используется, для будущего функционала)

**max_commits** (int, default: `5`)
- Максимальное количество коммитов в одном посте

**custom_prompt** (string, optional)
- Кастомный промпт для AI
- Должен содержать два `%s` для подстановки:
  1. Информация о коммитах
  2. Название репозитория

Пример кастомного промпта:
```
Ты - Senior разработчик. Информация о коммитах: %s. Напиши краткий технический пост для проекта %s в стиле tech lead. Без эмодзи.
```

---

## Error Responses

Все ошибки возвращаются в формате:
```json
{
  "error": "Описание ошибки"
}
```

Коды ошибок:
- `400` - Bad Request (неверные данные)
- `401` - Unauthorized (нет токена или он невалидный)
- `403` - Forbidden (доступ запрещён)
- `404` - Not Found (ресурс не найден)
- `409` - Conflict (конфликт, например email уже существует)
- `500` - Internal Server Error (серверная ошибка)

---

## Environment Variables (Backend)

Для SaaS режима необходимо установить:

```env
# PostgreSQL
DATABASE_URL=postgres://user:password@host:5432/commitcaster

# JWT для аутентификации
JWT_SECRET=your_random_secret_key_here

# Base URL приложения
BASE_URL=https://your-domain.com

# Порт
PORT=8080
```

---

## Security

1. **JWT токены** действительны 7 дней
2. **Пароли** хешируются с bcrypt
3. **GitHub webhooks** подписываются HMAC-SHA256
4. **CORS** настроен для всех доменов (настройте под себя в продакшене)

---

## Rate Limits

Пока не реализованы. Рекомендуется добавить в продакшене.

---

## Примеры интеграции

### React + Axios

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'https://your-domain.com/api'
});

// Interceptor для добавления токена
api.interceptors.request.use(config => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Регистрация
await api.post('/auth/register', {
  email: 'user@example.com',
  password: 'password123',
  name: 'John Doe'
});

// Получить настройки
const settings = await api.get('/settings');

// Обновить настройки
await api.put('/settings', {
  telegram_bot_token: '...',
  telegram_channel_id: '@mychannel'
});
```

---

Для вопросов и поддержки создавайте issues в GitHub репозитории.
