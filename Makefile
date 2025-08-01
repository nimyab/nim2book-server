.PHONY: swagger dev build docker_dev

swagger:
	swag init -g cmd/app/main.go

dev: swagger
	go run cmd/app/main.go

build: swagger
	go build -o bin/app cmd/app/main.go

dev-up:
	docker-compose -f docker-compose.dev.yml up -d

dev-down:
	docker-comose -f docker-compose.dev.yml down