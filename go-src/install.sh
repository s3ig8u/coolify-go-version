#!/bin/bash
set -e

# Coolify Go Port - Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | bash

REPO="s3ig8u/coolify-go-version"
REGISTRY="ghcr.io/s3ig8u/coolify-go-version"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="coolify-go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Coolify Go Installation Script${NC}"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

case $OS in
    linux|darwin) ;;
    *) echo -e "${RED}❌ Unsupported OS: $OS${NC}"; exit 1 ;;
esac

echo -e "${BLUE}📋 Detected platform: $OS/$ARCH${NC}"

# Get latest version from GitHub API
echo -e "${BLUE}🔍 Finding latest version...${NC}"
if command -v curl >/dev/null 2>&1; then
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
    LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
else
    echo -e "${RED}❌ Neither curl nor wget found. Please install one of them.${NC}"
    exit 1
fi

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${YELLOW}⚠️  Could not detect latest version, using 'latest'${NC}"
    LATEST_VERSION="latest"
fi

echo -e "${GREEN}✅ Latest version: $LATEST_VERSION${NC}"

# Installation method selection
echo ""
echo -e "${BLUE}📦 Choose installation method:${NC}"
echo "1) Docker (Recommended)"
echo "2) Binary installation"
echo "3) Docker Compose"
echo ""
read -p "Enter your choice (1-3): " INSTALL_METHOD

case $INSTALL_METHOD in
    1)
        echo -e "${BLUE}🐳 Installing with Docker...${NC}"
        
        # Check if Docker is installed
        if ! command -v docker >/dev/null 2>&1; then
            echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
            echo "Visit: https://docs.docker.com/get-docker/"
            exit 1
        fi
        
        # Stop existing container if running
        docker stop coolify-go 2>/dev/null || true
        docker rm coolify-go 2>/dev/null || true
        
        # Pull and run new version
        echo -e "${BLUE}📥 Pulling Docker image...${NC}"
        docker pull "$REGISTRY:$LATEST_VERSION"
        
        echo -e "${BLUE}🏃 Starting Coolify Go...${NC}"
        docker run -d \
            --name coolify-go \
            --restart unless-stopped \
            -p 8080:8080 \
            "$REGISTRY:$LATEST_VERSION"
        
        echo -e "${GREEN}✅ Coolify Go started successfully!${NC}"
        echo -e "${BLUE}🌐 Access it at: http://localhost:8080${NC}"
        ;;
        
    2)
        echo -e "${BLUE}📦 Installing binary...${NC}"
        
        # Determine binary name
        BINARY_FILE="coolify-go-$LATEST_VERSION-$OS-$ARCH"
        if [ "$OS" = "windows" ]; then
            BINARY_FILE="$BINARY_FILE.exe"
        fi
        
        # Download URL
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$BINARY_FILE"
        
        echo -e "${BLUE}📥 Downloading from: $DOWNLOAD_URL${NC}"
        
        # Create temporary directory
        TMP_DIR=$(mktemp -d)
        cd "$TMP_DIR"
        
        # Download binary
        if command -v curl >/dev/null 2>&1; then
            curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL"
        elif command -v wget >/dev/null 2>&1; then
            wget -O "$BINARY_NAME" "$DOWNLOAD_URL"
        fi
        
        # Make executable
        chmod +x "$BINARY_NAME"
        
        # Install to system
        if [ "$EUID" -eq 0 ]; then
            mv "$BINARY_NAME" "$INSTALL_DIR/"
            echo -e "${GREEN}✅ Installed to $INSTALL_DIR/$BINARY_NAME${NC}"
        else
            echo -e "${YELLOW}⚠️  Installing to current directory (not in PATH)${NC}"
            mv "$BINARY_NAME" "./"
            echo -e "${BLUE}💡 To install system-wide, run: sudo mv $BINARY_NAME $INSTALL_DIR/${NC}"
        fi
        
        # Clean up
        cd - >/dev/null
        rm -rf "$TMP_DIR"
        
        echo -e "${GREEN}✅ Installation complete!${NC}"
        echo -e "${BLUE}🚀 Run: $BINARY_NAME${NC}"
        ;;
        
    3)
        echo -e "${BLUE}🐳 Installing with Docker Compose...${NC}"
        
        # Check if Docker Compose is installed
        if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
            echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose first.${NC}"
            exit 1
        fi
        
        # Create docker-compose.yml
        cat > docker-compose.yml << EOF
version: '3.8'

services:
  coolify-go:
    image: $REGISTRY:$LATEST_VERSION
    container_name: coolify-go
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - GO_ENV=production
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

  # Optional: Add PostgreSQL and Redis
  # postgres:
  #   image: postgres:15
  #   environment:
  #     POSTGRES_DB: coolify
  #     POSTGRES_USER: coolify
  #     POSTGRES_PASSWORD: password
  #   volumes:
  #     - postgres_data:/var/lib/postgresql/data
  #   ports:
  #     - "5432:5432"
  #
  # redis:
  #   image: redis:7-alpine
  #   volumes:
  #     - redis_data:/data
  #   ports:
  #     - "6379:6379"

# volumes:
#   postgres_data:
#   redis_data:
EOF
        
        echo -e "${GREEN}✅ Created docker-compose.yml${NC}"
        echo -e "${BLUE}🏃 Starting services...${NC}"
        
        # Start services
        if command -v docker-compose >/dev/null 2>&1; then
            docker-compose up -d
        else
            docker compose up -d
        fi
        
        echo -e "${GREEN}✅ Coolify Go started with Docker Compose!${NC}"
        echo -e "${BLUE}🌐 Access it at: http://localhost:8080${NC}"
        ;;
        
    *)
        echo -e "${RED}❌ Invalid choice${NC}"
        exit 1
        ;;
esac

# Wait for service to be ready
echo -e "${BLUE}🔍 Waiting for service to be ready...${NC}"
for i in {1..30}; do
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is ready!${NC}"
        break
    fi
    echo -n "."
    sleep 2
done

# Show status
echo ""
echo -e "${BLUE}📊 Service Status:${NC}"
curl -s http://localhost:8080/health 2>/dev/null | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health

echo ""
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo -e "${BLUE}📚 Documentation: https://docs.coolify.io${NC}"
echo -e "${BLUE}🐛 Issues: https://github.com/$REPO/issues${NC}"
echo -e "${BLUE}💬 Community: https://discord.coolify.io${NC}"
