# Анализ проекта Nim2Book Server - Найденные проблемы и рекомендации

**Дата анализа:** 16 февраля 2026  
**Аналитик:** Senior Go Developer (6 лет опыта)

---

## 🔴 Критические проблемы (требуют немедленного решения)

### 1. Безопасность: Слабая конфигурация CORS
**Проблема:**  
В файле `internal/app/http.go:46` используется `middleware.CORS()` без настроек, что открывает API для любых доменов.

**Риски:**
- XSS атаки
- CSRF атаки
- Несанкционированный доступ к API

**Решение:**
```go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{
        "https://yourdomain.com",
        "https://app.yourdomain.com",
    },
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
    AllowCredentials: true,
    MaxAge: 300,
}))
```

### 2. Некорректный путь в Dockerfile
**Проблема:**  
В `Dockerfile:9` указан неверный путь: `go build -o main ./cmd/main.go`, но файл находится в `./cmd/app/main.go`

**Решение:**
```dockerfile
RUN go build -o main ./cmd/app/main.go
```

### 3. Отсутствие Graceful Shutdown для адаптеров
**Проблема:**  
При остановке приложения не закрываются соединения с PostgreSQL, Redis, MinIO и gRPC клиентом.

**Риски:**
- Утечка соединений
- Потеря данных
- Проблемы при рестарте

**Решение:**
1. Добавить интерфейс Closable для всех адаптеров:
```go
// internal/adapter/adapter.go
package adapter

type Closable interface {
    Close() error
}
```

2. Реализовать Close() для каждого адаптера:
```go
// internal/adapter/postgres/postgres.go
func (p *Postgres) Close() error {
    return p.Client.Close()
}

// internal/adapter/redis_cache/redis_cache.go
func (r *RedisCache) Close() error {
    return r.client.Close()
}

// internal/adapter/minio/minio.go (если необходимо)
func (m *Minio) Close() error {
    // MinIO клиент обычно не требует явного закрытия
    return nil
}
```

3. Обновить `internal/app/app.go`:
```go
func (a *App) Shutdown(ctx context.Context) error {
    slog.Info("Shutting down application...")

    // Shutdown HTTP server first
    if err := a.server.Shutdown(ctx); err != nil {
        return err
    }

    // Close database connections
    if client, err := do.Invoke[*ent.Client](a.injector); err == nil {
        if err := client.Close(); err != nil {
            slog.Error("Failed to close database connection", slog.Any("error", err))
        }
    }

    // Close Redis
    if redis, err := do.Invoke[*redis_cache.RedisCache](a.injector); err == nil {
        if err := redis.Close(); err != nil {
            slog.Error("Failed to close redis connection", slog.Any("error", err))
        }
    }

    // Shutdown DI container
    if err := a.injector.Shutdown(); err != nil {
        return err
    }

    slog.Info("Application shut down successfully")
    return nil
}
```

### 4. Использование context.Background() вместо request context
**Проблема:**  
Во многих сервисах используется `context.Background()` вместо контекста из HTTP запроса.

**Риски:**
- Невозможность отменить длительные операции
- Утечка ресурсов
- Отсутствие трассировки запросов

**Примеры:**
- `internal/services/auth/login/service.go:46`
- `internal/services/book/get_books/service.go:32`

**Решение:**
1. Добавить контекст в методы сервисов:
```go
func (s *Service) GetBooks(ctx context.Context, input *Input) (*Output, error) {
    // ...
    books, err := s.bookRepo.SearchWithFilters(
        ctx, // <- передавать контекст из запроса
        input.Title,
        input.Author,
        input.GenreId,
        repository.QueryOptions{
            Limit:  booksPerPage,
            Offset: offset,
        },
    )
    // ...
}
```

2. Обновить хендлеры для передачи контекста:
```go
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
    return func(c echo.Context) error {
        // ...
        output, err := svc.GetBooks(c.Request().Context(), input)
        // ...
    }
}
```

### 5. Отсутствие валидации конфигурации
**Проблема:**  
В `config/config.go` отсутствует валидация обязательных параметров.

**Риски:**
- Runtime ошибки при обращении к несконфигурированным сервисам
- Сложность диагностики проблем

**Решение:**
```go
func Load() (*Config, error) {
    maxRequestCount, _ := strconv.Atoi(os.Getenv("MAX_REQUEST_COUNT"))
    jwtAccessTime, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TIME"))
    jwtRefreshTime, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TIME"))
    waitMilliseconds, _ := strconv.Atoi(os.Getenv("WAIT_MILLISECONDS"))

    cfg := &Config{
        // ... existing code ...
    }

    // Добавить валидацию
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return cfg, nil
}

func (c *Config) Validate() error {
    var errors []string

    if c.PostgresURL == "" {
        errors = append(errors, "POSTGRES_URL is required")
    }
    if c.RedisURL == "" {
        errors = append(errors, "REDIS_URL is required")
    }
    if c.JWTSecret == "" {
        errors = append(errors, "JWT_SECRET is required")
    }
    if c.MinioURL == "" {
        errors = append(errors, "MINIO_URL is required")
    }

    if len(errors) > 0 {
        return fmt.Errorf("configuration errors: %s", strings.Join(errors, "; "))
    }

    return nil
}
```

---

## 🟡 Высокий приоритет

### 6. Отсутствие Rate Limiting
**Проблема:**  
Отсутствует ограничение количества запросов на критические эндпоинты (login, register, translate).

**Риски:**
- Brute force атаки
- DDoS атаки
- Перегрузка сервера

**Решение:**
1. Установить middleware для rate limiting:
```bash
go get github.com/ulule/limiter/v3
go get github.com/ulule/limiter/v3/drivers/store/redis
```

2. Создать middleware `internal/middleware/rate_limit.go`:
```go
package middleware

import (
    "github.com/labstack/echo/v4"
    "github.com/ulule/limiter/v3"
    mw "github.com/ulule/limiter/v3/drivers/middleware/echo"
    "github.com/ulule/limiter/v3/drivers/store/redis"
)

func RateLimit(redisClient *redis.Client) echo.MiddlewareFunc {
    rate := limiter.Rate{
        Period: 1 * time.Minute,
        Limit:  10,
    }
    
    store, err := redis.NewStore(redisClient)
    if err != nil {
        panic(err)
    }
    
    instance := limiter.New(store, rate)
    return mw.NewMiddleware(instance)
}
```

3. Применить к критическим маршрутам:
```go
apiV1.POST("/auth/login", login.MakeHTTPv1Handler(svc, a.config), rateLimitMiddleware)
apiV1.POST("/auth/register", register.MakeHTTPv1Handler(svc), rateLimitMiddleware)
```

### 7. Отсутствие Health Checks
**Проблема:**  
Нет проверки работоспособности зависимостей (PostgreSQL, Redis, MinIO, gRPC).

**Решение:**
Создать `internal/services/health/service.go`:
```go
package health

import (
    "context"
    "time"

    "github.com/nimyab/nim2book-back/ent"
    "github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
)

type Service struct {
    db    *ent.Client
    redis *redis_cache.RedisCache
}

type HealthStatus struct {
    Status   string            `json:"status"`
    Services map[string]string `json:"services"`
}

func New(db *ent.Client, redis *redis_cache.RedisCache) *Service {
    return &Service{db: db, redis: redis}
}

func (s *Service) Check() *HealthStatus {
    status := &HealthStatus{
        Status:   "ok",
        Services: make(map[string]string),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Check PostgreSQL
    if err := s.db.DB().PingContext(ctx); err != nil {
        status.Services["postgres"] = "unhealthy"
        status.Status = "degraded"
    } else {
        status.Services["postgres"] = "healthy"
    }

    // Check Redis
    if err := s.redis.Get(ctx, "health-check"); err != nil {
        status.Services["redis"] = "unhealthy"
        status.Status = "degraded"
    } else {
        status.Services["redis"] = "healthy"
    }

    return status
}
```

Обновить эндпоинт:
```go
apiV1.GET("/health", func(c echo.Context) error {
    healthSvc := do.MustInvoke[*health.Service](a.injector)
    return c.JSON(200, healthSvc.Check())
})
```

### 8. Недостаточное тестовое покрытие
**Проблема:**  
Только 2 теста в проекте (`jwt_test.go` и `contains_letters_test.go`).

**Решение:**
1. Создать тесты для репозиториев (используя `ent/enttest`):
```go
// internal/repository/user_repo_test.go
package repository_test

import (
    "context"
    "testing"

    "github.com/nimyab/nim2book-back/ent/enttest"
    "github.com/nimyab/nim2book-back/internal/repository"
    "github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    defer client.Close()

    repo := repository.NewUserRepository(client)

    // TODO: Add test cases
}
```

2. Создать тесты для сервисов (используя моки):
```go
// internal/services/auth/login/service_test.go
package login_test

import (
    "context"
    "testing"

    "github.com/nimyab/nim2book-back/internal/services/auth/login"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByBasicAccountEmail(ctx context.Context, email string) (*domain.User, error) {
    args := m.Called(ctx, email)
    return args.Get(0).(*domain.User), args.Error(1)
}

func TestService_Login(t *testing.T) {
    // TODO: Add test cases
}
```

3. Добавить target в Makefile:
```makefile
test-coverage-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html
```

### 9. Отсутствие структурированного логирования в критических местах
**Проблема:**  
Непоследовательное использование structured logging, некоторые места используют простые string messages.

**Решение:**
Создать helper для консистентного логирования:
```go
// pkg/logger/context.go
package logger

import (
    "log/slog"
)

func Error(msg string, err error, operation string, args ...any) {
    attrs := []any{
        slog.String("operation", operation),
        slog.Any("error", err),
    }
    attrs = append(attrs, args...)
    slog.Error(msg, attrs...)
}

func Info(msg string, operation string, args ...any) {
    attrs := []any{
        slog.String("operation", operation),
    }
    attrs = append(attrs, args...)
    slog.Info(msg, attrs...)
}

func Warn(msg string, operation string, args ...any) {
    attrs := []any{
        slog.String("operation", operation),
    }
    attrs = append(attrs, args...)
    slog.Warn(msg, attrs...)
}
```

### 10. Отсутствие обработки медленных запросов (Timeout)
**Проблема:**  
HTTP клиенты для LibreTranslate и YandexDictionary не настроены с timeout.

**Решение:**
```go
// internal/services/libretranslate/translate/service.go
func New(baseURL string) *Service {
    return &Service{
        baseURL: baseURL,
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}
```

---

## 🟢 Средний приоритет

### 11. Отсутствие миграций в репозитории
**Проблема:**  
Миграции генерируются, но не хранятся в Git.

**Решение:**
1. Создать директорию `ent/migrate/migrations/`
2. Добавить в `.gitignore` исключение:
```
!ent/migrate/migrations/**
```
3. Commit существующих миграций

### 12. Отсутствие CI/CD Pipeline
**Проблема:**  
Нет автоматизации тестирования и деплоя.

**Решение:**
Создать `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_USER: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
        env:
          POSTGRES_URL: postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.txt

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build
        run: go build -v ./cmd/app/main.go
```

### 13. Отсутствие линтера
**Проблема:**  
Нет конфигурации для golangci-lint.

**Решение:**
Создать `.golangci.yml`:
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - goconst
    - gocyclo
    - dupl
    - gosec
    - unconvert
    - exportloopref

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
  misspell:
    locale: US

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

Добавить в Makefile:
```makefile
lint:
	golangci-lint run --config .golangci.yml

lint-fix:
	golangci-lint run --fix --config .golangci.yml
```

### 14. Недостаточная документация API
**Проблема:**  
Swagger комментарии присутствуют не везде.

**Решение:**
Добавить swagger annotations для всех эндпоинтов:
```go
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      Input   true  "Login credentials"
// @Success      200    {object}  Output
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Router       /auth/login [post]
func MakeHTTPv1Handler(svc *Service, cfg *config.Config) echo.HandlerFunc {
    // ...
}
```

### 15. Отсутствие Prometheus метрик
**Проблема:**  
Нет мониторинга производительности и здоровья приложения.

**Решение:**
1. Установить зависимость:
```bash
go get github.com/labstack/echo-contrib/echoprometheus
```

2. Добавить middleware:
```go
// internal/app/http.go
import "github.com/labstack/echo-contrib/echoprometheus"

func (a *App) setupHTTPServer() {
    e := echo.New()
    
    // ... existing middleware ...
    
    e.Use(echoprometheus.NewMiddleware("nim2book"))
    e.GET("/metrics", echoprometheus.NewHandler())
    
    // ... rest of setup ...
}
```

### 16. Глобальная переменная socketHub
**Проблема:**  
В `internal/controller/websocket/socker.hub.go` используется глобальная переменная.

**Решение:**
```go
// Вместо глобальной переменной, создать фабрику
type HubFactory struct {
    hub *SocketHub
}

func NewHubFactory() *HubFactory {
    hub := &SocketHub{
        connections: make(map[domain.ID]*SocketConn),
        registerCh:  make(chan *SocketConn),
        messageCh:   make(chan *Message),
    }
    go hub.run()
    return &HubFactory{hub: hub}
}

// Зарегистрировать в DI контейнере
do.Provide(a.injector, func(i do.Injector) (*HubFactory, error) {
    return NewHubFactory(), nil
})

// Передавать через параметры
func MakeSocketConnHandler(cfg *config.Config, hubFactory *HubFactory) echo.HandlerFunc {
    return func(c echo.Context) error {
        // использовать hubFactory.hub
    }
}
```

### 17. Небезопасное хранение паролей в логах
**Проблема:**  
Есть риск логирования sensitive data.

**Решение:**
1. Создать middleware для sanitization:
```go
// internal/middleware/sensitive.go
package middleware

import (
    "bytes"
    "io"
    "strings"

    "github.com/labstack/echo/v4"
)

var sensitiveFields = []string{"password", "token", "secret", "authorization"}

func SanitizeRequestBody() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            if c.Request().Body != nil {
                bodyBytes, _ := io.ReadAll(c.Request().Body)
                c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
                
                body := string(bodyBytes)
                for _, field := range sensitiveFields {
                    if strings.Contains(strings.ToLower(body), field) {
                        // Не логировать тело запроса
                        c.Set("skip_body_log", true)
                        break
                    }
                }
            }
            return next(c)
        }
    }
}
```

### 18. Отсутствие версионирования Docker образов
**Проблема:**  
В docker-compose используются `latest` теги.

**Решение:**
```yaml
# docker-compose.dev.yml
services:
  libretranslate:
    image: libretranslate/libretranslate:v1.5.7  # Фиксированная версия
  
  redis:
    image: redis:7.2-alpine  # Фиксированная версия
  
  postgres:
    image: postgres:17-alpine  # Фиксированная версия
  
  minio:
    image: minio/minio:RELEASE.2024-01-28T22-35-53Z  # Фиксированная версия
```

---

## 🔵 Низкий приоритет (улучшения)

### 19. Использование JSON для metadata
**Проблема:**  
В схеме User используется `JSON` для metadata, что может быть неоптимально.

**Рекомендация:**
Рассмотреть использование JSONB в PostgreSQL для лучшей производительности:
```go
field.JSON("metadata", map[string]any{}).
    Default(map[string]any{}).
    SchemaType(map[string]string{
        dialect.Postgres: "jsonb",
    }),
```

### 20. Отсутствие pre-commit hooks
**Решение:**
Создать `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        args: [-w, -s]
        language: system
        files: \.go$
      
      - id: go-test
        name: go test
        entry: go test
        args: [./...]
        language: system
        pass_filenames: false
```

### 21. Markdown линтер ошибки в README
**Проблема:**  
README.md содержит markdown ошибки (по результатам линтера).

**Решение:**
Исправить форматирование согласно стандартам markdown:
- Добавить пустые строки вокруг списков
- Использовать заголовки вместо жирного текста
- Удалить двоеточия из заголовков
- Указать язык для code blocks

### 22. Рефакторинг больших сервисов
**Проблема:**  
Сервисы `translate_book` и `translate_personal_user_book` очень большие (>350 строк).

**Рекомендация:**
Разбить на более мелкие компоненты:
- Extractor (извлечение текста из книги)
- Translator (перевод)
- Aligner (выравнивание слов)
- Uploader (загрузка результатов)

### 23. Добавить OpenTelemetry для трассировки
**Рекомендация:**
Интегрировать OpenTelemetry для distributed tracing:
```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho
```

### 24. Database Connection Pooling
**Рекомендация:**
Настроить параметры connection pool для PostgreSQL:
```go
func New(config *Config) (*ent.Client, error) {
    client, err := ent.Open("postgres", config.PostgresURL)
    if err != nil {
        return nil, err
    }
    
    db := client.DB()
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return client, nil
}
```

---

## 📊 Статистика проекта

- **Всего файлов Go:** ~150+
- **Покрытие тестами:** <5% (критически низкое)
- **Количество сервисов:** ~20+
- **Внешние зависимости:** PostgreSQL, Redis, MinIO, LibreTranslate, Word Aligner gRPC
- **Используемые фреймворки:** Echo, Ent, Samber/do

---

## 🎯 Приоритетный план действий

### Неделя 1-2: Критические проблемы
1. ✅ Исправить Dockerfile (5 минут)
2. ✅ Настроить CORS правильно (15 минут)
3. ✅ Добавить graceful shutdown (2 часа)
4. ✅ Заменить context.Background() на request context (4 часа)
5. ✅ Добавить валидацию конфигурации (1 час)

### Неделя 3-4: Высокий приоритет
6. ✅ Внедрить rate limiting (3 часа)
7. ✅ Добавить health checks (2 часа)
8. ✅ Настроить HTTP client timeouts (1 час)
9. ✅ Улучшить логирование (2 часа)

### Неделя 5-6: Средний приоритет
10. ✅ Настроить CI/CD (4 часа)
11. ✅ Добавить golangci-lint (2 часа)
12. ✅ Написать тесты для критических сервисов (16 часов)
13. ✅ Добавить Prometheus метрики (3 часа)

### Неделя 7-8: Улучшения
14. ✅ Документировать API полностью (6 часов)
15. ✅ Рефакторинг больших сервисов (8 часов)
16. ✅ Добавить OpenTelemetry (4 часа)

---

## 📚 Полезные ресурсы

- [Ent Documentation](https://entgo.io/docs/getting-started)
- [Echo Best Practices](https://echo.labstack.com/guide/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_SCP_Cheat_Sheet.html)
- [Effective Go](https://golang.org/doc/effective_go)

---

## 📝 Заключение

Проект имеет **хорошую архитектурную основу** с использованием современных подходов (Clean Architecture, DI, ORM). Основные проблемы связаны с:

1. **Безопасностью** - требует немедленного внимания
2. **Обработкой ошибок и контекстов** - влияет на надежность
3. **Тестированием** - критически низкое покрытие
4. **DevOps практиками** - отсутствие CI/CD и мониторинга

При последовательном решении указанных проблем проект достигнет production-ready состояния за **6-8 недель** работы одного разработчика.

**Общая оценка:** 6/10  
**Потенциал:** 9/10  

Проект демонстрирует понимание современных практик разработки на Go, но требует доработки в области безопасности, тестирования и операционной готовности.
