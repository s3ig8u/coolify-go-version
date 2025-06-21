# Coolify Go

A Go-based port of Coolify - an open-source & self-hostable alternative to Heroku / Netlify / Vercel.

> **Note**: This is a minimal Go port of Coolify. For the official Coolify project, visit [coolify.io](https://coolify.io)

## 🚀 Quick Start

### One-Line Installation
```bash
# Install with enhanced script (requires root)
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash
```

### Alternative Installation Methods

#### Docker
```bash
docker run -d --name coolify-go -p 8080:8080 ghcr.io/s3ig8u/coolify-go-version:latest
```

#### Docker Compose (Recommended)
```bash
# Creates full stack with PostgreSQL and Redis
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash
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

### Quick Development
```bash
# Install dependencies
make deps

# Run in development mode
make dev

# Build the application
make build

# Run the built binary
make run

# View all available commands
make help
```

### Database Operations
```bash
# Connect to PostgreSQL (when running via install script)
psql "postgres://coolify_go:$(grep DB_PASSWORD /data/coolify-go/.env | cut -d= -f2)@localhost:5432/coolify_go"

# Connect to Redis
redis-cli -p 6379 -a $(grep REDIS_PASSWORD /data/coolify-go/.env | cut -d= -f2)
```

### Docker Development
```bash
# Build Docker image
make docker

# Start with Docker Compose
make compose-up

# View logs
make logs

# Stop services
make compose-down
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
# Test comment
