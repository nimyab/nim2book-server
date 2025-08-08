.PHONY: swagger dev build docker_dev

include .env

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