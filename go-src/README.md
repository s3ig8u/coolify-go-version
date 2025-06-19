# Coolify Go

A Go-based port of Coolify - an open-source & self-hostable alternative to Heroku / Netlify / Vercel.

> **Note**: This is a community port of Coolify written in Go. For the official Coolify project, visit [coolify.io](https://coolify.io)

## 🚀 Quick Start

### One-Line Installation
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify/v4.x/go-src/install.sh | bash

# Or update existing installation
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify/v4.x/go-src/update.sh | bash
```

### Alternative Installation Methods

#### Docker
```bash
docker run -d --name coolify-go -p 8080:8080 ghcr.io/s3ig8u/coolify-go:latest
```

#### Binary Download
```bash
# Linux AMD64
curl -L -o coolify-go https://github.com/s3ig8u/coolify/releases/latest/download/coolify-go-linux-amd64
chmod +x coolify-go
./coolify-go

# macOS ARM64
curl -L -o coolify-go https://github.com/s3ig8u/coolify/releases/latest/download/coolify-go-darwin-arm64
chmod +x coolify-go
./coolify-go
```

## 🛠️ Development with Nix

1. **Install Nix** (if not already installed):
   ```bash
   curl -L https://nixos.org/nix/install | sh
   ```

2. **Quick development cycle**:
   ```bash
   cd go-src
   make quick
   ```

3. **Visit**: http://localhost:8080

### Development Commands

```bash
# Enter Nix shell
make shell

# Start databases and enter interactive shell
make dev

# Run with live reload
make live

# Stop everything
make stop

# Build release artifacts
make release

# Check current status
make status
```

## Manual Setup (without Nix)

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### Installation

1. **Clone and setup**:
   ```bash
   cd go-src
   cp .env.example .env
   go mod tidy
   ```

2. **Start databases**:
   ```bash
   docker-compose up -d
   ```

3. **Run the application**:
   ```bash
   go run main.go
   ```

## Development

### Live Reload
```bash
# Install air for live reload
go install github.com/cosmtrek/air@latest

# Run with live reload
air
```

### Database Operations
```bash
# Connect to PostgreSQL
psql "postgres://coolify:password@localhost:5432/coolify"

# Connect to Redis
redis-cli -p 6379
```

### Testing
```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Project Structure

```
go-src/
├── main.go              # Entry point
├── go.mod               # Go dependencies
├── flake.nix           # Nix development environment
├── docker-compose.yml  # Database services
├── .env.example        # Environment variables template
├── config/             # Configuration management
├── models/             # Database models
├── handlers/           # HTTP handlers
├── middleware/         # HTTP middleware
├── services/           # Business logic
├── utils/              # Utility functions
└── migrations/         # Database migrations
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

- **Database**: PostgreSQL connection settings
- **Redis**: Cache and queue configuration
- **JWT**: Authentication secrets
- **OAuth**: GitHub, GitLab, Google providers
- **Docker**: Container management
- **SMTP**: Email notifications

## Features

- [ ] User authentication & teams
- [ ] Server management via SSH
- [ ] Application deployment
- [ ] Database provisioning
- [ ] Real-time logs & monitoring
- [ ] Webhook integrations
- [ ] Preview environments
- [ ] Backup & restore

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
