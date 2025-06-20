#!/bin/bash
set -e

# Coolify Go Installation Script
REPO="s3ig8u/coolify-go-version"
REGISTRY="ghcr.io/s3ig8u/coolify-go-version"
DOCKER_VERSION="27.0"
DATE=$(date +"%Y%m%d-%H%M%S")

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Coolify Go Installation${NC}"

# Check root
if [ $EUID != 0 ]; then
    echo -e "${RED}❌ Please run as root or with sudo${NC}"
    exit 1
fi

# Detect platform
OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"' 2>/dev/null || echo "unknown")
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

# Validate OS
case "$OS_TYPE" in
    ubuntu|debian|centos|fedora|rhel|rocky|almalinux) ;;
    *) echo -e "${RED}❌ Unsupported OS: $OS_TYPE${NC}"; exit 1 ;;
esac

echo -e "${BLUE}📋 Platform: $OS_TYPE ($ARCH)${NC}"

# Get version
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "v1.3.0")
echo -e "${GREEN}✅ Version: $LATEST_VERSION${NC}"

# Create basic directories
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p /data/coolify-go/{source,ssh,applications,databases}

# Install Docker if needed
if ! command -v docker >/dev/null 2>&1; then
    echo -e "${BLUE}🐳 Installing Docker...${NC}"
    case "$OS_TYPE" in
        ubuntu|debian)
            apt-get update -y >/dev/null
            apt-get install -y docker.io >/dev/null
            systemctl enable docker >/dev/null
            systemctl start docker >/dev/null
            ;;
        centos|fedora|rhel|rocky|almalinux)
            dnf install -y docker >/dev/null
            systemctl enable docker >/dev/null
            systemctl start docker >/dev/null
            ;;
    esac
fi

# Configure Docker daemon
echo -e "${BLUE}🔧 Configuring Docker...${NC}"
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker >/dev/null 2>&1

# Create environment file
echo -e "${BLUE}⚙️  Creating configuration...${NC}"
cat > /data/coolify-go/.env << EOF
APP_NAME=Coolify-Go
APP_ENV=production
APP_PORT=8080
DB_PASSWORD=$(openssl rand -base64 32 2>/dev/null || echo "changeme")
REDIS_PASSWORD=$(openssl rand -base64 32 2>/dev/null || echo "changeme")
JWT_SECRET=$(openssl rand -base64 64 2>/dev/null || echo "changeme")
EOF

# Deploy with Docker Compose
echo -e "${BLUE}🚀 Deploying Coolify Go...${NC}"
cat > /data/coolify-go/docker-compose.yml << EOF
version: '3.8'
services:
  coolify-go:
    image: $REGISTRY:$LATEST_VERSION
    container_name: coolify-go
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - /data/coolify-go:/data
      - /var/run/docker.sock:/var/run/docker.sock
    env_file:
      - .env
    
  postgres:
    image: postgres:15
    container_name: coolify-go-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: coolify_go
      POSTGRES_USER: coolify_go
      POSTGRES_PASSWORD: \${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
      
  redis:
    image: redis:7-alpine
    container_name: coolify-go-redis
    restart: unless-stopped
    command: redis-server --requirepass \${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

volumes:
  postgres_data:
  redis_data:
EOF

cd /data/coolify-go
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d
else
    docker compose up -d
fi

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
