.PHONY: swagger dev build docker_dev

include .env

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

swagger:
	swag init -g cmd/app/main.go

dev: swagger
	go run cmd/app/main.go

build: swagger
	go build -o bin/app cmd/app/main.go

docker-up:
	docker-compose -f docker-compose.dev.yml up -d

docker-down:
	docker-compose -f docker-compose.dev.yml down

migrate-create:
	goose -dir postgres/migrations create $(NAME) sql

migrate-up:
	goose -dir postgres/migrations postgres "$(POSTGRES_URL)" up

migrate-down:
	goose -dir postgres/migrations postgres "$(POSTGRES_URL)" down

sql-generate:
	sqlc generate

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...