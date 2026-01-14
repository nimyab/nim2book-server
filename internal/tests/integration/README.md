# Integration Tests

Интеграционные тесты для nim2book-server с использованием реальных зависимостей (PostgreSQL, Redis, MinIO).

## Описание

Эти тесты проверяют работу приложения с реальными сервисами:
- **PostgreSQL** - тестовая база данных
- **Redis** - кеширование
- **MinIO** - S3-совместимое хранилище

В отличие от unit-тестов, интеграционные тесты проверяют:
- Реальные SQL-запросы и миграции
- Работу с PostgreSQL массивами и JSONB
- Транзакции и constraints базы данных
- Кеширование в Redis
- Полные flow сервисов

## Требования

- Docker и Docker Compose
- Go 1.21+
- Make (опционально, для удобства)

## Быстрый старт

### 1. Запустить тестовое окружение

```bash
make test-up
```

Эта команда поднимет Docker-контейнеры:
- PostgreSQL на порту `5433`
- Redis на порту `6380`
- MinIO на порту `9001`

### 2. Запустить тесты

```bash
# Запустить тесты и автоматически остановить контейнеры
make test-integration

# Или запустить тесты, оставив контейнеры работающими (для отладки)
make test-integration-keep

# Или напрямую через go test
go test -v ./internal/tests/integration/... -timeout 5m
```

### 3. Остановить тестовое окружение

```bash
# Остановить контейнеры
make test-down

# Остановить и удалить volumes (полная очистка)
make test-clean
```

## Структура тестов

```
integration/
├── setup_test.go          # Настройка окружения, helper-функции
├── repositories_test.go   # Тесты репозиториев с реальной БД
├── services_test.go       # Тесты сервисов
└── README.md             # Этот файл
```

## Что покрывают тесты

### Repositories (repositories_test.go)

**UserRepository:**
- ✅ Создание пользователей (email/password и Google)
- ✅ Поиск по email и Google sub
- ✅ Обновление metadata (JSONB)
- ✅ Проверка unique constraints
- ✅ Проверка database constraints (one account type)

**BookRepository:**
- ✅ CRUD операции
- ✅ Работа с PostgreSQL массивами (chapter_paths)
- ✅ Поиск по автору и названию
- ✅ Проверка уникальных индексов (title + author)
- ✅ Обработка спецсимволов в массивах

**DictionaryRepository:**
- ✅ Создание и получение словарных статей
- ✅ Работа с JSONB content
- ✅ Поиск по языку и тексту
- ✅ Bulk операции

**FcmTokenRepository:**
- ✅ Добавление и удаление токенов
- ✅ Получение токенов по user_id
- ✅ Cascade delete при удалении пользователя
- ✅ Проверка уникальности токенов

**Concurrency Tests:**
- ✅ Параллельное создание книг с уникальными constraints

### Services (services_test.go)

**Auth Flow:**
- ✅ Register → Login → Access/Refresh tokens
- ✅ Обработка дублирующихся email
- ✅ Проверка неверных паролей

**User Services:**
- ✅ Me - получение профиля
- ✅ Metadata - обновление настроек
- ✅ Персистентность данных

**Book Services:**
- ✅ Get Book by ID
- ✅ Get Books с фильтрацией
- ✅ Update Book

**FCM Token Services:**
- ✅ Add/Delete FCM tokens
- ✅ Обработка дубликатов
- ✅ Проверка каскадного удаления

**Dictionary Service:**
- ✅ Lookup с кешированием
- ✅ Работа с БД и cache

**Complete User Journey:**
- ✅ Полный flow: register → login → profile → settings → browse books

**Concurrency:**
- ✅ Параллельная регистрация пользователей
- ✅ Параллельное добавление токенов

## Helper-функции

В `setup_test.go` доступны helper-функции:

```go
// Очистка всех данных из БД
cleanupDatabase(t)

// Создание тестовых данных
user := createTestUser(t, "email@example.com", "password")
user := createTestUserGoogle(t, "google-sub", "email@example.com", "Name")
book := createTestBook(t, "Title", "Author", []string{"ch1.json"})
dict := createTestDictionary(t, "word", "en-ru", content)
token := createTestFcmToken(t, "token", user)
```

## Конфигурация

Тестовая конфигурация находится в `.env.test`:
- База данных на порту 5433
- Redis на порту 6380
- MinIO на порту 9001
- Тестовые credentials

## Отладка

### Проверить статус контейнеров

```bash
docker-compose -f docker-compose.test.yml ps
```

### Посмотреть логи

```bash
# Все сервисы
docker-compose -f docker-compose.test.yml logs

# Конкретный сервис
docker-compose -f docker-compose.test.yml logs postgres-test
docker-compose -f docker-compose.test.yml logs redis-test
docker-compose -f docker-compose.test.yml logs minio-test
```

### Подключиться к PostgreSQL

```bash
docker exec -it nim2book-postgres-test psql -U testuser -d nim2book_test
```

### Подключиться к Redis

```bash
docker exec -it nim2book-redis-test redis-cli
```

### Посмотреть MinIO Web UI

Откройте браузер: http://localhost:9002
- Username: `testminio`
- Password: `testminio123`

## Запуск отдельных тестов

```bash
# Только тесты репозиториев
go test -v ./internal/tests/integration/ -run TestUserRepository

# Только тесты сервисов
go test -v ./internal/tests/integration/ -run TestAuthService

# Конкретный тест
go test -v ./internal/tests/integration/ -run TestUserRepository_CreateAndGet
```

## CI/CD Integration

Для CI/CD можно использовать:

```yaml
# Пример для GitHub Actions
- name: Start test services
  run: make test-up

- name: Run integration tests
  run: go test -v ./internal/tests/integration/... -timeout 5m

- name: Stop test services
  run: make test-down
  if: always()
```

## Troubleshooting

### Порты уже заняты

Если порты 5433, 6380 или 9001 уже используются, измените их в `docker-compose.test.yml`.

### Тесты падают с "connection refused"

Подождите дольше после `make test-up`. Сервисам нужно время для инициализации:

```bash
make test-up
sleep 15
go test -v ./internal/tests/integration/...
```

### Ошибки миграций

Проверьте, что все миграции корректны и применяются:

```bash
docker exec -it nim2book-postgres-test psql -U testuser -d nim2book_test -c "\dt"
```

### Cleanup после сбоя

Если тесты упали и оставили мусор в БД:

```bash
make test-clean
make test-up
```

## Best Practices

1. **Всегда используйте `cleanupDatabase(t)` в начале теста** для изоляции
2. **Используйте helper-функции** для создания тестовых данных
3. **Проверяйте не только success path**, но и error cases
4. **Тестируйте constraints и edge cases** (пустые массивы, nil значения)
5. **Используйте `require` для критических проверок**, `assert` для остальных
6. **Именуйте тесты понятно**: `TestService_Operation_Scenario`

## Что НЕ тестируется

❌ Внешние API (Yandex Dictionary, LibreTranslate) - используются моки
❌ Firebase Messaging - мок
❌ Google OAuth - мок
❌ gRPC Word Aligner - мок

Эти зависимости тестируются отдельно через unit-тесты с моками.

## Производительность

Полный прогон всех интеграционных тестов занимает ~30-60 секунд:
- Setup: ~10 секунд
- Repository tests: ~15-20 секунд
- Service tests: ~10-15 секунд
- Teardown: ~5 секунд

## Дальнейшее развитие

Планы по расширению тестов:
- [ ] E2E HTTP API тесты
- [ ] WebSocket тесты для notifications
- [ ] Тесты для gRPC клиентов
- [ ] Performance/Load тесты
- [ ] Тесты для миграций (up/down)