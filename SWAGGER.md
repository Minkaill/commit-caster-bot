# Swagger Documentation

CommitCaster использует Swagger для автоматической генерации API документации.

## Генерация документации

### Windows

```bash
.\generate-swagger.bat
```

### Linux/Mac

```bash
chmod +x generate-swagger.sh
./generate-swagger.sh
```

### Вручную

```bash
# Установка swag (один раз)
go install github.com/swaggo/swag/cmd/swag@latest

# Генерация
swag init -g cmd/bot/main.go -o docs
```

## Просмотр документации

1. Сгенерируйте документацию (см. выше)
2. Запустите сервер:
   ```bash
   go run cmd/bot/main.go
   ```
3. Откройте в браузере:
   ```
   http://localhost:8080/swagger/index.html
   ```

## Swagger UI

В Swagger UI вы можете:
- Просмотреть все endpoints
- Посмотреть модели данных
- Протестировать API прямо в браузере
- Скачать OpenAPI спецификацию

## Аутентификация в Swagger

Для защищённых endpoints:

1. Нажмите кнопку **"Authorize"** в правом верхнем углу
2. Введите: `Bearer <ваш_jwt_token>`
3. Нажмите "Authorize"
4. Теперь можете тестировать защищённые endpoints

## Обновление документации

После изменения API:

1. Добавьте/обновите Swagger аннотации в коде:
   ```go
   // @Summary Краткое описание
   // @Description Подробное описание
   // @Tags tag_name
   // @Accept json
   // @Produce json
   // @Param request body YourType true "Description"
   // @Success 200 {object} ResponseType
   // @Failure 400 {object} map[string]string
   // @Router /path [method]
   func YourHandler(c *gin.Context) {
       // ...
   }
   ```

2. Перегенерируйте документацию:
   ```bash
   swag init -g cmd/bot/main.go -o docs
   ```

3. Перезапустите сервер

## Swagger аннотации

### Общая информация (в main.go)

```go
// @title CommitCaster API
// @version 1.0
// @description API описание

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
```

### Для endpoints (в handlers)

```go
// @Summary Краткое описание
// @Description Подробное описание
// @Tags category
// @Accept json
// @Produce json
// @Param id path int true "ID параметр"
// @Param request body RequestType true "Тело запроса"
// @Success 200 {object} ResponseType
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /endpoint/{id} [get]
```

### Типы параметров

- `@Param name path type required "description"` - path параметр
- `@Param name query type required "description"` - query параметр
- `@Param name header type required "description"` - header
- `@Param request body Type required "description"` - body

### Типы ответов

- `{object} Type` - объект
- `{array} Type` - массив
- `{string} string` - строка
- `map[string]string` - map

## Пример endpoint с полными аннотациями

```go
// CreateUser создаёт нового пользователя
// @Summary Создание пользователя
// @Description Создаёт нового пользователя с валидацией данных
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Данные пользователя"
// @Success 201 {object} User
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 409 {object} map[string]string "Пользователь существует"
// @Failure 500 {object} map[string]string "Серверная ошибка"
// @Security BearerAuth
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
    // ...
}
```

## Файлы документации

После генерации создаются файлы:
- `docs/docs.go` - Go код документации
- `docs/swagger.json` - OpenAPI JSON спецификация
- `docs/swagger.yaml` - OpenAPI YAML спецификация

**Важно:** Не редактируйте эти файлы вручную! Они генерируются автоматически.

## Экспорт спецификации

OpenAPI спецификация доступна по адресам:
- JSON: `http://localhost:8080/swagger/doc.json`
- YAML: `http://localhost:8080/swagger/doc.yaml`

Можете импортировать в:
- Postman
- Insomnia
- API клиенты
- Генераторы кода

## Troubleshooting

**Ошибка: "swag: command not found"**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
# Убедитесь что $GOPATH/bin в PATH
```

**Ошибка: "cannot find package commitcaster/docs"**
```bash
# Сначала сгенерируйте документацию
swag init -g cmd/bot/main.go -o docs
```

**Swagger UI не открывается**
- Проверьте что сервер запущен в SaaS режиме (с DATABASE_URL)
- Проверьте порт: http://localhost:8080/swagger/index.html
- Убедитесь что документация сгенерирована

**Изменения не отображаются**
- Перегенерируйте документацию
- Перезапустите сервер
- Очистите кеш браузера (Ctrl+F5)

## CI/CD Integration

Для автоматической генерации в CI/CD:

```yaml
# .github/workflows/build.yml
- name: Generate Swagger docs
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g cmd/bot/main.go -o docs
```

## Дополнительная информация

- [Документация Swaggo](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Примеры аннотаций](https://github.com/swaggo/swag#declarative-comments-format)
