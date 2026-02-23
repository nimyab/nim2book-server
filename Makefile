.PHONY: swagger dev build install-tools test test-coverage atlasdiff atlasapply ent-generate mocks

include .env

# Git hooks
install-hooks:
	go tool lefthook install

# Mocks generation
mocks:
	go run github.com/vektra/mockery/v3@latest

# Swagger documentation generation
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go

# Development server
dev: swagger
	go run cmd/app/main.go

# Build the application
build: swagger
	go build -o bin/app cmd/app/main.go

# Docker commands
docker-up:
	docker-compose -f docker-compose.dev.yml up -d

docker-down:
	docker-compose -f docker-compose.dev.yml down

# Testing commands
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...

altlasdiff:
	atlas migrate diff $(name) --dir "file://ent/migrate/migrations" --to "ent://ent/schema" --dev-url "docker://postgres/17" 

atlasapply:
	atlas migrate apply --dir "file://ent/migrate/migrations" --url "$(POSTGRES_URL)"

ent-generate:
	go generate ./ent
