# Nim2Book Server

## 🚀 Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/nimyab/nim2book-server.git
cd nim2book-server
```

### 2. Настройка окружения

Создайте файл `.env` в корне проекта и скопируйте данные из `.env.example`

### 3. Установка зависимостей

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/swaggo/swag/cmd/swag@latest
go mod download
```

### 4. Запуск через Docker

```bash
make docker-up
```

### 5. Применение миграций

```bash
make migrate-up
```

### 6. Запуск приложения

```bash
# Режим разработки (с автогенерацией Swagger)
make dev

# Или сборка бинарника
make build
./bin/app
```

## 📚 API Документация

После запуска приложения, Swagger UI доступен по адресу:

```
http://localhost:5050/swagger/
```
