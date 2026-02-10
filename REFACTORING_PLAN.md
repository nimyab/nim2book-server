# План рефакторинга Nim2Book

## Приоритет 1: Критические проблемы (1-2 недели)

### 1.1 Убрать глобальные переменные сервисов
**Задача:** Передавать сервисы через зависимости, а не глобальные переменные.

**До:**
```go
var service *Service
func New(...) *Service {
    service = &Service{...}
    return service
}
func HTTPv1(c echo.Context) error {
    output, err := service.Login(input)  // глобальная переменная!
}
```

**После:**
```go
type Service struct {
    pg Postgres
}

func New(pg Postgres) *Service {
    return &Service{pg: pg}  // без глобального состояния
}

// Handler принимает сервис как зависимость
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
    return func(c echo.Context) error {
        output, err := svc.Login(input)
        // ...
    }
}
```

### 1.2 Создать структуру приложения (App)
```go
// internal/app/app.go
type App struct {
    Config *config.Config
    
    // Adapters
    DB    *postgres_sqlc.Postgres
    S3    *minio.Client
    Cache *redis_cache.Client
    
    // Services
    AuthService *auth.Service
    BookService *book.Service
    // ...
    
    // HTTP Server
    Server *echo.Echo
}

func New(cfg *config.Config) (*App, error) {
    app := &App{Config: cfg}
    
    // Инициализация адаптеров
    if err := app.initAdapters(); err != nil {
        return nil, err
    }
    
    // Инициализация сервисов
    app.initServices()
    
    // Настройка роутов
    app.setupRoutes()
    
    return app, nil
}
```

### 1.3 Исправить конфигурацию
**До:**
```go
var appConfig *Config
func init() { appConfig = &Config{...} }
func GetConfig() *Config { return appConfig }
```

**После:**
```go
func Load() (*Config, error) {
    cfg := &Config{}
    
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    
    return cfg, nil
}

// В main.go:
cfg, err := config.Load()
```

## Приоритет 2: Clean Architecture (2-3 недели)

### 2.1 Создать слой Use Cases
```
internal/
  usecase/
    auth/
      login.go
      register.go
    book/
      get_books.go
      translate_book.go
```

**Пример:**
```go
// internal/usecase/auth/login.go
type LoginUseCase struct {
    userRepo   domain.UserRepository
    tokenGen   domain.TokenGenerator
    hasher     domain.PasswordHasher
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*domain.AuthResult, error) {
    user, err := uc.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    
    if !uc.hasher.Compare(user.PasswordHash, password) {
        return nil, ErrInvalidCredentials
    }
    
    tokens, err := uc.tokenGen.Generate(user)
    if err != nil {
        return nil, err
    }
    
    return &domain.AuthResult{User: user, Tokens: tokens}, nil
}
```

### 2.2 Переместить интерфейсы в domain
```go
// internal/domain/repository.go
type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
}

type BookRepository interface {
    GetByID(ctx context.Context, id Id) (*Book, error)
    List(ctx context.Context, filter BookFilter) ([]Book, error)
    Create(ctx context.Context, book *Book) error
}
```

### 2.3 Реализовать интерфейсы в адаптерах
```go
// internal/adapter/postgres_sqlc/user_repository.go
type UserRepository struct {
    queries *sqlc.Queries
}

func NewUserRepository(db *Postgres) *UserRepository {
    return &UserRepository{queries: db.Queries}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    // реализация
}
```

## Приоритет 3: Улучшения (1-2 недели)

### 3.1 Управление транзакциями
```go
// pkg/transaction/manager.go
type Manager interface {
    InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// internal/adapter/postgres_sqlc/tx_manager.go
func (tm *TxManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := tm.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    ctx = context.WithValue(ctx, txKey, tx)
    
    if err := fn(ctx); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

### 3.2 Dependency Injection с Wire/Fx
Использовать google/wire для автоматической генерации DI кода.

### 3.3 Добавить тесты
После рефакторинга можно легко писать unit-тесты:
```go
func TestLoginUseCase_Execute(t *testing.T) {
    mockRepo := &MockUserRepository{}
    mockTokenGen := &MockTokenGenerator{}
    mockHasher := &MockPasswordHasher{}
    
    uc := NewLoginUseCase(mockRepo, mockTokenGen, mockHasher)
    
    // тестируем изолированно
}
```

## Целевая архитектура

```
cmd/app/main.go                    # Точка входа
  └── internal/app/app.go          # Композиция приложения

internal/
  ├── domain/                      # Бизнес-логика
  │   ├── entities (User, Book)
  │   ├── value objects (Id, Email)
  │   ├── interfaces (UserRepository)
  │   └── business errors
  │
  ├── usecase/                     # Сценарии использования
  │   ├── auth/
  │   ├── book/
  │   └── user/
  │
  ├── adapter/                     # Реализации интерфейсов
  │   ├── repository/
  │   │   └── postgres/           # Реализация репозиториев
  │   ├── storage/
  │   │   └── minio/
  │   ├── cache/
  │   │   └── redis/
  │   └── external/
  │       └── firebase/
  │
  └── controller/                  # Входные адаптеры
      ├── http/                    # REST API
      ├── grpc/                    # gRPC API
      └── websocket/

config/                            # Конфигурация
pkg/                              # Переиспользуемые пакеты
```

## Порядок выполнения

1. ✅ Создать структуру App
2. ✅ Убрать глобальные переменные из сервисов
3. ✅ Исправить конфигурацию
4. ✅ Переписать роутинг для использования App
5. ⬜ Создать интерфейсы в domain
6. ⬜ Создать usecase слой
7. ⬜ Переместить реализации в adapters
8. ⬜ Добавить управление транзакциями
9. ⬜ Внедрить DI container
10. ⬜ Написать тесты

## Результаты после рефакторинга

✅ Легко тестируется  
✅ Легко добавлять новые фичи  
✅ Независимые слои  
✅ Можно менять БД без изменения бизнес-логики  
✅ Можно добавить gRPC handler рядом с HTTP  
✅ Чистая архитектура (Clean Architecture)  
