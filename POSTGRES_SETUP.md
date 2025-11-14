# PostgreSQL Setup для CommitCaster

## Вариант 1: Docker (Самый простой) ⭐

### Требования:
- Docker Desktop для Windows

### Запуск:

```bash
# 1. Запусти PostgreSQL через docker-compose
docker-compose up -d

# 2. Проверь что запустился
docker ps

# 3. Создай .env файл
DATABASE_URL=postgres://postgres:postgres@localhost:5432/commitcaster
JWT_SECRET=$(openssl rand -hex 32)
BASE_URL=http://localhost:8080
PORT=8080

# 4. Запусти бота
go run cmd/bot/main.go
```

### Остановка:
```bash
docker-compose down
```

### Команды:

```bash
# Посмотреть логи
docker-compose logs -f

# Подключиться к БД
docker exec -it commitcaster-db psql -U postgres -d commitcaster

# Удалить всё (включая данные)
docker-compose down -v
```

---

## Вариант 2: Локальная установка

### Windows

**Установка:**
1. Скачай с https://www.postgresql.org/download/windows/
2. Запусти установщик PostgreSQL 15+
3. Запомни пароль для пользователя `postgres`
4. Порт: оставь 5432

**Настройка:**

```cmd
# Открой PowerShell или CMD
# Подключись к PostgreSQL
psql -U postgres

# Создай базу данных
CREATE DATABASE commitcaster;

# Выйди
\q
```

**Environment Variables:**
```env
DATABASE_URL=postgres://postgres:твой_пароль@localhost:5432/commitcaster
JWT_SECRET=your_random_32_char_secret
BASE_URL=http://localhost:8080
```

### Генерация JWT_SECRET:

**Windows (PowerShell):**
```powershell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | % {[char]$_})
```

**Git Bash:**
```bash
openssl rand -hex 32
```

---

## Вариант 3: Облачный PostgreSQL (Бесплатно)

### Supabase ⭐ РЕКОМЕНДУЮ

**Почему:**
- Бесплатный тариф
- 500MB хранилище
- Автоматические бэкапы
- API из коробки

**Шаги:**

1. Зайди на https://supabase.com
2. Sign up (можно через GitHub)
3. **New project**:
   - Name: `commitcaster`
   - Database Password: придумай и запомни
   - Region: выбери ближайший
4. Подожди пока создастся (~2 минуты)
5. Перейди в **Settings** → **Database**
6. Скопируй **Connection string** (Pooler mode)
7. Замени `[YOUR-PASSWORD]` на твой пароль

**Пример:**
```env
DATABASE_URL=postgresql://postgres.xxxxx:password@aws-0-eu-central-1.pooler.supabase.com:6543/postgres
JWT_SECRET=your_random_secret
BASE_URL=https://your-domain.com
```

### Neon

1. https://neon.tech
2. Sign up
3. Create project
4. Copy connection string
5. Вставь в .env

### Railway

1. https://railway.app
2. New Project → Deploy PostgreSQL
3. Copy connection URL
4. Вставь в .env

---

## Проверка подключения

После настройки проверь подключение:

```bash
# Запусти бота
go run cmd/bot/main.go

# Должно появиться:
# 🌐 Starting in SaaS mode (multi-user)
# ✅ Connected to PostgreSQL
# Running database migrations...
# ✅ Database migrations completed
```

### Если ошибка подключения:

**"failed to connect to database"**
- Проверь DATABASE_URL
- Проверь что PostgreSQL запущен
- Проверь пароль

**"JWT_SECRET not set"**
- Добавь JWT_SECRET в .env

**"database does not exist"**
```bash
# Создай БД вручную
psql -U postgres
CREATE DATABASE commitcaster;
\q
```

---

## Миграции

Миграции выполняются автоматически при запуске.

Создаются таблицы:
- `users` - пользователи
- `user_settings` - настройки пользователей

---

## Управление БД

### Подключение к БД:

**Локальный PostgreSQL:**
```bash
psql -U postgres -d commitcaster
```

**Docker:**
```bash
docker exec -it commitcaster-db psql -U postgres -d commitcaster
```

### Полезные команды:

```sql
-- Посмотреть таблицы
\dt

-- Посмотреть пользователей
SELECT * FROM users;

-- Посмотреть настройки
SELECT * FROM user_settings;

-- Удалить всех пользователей
TRUNCATE users, user_settings CASCADE;

-- Выйти
\q
```

### GUI клиенты:

- **pgAdmin** - https://www.pgadmin.org/
- **DBeaver** - https://dbeaver.io/
- **TablePlus** - https://tableplus.com/

---

## Production Setup

Для продакшена используй:

1. **Heroku Postgres** - автоматически при деплое
2. **Railway** - встроенный PostgreSQL
3. **Supabase** - бесплатный план
4. **AWS RDS** - если нужна масштабируемость

---

## Troubleshooting

**Port 5432 занят:**
```bash
# Найди процесс
netstat -ano | findstr :5432

# Останови PostgreSQL сервис
# Windows Services → PostgreSQL → Stop
```

**Пароль не подходит:**
```bash
# Сброс пароля (локальный PostgreSQL)
# Отредактируй pg_hba.conf
# Смени 'md5' на 'trust'
# Перезапусти PostgreSQL
# Подключись и смени пароль:
ALTER USER postgres PASSWORD 'новый_пароль';
```

**Docker не запускается:**
```bash
# Проверь Docker Desktop запущен
# Проверь нет ли других контейнеров на порту 5432
docker ps -a
docker rm -f commitcaster-db
docker-compose up -d
```

---

## Рекомендация

Для **локальной разработки**: используй **Docker** (docker-compose.yml)

Для **продакшена**: используй **Supabase** или **Railway**

Это проще всего и работает из коробки! 🚀
