# Coolify Go Development Makefile

.PHONY: shell dev start stop clean test build release deploy update version

# Enter Nix development shell
shell:
	./shell.sh

# Start development environment (databases + shell)
dev:
	./shell.sh -c "docker-compose up -d && exec bash"

# Start the Go server (inside Nix)
start:
	./shell.sh -c "go run main.go"

# Start with live reload (inside Nix)
live:
	./shell.sh -c "air"

# Stop all services
stop:
	docker-compose down

# Clean up containers and volumes
clean:
	docker-compose down -v
	docker system prune -f

# Run tests (inside Nix)
test:
	./shell.sh -c "go test ./..."

# Build binary (inside Nix)
build:
	./shell.sh -c "go build -o coolify-go ."

# Quick development cycle
quick:
	docker-compose up -d
	./shell.sh -c "go run main.go"

# Build release (cross-platform binaries + Docker)
release:
	./shell.sh -c "./scripts/build.sh"

# Deploy to environment (usage: make deploy ENV=staging VERSION=v1.0.1)
deploy:
	./scripts/deploy.sh $(ENV) $(VERSION)

# Update running application with zero downtime
update:
	@echo "🔄 Updating application..."
	./scripts/build.sh
	./scripts/deploy.sh staging latest

# Show version information
version:
	./shell.sh -c "go run . -version"

# Show current running version
status:
	@echo "📊 Application Status:"
	@curl -s http://localhost:8080/health | jq . || echo "Service not running"

# Show help
help:
	@echo "Coolify Go Development Commands:"
	@echo "  make shell     - Enter Nix development shell"
	@echo "  make dev       - Start databases and enter shell"
	@echo "  make start     - Run Go server"
	@echo "  make live      - Run with live reload (air)"
	@echo "  make stop      - Stop all services"
	@echo "  make clean     - Clean containers and volumes"
	@echo "  make test      - Run tests"
	@echo "  make build     - Build binary"
	@echo "  make quick     - Quick start (databases + server)"
	@echo "  make release   - Build release (binaries + Docker)"
	@echo "  make deploy    - Deploy to environment"
	@echo "  make update    - Update running application"
	@echo "  make version   - Show version information"
	@echo "  make status    - Show current running version"
