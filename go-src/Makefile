.PHONY: build run dev docker clean test test-install version quick-test help migrate migrate-status migrate-rollback migrate-schema-info

# Variables
BINARY_NAME=coolify-go
DOCKER_IMAGE=coolify-go
VERSION?=v1.3.0

# Build the application
build:
	go build -o $(BINARY_NAME) .

# Run the application
run: build
	./$(BINARY_NAME)

# Development with live reload
dev:
	go run .

# Build Docker image
docker:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker build -t $(DOCKER_IMAGE):latest .

# Run with Docker Compose
compose-up:
	docker-compose up -d

# Stop Docker Compose
compose-down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f coolify-go

# Database migrations
migrate:
	go run cmd/migrate/main.go -action=migrate

migrate-status:
	go run cmd/migrate/main.go -action=status

migrate-schema-info:
	go run cmd/migrate/main.go -action=schema-info

migrate-rollback:
	go run cmd/migrate/main.go -action=rollback -step=1

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	docker image prune -f

# Test the application
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy
	go mod download

# Test installation process
test-install:
	chmod +x test-install.sh
	./test-install.sh

# Version information
version:
	@echo "$(VERSION)"

# Quick test (build + version check)
quick-test: build
	./$(BINARY_NAME) -version

# Show help
help:
	@echo "Available commands:"
	@echo "  build        - Build the binary"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Run in development mode"
	@echo "  docker       - Build Docker image"
	@echo "  compose-up   - Start with Docker Compose"
	@echo "  compose-down - Stop Docker Compose"
	@echo "  logs         - View application logs"
	@echo "  migrate      - Run database migrations"
	@echo "  migrate-status - Check migration status"
	@echo "  migrate-schema-info - Show schema hash information"
	@echo "  migrate-rollback - Rollback last migration"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-install - Test installation process"
	@echo "  fmt          - Format code"
	@echo "  deps         - Install dependencies"
	@echo "  version      - Show version"
	@echo "  quick-test   - Quick build and version test"
