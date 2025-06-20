#!/bin/bash
set -e

# Simple Coolify Go Installation Script (requires Docker pre-installed)
REGISTRY="shrtso.azurecr.io/coolify-go"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Coolify Go Simple Installation${NC}"

# Check if Docker is available
if ! command -v docker >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first:${NC}"
    echo -e "${YELLOW}curl -fsSL https://get.docker.com | sh${NC}"
    exit 1
fi

# Check if running as root or user is in docker group
if [ "$EUID" -ne 0 ] && ! groups $USER | grep -q docker; then
    echo -e "${RED}❌ Please run as root or add user to docker group:${NC}"
    echo -e "${YELLOW}sudo usermod -aG docker \$USER && newgrp docker${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker is available${NC}"

# Create directories
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p /data/coolify-go/{source,ssh,applications,databases}

# Create environment file
echo -e "${BLUE}⚙️  Creating configuration...${NC}"
cat > /data/coolify-go/.env << EOF
APP_NAME=Coolify-Go
APP_ENV=production
APP_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_NAME=coolify_go
DB_USER=coolify_go
DB_PASSWORD=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
JWT_SECRET=$(openssl rand -base64 64 2>/dev/null || date +%s | sha256sum | base64)
EOF

# Try to pull from Azure Container Registry
echo -e "${BLUE}📦 Pulling from Azure Container Registry: $REGISTRY:latest${NC}"

if docker pull "$REGISTRY:latest" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Successfully pulled from Azure registry${NC}"
    USE_IMAGE="$REGISTRY:latest"
else
    echo -e "${YELLOW}⚠️  Azure registry not accessible, using fallback...${NC}"
    
    # Try to build locally
    if [ -f "Dockerfile" ]; then
        echo -e "${BLUE}🔨 Building locally...${NC}"
        docker build -t coolify-go:latest . >/dev/null 2>&1
        USE_IMAGE="coolify-go:latest"
    else
        echo -e "${RED}❌ No Dockerfile found and registry unavailable${NC}"
        exit 1
    fi
fi

# Create docker-compose.yml
cd /data/coolify-go
cat > docker-compose.yml << EOF
version: '3.8'
services:
  coolify-go:
    image: $USE_IMAGE
    container_name: coolify-go
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - /data/coolify-go:/data
      - /var/run/docker.sock:/var/run/docker.sock
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    
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

# Start services
echo -e "${BLUE}🚀 Starting services...${NC}"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d
elif docker compose version >/dev/null 2>&1; then
    docker compose up -d
else
    echo -e "${RED}❌ Docker Compose not available${NC}"
    exit 1
fi

# Wait for service
echo -e "${BLUE}🔍 Waiting for service...${NC}"
for i in {1..30}; do
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is ready!${NC}"
        break
    fi
    echo -n "."
    sleep 2
done

echo ""
echo -e "${GREEN}🎉 Installation completed!${NC}"
echo -e "${BLUE}🌐 Access: http://localhost:8080${NC}"
echo -e "${BLUE}📊 Health: http://localhost:8080/health${NC}"
