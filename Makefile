.PHONY: help build run test lint migrate-up migrate-down docker-up docker-down docker-logs clean

# Default target
help:
	@echo "SIGIF-GO - Inventory & Sales Management System"
	@echo ""
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make lint          - Run linter"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make docker-up     - Start all services with Docker Compose"
	@echo "  make docker-down   - Stop all Docker services"
	@echo "  make docker-logs   - View Docker logs"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make generate      - Generate code (mocks, etc.)"

# Build the application
build:
	CGO_ENABLED=1 go build -o bin/sigif-server ./cmd/server
	CGO_ENABLED=1 go build -o bin/sigif-migrate ./cmd/migrate

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run ./...

# Run migrations
migrate-up:
	go run ./cmd/migrate

# Docker commands
docker-up:
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker/docker-compose.yml down

docker-logs:
	docker-compose -f deployments/docker/docker-compose.yml logs -f

docker-build:
	docker-compose -f deployments/docker/docker-compose.yml build

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate code
generate:
	go generate ./...

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet lint test

# Development setup
dev-setup: install-tools deps docker-up
	@echo "Waiting for database to be ready..."
	@sleep 10
	@make migrate-up
	@echo "Development environment ready!"

# Production build
prod-build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/sigif-server ./cmd/server
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/sigif-migrate ./cmd/migrate