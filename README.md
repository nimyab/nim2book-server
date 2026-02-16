# Nim2Book Server

Backend сервер для Nim2Book - платформы для чтения книг с интегрированным переводчиком.

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/nimyab/nim2book-server.git
cd nim2book-server
```

### 2. Настройка окружения

Создайте файл `.env` в корне проекта:

### 3. Установка инструментов разработки

Установите все необходимые инструменты одной командой:

```bash
make install-tools
```

Эта команда установит:
- `swag` - генератор Swagger документации
- `goose` - инструмент миграций

Альтернативно, установите зависимости Go:

```bash
go mod download
```

### 4. Запуск базы данных

```bash
make docker-up
```

### 5. Применение миграций

**Использование Goose**

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

## 📋 Доступные команды

Для полного списка команд:

```bash
make help
```

### Основные команды:

**Разработка:**
- `make dev` - запуск dev сервера
- `make build` - сборка приложения
- `make swagger` - генерация Swagger документации

**Docker:**
- `make docker-up` - запуск контейнеров
- `make docker-down` - остановка контейнеров

**Goose миграции:**
- `make migrate-create NAME=<name>` - создать новую миграцию
- `make migrate-up` - применить все миграции
- `make migrate-down` - откатить миграции

**Тестирование:**
- `make test` - запуск тестов
- `make test-coverage` - тесты с покрытием

## 📚 Документация

### API Документация

После запуска приложения, Swagger UI доступен по адресу:

```
http://localhost:5050/swagger/
```
