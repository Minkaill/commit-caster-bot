# CommitCaster - Режимы работы

CommitCaster поддерживает два режима работы:

## 1. Single-User Mode (Персональное использование)

**Для кого:** Один пользователь, личный блог

**Что нужно:**
- Telegram Bot Token
- Telegram Channel ID
- Groq API Key
- (опционально) GitHub Secret

**Файл настроек:** `.env`

**Запуск:**
```bash
go run cmd/bot/main.go
```

**Webhook URL:**
```
http://your-domain.com/webhook/github
```

---

## 2. SaaS Mode (Публичный сервис)

**Для кого:** Множество пользователей, каждый со своими настройками

**Что нужно:**
- PostgreSQL база данных
- JWT Secret
- Base URL

Пользователи сами вводят свои токены через веб-интерфейс.

**Файл настроек:** `.env` с `DATABASE_URL`

**Запуск:**
```bash
go run cmd/bot/main_saas.go
```

**Webhook URL (для каждого пользователя):**
```
http://your-domain.com/webhook/github/{unique_token}
```

---

## Как переключиться между режимами?

### Single-User → SaaS

1. Установите PostgreSQL
2. Добавьте в `.env`:
   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/commitcaster
   JWT_SECRET=your_random_secret
   BASE_URL=https://your-domain.com
   ```
3. Запустите `main_saas.go` вместо `main.go`
4. Создайте веб-интерфейс для пользователей

### SaaS → Single-User

1. Удалите или закомментируйте `DATABASE_URL` в `.env`
2. Добавьте в `.env`:
   ```env
   TELEGRAM_BOT_TOKEN=...
   TELEGRAM_CHANNEL_ID=...
   GROQ_API_KEY=...
   ```
3. Запустите `main.go` вместо `main_saas.go`

---

## Сравнение

| Параметр | Single-User | SaaS |
|----------|-------------|------|
| Пользователи | 1 | Множество |
| База данных | Не нужна | PostgreSQL |
| Аутентификация | Нет | JWT |
| API | Нет | Да |
| Веб-интерфейс | Не нужен | Нужен |
| Деплой | Простой | Средней сложности |
| Webhook URL | Один общий | Уникальный для каждого |
| Настройки | .env файл | В БД через API |

---

## Когда использовать какой режим?

**Single-User подходит если:**
- Вы используете бота только для себя
- У вас один Telegram канал
- Не нужен веб-интерфейс
- Хотите простоту настройки

**SaaS подходит если:**
- Хотите предоставить сервис другим пользователям
- Нужна веб-панель управления
- Планируете монетизацию
- Требуется multi-tenancy

---

## Production рекомендации

**Single-User:**
- Можно развернуть на Railway, Heroku (free tier)
- Не требует БД
- Низкие требования к ресурсам

**SaaS:**
- Обязательно используйте HTTPS
- Настройте CORS для вашего frontend домена
- Добавьте rate limiting
- Мониторинг и логирование
- Backup базы данных

---

## Примеры использования

### Single-User
```bash
# .env
TELEGRAM_BOT_TOKEN=123...
TELEGRAM_CHANNEL_ID=@mychannel
GROQ_API_KEY=gsk_...
PORT=8080

# Запуск
go run cmd/bot/main.go
```

### SaaS
```bash
# .env
DATABASE_URL=postgres://...
JWT_SECRET=abc123...
BASE_URL=https://commitcaster.com
PORT=8080

# Запуск
go run cmd/bot/main_saas.go
```

---

См. также:
- [README.md](./README.md) - Single-User documentation
- [README_SAAS.md](./README_SAAS.md) - SaaS documentation
- [API.md](./API.md) - API reference
