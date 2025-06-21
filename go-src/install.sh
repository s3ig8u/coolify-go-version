#!/bin/bash
set -e

# Coolify Go Installation Script
REPO="s3ig8u/coolify-go-version"
REGISTRY="shrtso.azurecr.io/coolify-go"
GITHUB_REPO="https://github.com/s3ig8u/coolify-go-version.git"
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
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "")

# Use fallback version if API call failed
if [ -z "$LATEST_VERSION" ]; then
    LATEST_VERSION="v1.4.0"
    echo -e "${YELLOW}⚠️  GitHub API unavailable, using fallback version: $LATEST_VERSION${NC}"
else
    echo -e "${GREEN}✅ Version: $LATEST_VERSION${NC}"
fi

# Create basic directories
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p /data/coolify-go/{source,ssh,applications,databases}

# Install Docker using official method
if ! command -v docker >/dev/null 2>&1; then
    echo -e "${BLUE}🐳 Installing Docker (official method)...${NC}"
    
    # Use Docker's official installation script
    curl -fsSL https://get.docker.com | sh >/dev/null 2>&1
    
    # Start and enable Docker
    systemctl enable docker >/dev/null 2>&1
    systemctl start docker >/dev/null 2>&1
    
    echo -e "${GREEN}✅ Docker installed successfully${NC}"
else
    echo -e "${GREEN}✅ Docker already installed${NC}"
fi

# Install docker-compose if not present
if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
    echo -e "${BLUE}📦 Installing docker-compose...${NC}"
    
    # Try to install docker-compose-plugin first
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update -y >/dev/null 2>&1
        apt-get install -y docker-compose-plugin >/dev/null 2>&1 || {
            # Fallback to standalone installation
            curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose >/dev/null 2>&1
            chmod +x /usr/local/bin/docker-compose >/dev/null 2>&1
        }
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y docker-compose >/dev/null 2>&1 || {
            # Fallback to standalone installation
            curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose >/dev/null 2>&1
            chmod +x /usr/local/bin/docker-compose >/dev/null 2>&1
        }
    else
        # Direct standalone installation
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose >/dev/null 2>&1
        chmod +x /usr/local/bin/docker-compose >/dev/null 2>&1
    fi
    
    echo -e "${GREEN}✅ Docker Compose installed${NC}"
else
    echo -e "${GREEN}✅ Docker Compose already available${NC}"
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
sleep 3

# Create environment file
echo -e "${BLUE}⚙️  Creating configuration...${NC}"

# Generate safe passwords (alphanumeric only to avoid special characters)
DB_PASS=$(openssl rand -hex 16 2>/dev/null || echo "secure_db_password_123")
REDIS_PASS=$(openssl rand -hex 16 2>/dev/null || echo "secure_redis_password_123")
JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || echo "secure_jwt_secret_here_replace_in_production")

cat > /data/coolify-go/.env << EOF
APP_NAME=Coolify-Go
APP_ENV=production
APP_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_NAME=coolify_go
DB_USER=coolify_go
DB_PASSWORD=${DB_PASS}
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASS}
JWT_SECRET=${JWT_SECRET}
EOF

# Try to pull from registry first, fallback to local build
echo -e "${BLUE}🚀 Deploying Coolify Go...${NC}"

# Try to pull from registry first, fallback to local build if needed
REGISTRY_IMAGE="$REGISTRY:latest"
echo -e "${BLUE}📦 Pulling from Azure Container Registry: $REGISTRY_IMAGE${NC}"

if docker pull "$REGISTRY_IMAGE" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Successfully pulled from Azure registry${NC}"
    USE_REGISTRY_IMAGE="$REGISTRY_IMAGE"
else
    echo -e "${YELLOW}⚠️  Registry image not available, building from source...${NC}"
    
    # Clone the repository to build locally as fallback
    cd /tmp
    if git clone "$GITHUB_REPO" coolify-go-source >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Source code cloned successfully${NC}"
        cd coolify-go-source/go-src
        
        # Build the image locally
        echo -e "${BLUE}🔨 Building Docker image locally...${NC}"
        if docker build -t coolify-go:latest . >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Local build successful${NC}"
            USE_REGISTRY_IMAGE="coolify-go:latest"
        else
            echo -e "${RED}❌ Local build failed${NC}"
            exit 1
        fi
        
        # Clean up source code
        cd /
        rm -rf /tmp/coolify-go-source
    else
        echo -e "${RED}❌ Failed to clone repository. Check internet connection.${NC}"
        exit 1
    fi
fi

# Create docker-compose.yml
cd /data/coolify-go
cat > docker-compose.yml << EOF
version: '3.8'
services:
  coolify-go:
    image: $USE_REGISTRY_IMAGE
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
else
    docker compose up -d
fi

# Wait for service to be ready
echo -e "${BLUE}🔍 Waiting for service to be ready...${NC}"
for i in {1..60}; do
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

# Get external IP (IPv4 only)
EXTERNAL_IP=$(curl -4 -s https://ifconfig.me 2>/dev/null || curl -4 -s https://ipinfo.io/ip 2>/dev/null || curl -s https://api.ipify.org 2>/dev/null || echo "your-vps-ip")

# Handle IPv6 addresses by wrapping in brackets
if [[ $EXTERNAL_IP == *":"* ]] && [[ $EXTERNAL_IP != "your-vps-ip" ]]; then
    EXTERNAL_URL="http://[$EXTERNAL_IP]:8080"
    HEALTH_URL="http://[$EXTERNAL_IP]:8080/health"
else
    EXTERNAL_URL="http://$EXTERNAL_IP:8080"
    HEALTH_URL="http://$EXTERNAL_IP:8080/health"
fi

echo ""
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo -e "${BLUE}🌐 Access your application at:${NC}"
echo -e "   Local:    http://localhost:8080"
echo -e "   External: $EXTERNAL_URL"
echo ""
echo -e "${BLUE}📊 Health check: $HEALTH_URL${NC}"
echo -e "${BLUE}📁 Data directory: /data/coolify-go${NC}"
echo -e "${BLUE}⚙️  Configuration: /data/coolify-go/.env${NC}"
echo ""
echo -e "${YELLOW}📚 For troubleshooting:${NC}"
echo -e "   docker logs coolify-go"
echo -e "   docker ps"
echo -e "   docker-compose logs"
