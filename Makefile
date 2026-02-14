.PHONY: swagger dev build docker_dev install-tools sql-gen test test-coverage

include .env

# Install development tools
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Swagger documentation generation
swagger:
	swag init -g cmd/app/main.go

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

# Goose migration commands
migrate-create:
	goose -dir db/migrations create $(NAME) sql

migrate-up:
	goose -dir db/migrations postgres "$(POSTGRES_URL)" up

migrate-down:
	goose -dir db/migrations postgres "$(POSTGRES_URL)" down

# SQLC commands
# Generate Go code from SQL queries
sqlc-gen:
	sqlc generate

# Verify sqlc configuration and queries
sqlc-verify:
	sqlc verify

# Compile queries to check for errors
sqlc-compile:
	sqlc compile

# Testing commands
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...

altlasdiff:
	atlas migrate diff $(name) \
	  --dir "file://ent/migrate/migrations" \
  	--to "ent://ent/schema" \
 	 	--dev-url "docker://postgres/17" 

atlasapply:
	atlas migrate apply \
  --dir "file://ent/migrate/migrations" \
  --url "$(POSTGRES_URL)"

ent-generate:
	go generate ./ent
