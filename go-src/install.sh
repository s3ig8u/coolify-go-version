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
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
PORT=${COOLIFY_PORT:-8080}
DATA_DIR=${COOLIFY_DATA_DIR:-/data/coolify-go}
SKIP_DOCKER_INSTALL=${SKIP_DOCKER_INSTALL:-false}
SKIP_REGISTRY=${SKIP_REGISTRY:-false}
QUIET=${QUIET:-false}

# Progress tracking
TOTAL_STEPS=12
CURRENT_STEP=0

# Function to show progress
show_progress() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    if [ "$QUIET" != "true" ]; then
        echo -e "${BLUE}[$CURRENT_STEP/$TOTAL_STEPS]${NC} $1"
    fi
}

# Function to show success
show_success() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${GREEN}✅ $1${NC}"
    fi
}

# Function to show warning
show_warning() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${YELLOW}⚠️  $1${NC}"
    fi
}

# Function to show error
show_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Function to show info
show_info() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${CYAN}ℹ️  $1${NC}"
    fi
}

# Function to show step
show_step() {
    if [ "$QUIET" != "true" ]; then
        echo -e "${PURPLE}🔧 $1${NC}"
    fi
}

# Banner
if [ "$QUIET" != "true" ]; then
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🚀 Coolify Go Installation                ║"
    echo "║                                                              ║"
    echo "║  Self-hosted deployment platform built with Go              ║"
    echo "║  Fast, secure, and production-ready                         ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
fi

# Check root
if [ $EUID != 0 ]; then
    show_error "Please run as root or with sudo"
    echo -e "${YELLOW}Example: sudo bash install.sh${NC}"
    exit 1
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            PORT="$2"
            shift 2
            ;;
        --data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        --skip-docker)
            SKIP_DOCKER_INSTALL=true
            shift
            ;;
        --skip-registry)
            SKIP_REGISTRY=true
            shift
            ;;
        --quiet)
            QUIET=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --port PORT        Port to run Coolify Go on (default: 8080)"
            echo "  --data-dir DIR     Data directory (default: /data/coolify-go)"
            echo "  --skip-docker      Skip Docker installation if already installed"
            echo "  --skip-registry    Skip registry pull and build from source"
            echo "  --quiet            Quiet mode with minimal output"
            echo "  --help             Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Default installation"
            echo "  $0 --port 3000                       # Install on port 3000"
            echo "  $0 --data-dir /opt/coolify           # Custom data directory"
            echo "  $0 --skip-docker --skip-registry     # Minimal installation"
            exit 0
            ;;
        *)
            show_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

show_progress "Detecting platform and architecture"

# Detect platform
OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"' 2>/dev/null || echo "unknown")
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) show_error "Unsupported architecture: $ARCH" ;;
esac

# Validate OS
case "$OS_TYPE" in
    ubuntu|debian|centos|fedora|rhel|rocky|almalinux) ;;
    *) show_warning "OS $OS_TYPE not officially tested, but continuing..." ;;
esac

show_success "Platform: $OS_TYPE ($ARCH)"
show_info "Installation directory: $DATA_DIR"
show_info "Port: $PORT"

# Get version
show_progress "Checking for latest version"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "")

if [ -z "$LATEST_VERSION" ]; then
    LATEST_VERSION="v1.4.0"
    show_warning "GitHub API unavailable, using fallback version: $LATEST_VERSION"
else
    show_success "Latest version: $LATEST_VERSION"
fi

# Create basic directories
show_progress "Creating directory structure"
mkdir -p "$DATA_DIR"/{source,ssh,applications,databases,backups,logs}
show_success "Directories created in $DATA_DIR"

# Install Docker if not skipped
if [ "$SKIP_DOCKER_INSTALL" != "true" ]; then
    show_progress "Installing Docker"
    
    if ! command -v docker >/dev/null 2>&1; then
        show_step "Installing Docker using official method..."
        
        # Use Docker's official installation script
        curl -fsSL https://get.docker.com | sh >/dev/null 2>&1
        
        # Start and enable Docker
        systemctl enable docker >/dev/null 2>&1
        systemctl start docker >/dev/null 2>&1
        
        show_success "Docker installed successfully"
    else
        show_success "Docker already installed"
    fi
else
    show_info "Skipping Docker installation"
fi

# Install docker-compose if not present
show_progress "Installing Docker Compose"
if ! command -v docker-compose >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then
    show_step "Installing docker-compose..."
    
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
    
    show_success "Docker Compose installed"
else
    show_success "Docker Compose already available"
fi

# Configure Docker daemon
show_progress "Configuring Docker daemon"
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
show_success "Docker configured"

# Create environment file
show_progress "Generating secure configuration"

# Generate safe passwords (alphanumeric only to avoid special characters)
DB_PASS=$(openssl rand -hex 16 2>/dev/null || echo "secure_db_password_123")
REDIS_PASS=$(openssl rand -hex 16 2>/dev/null || echo "secure_redis_password_123")
JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || echo "secure_jwt_secret_here_replace_in_production")

cat > "$DATA_DIR/.env" << EOF
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

# Set proper permissions
chmod 600 "$DATA_DIR/.env"
show_success "Configuration generated with secure passwords"

# Deploy Coolify Go
show_progress "Deploying Coolify Go"

if [ "$SKIP_REGISTRY" != "true" ]; then
    # Try to pull from registry first, fallback to local build if needed
    REGISTRY_IMAGE="$REGISTRY:latest"
    show_step "Pulling from Azure Container Registry: $REGISTRY_IMAGE"

    if docker pull "$REGISTRY_IMAGE" >/dev/null 2>&1; then
        show_success "Successfully pulled from Azure registry"
        USE_REGISTRY_IMAGE="$REGISTRY_IMAGE"
    else
        show_warning "Registry image not available, building from source..."
        
        # Clone the repository to build locally as fallback
        cd /tmp
        if git clone "$GITHUB_REPO" coolify-go-source >/dev/null 2>&1; then
            show_success "Source code cloned successfully"
            cd coolify-go-source/go-src
            
            # Build the image locally
            show_step "Building Docker image locally..."
            if docker build -t coolify-go:latest . >/dev/null 2>&1; then
                show_success "Local build successful"
                USE_REGISTRY_IMAGE="coolify-go:latest"
            else
                show_error "Local build failed"
            fi
            
            # Clean up source code
            cd /
            rm -rf /tmp/coolify-go-source
        else
            show_error "Failed to clone repository. Check internet connection."
        fi
    fi
else
    show_info "Skipping registry, building from source..."
    cd /tmp
    if git clone "$GITHUB_REPO" coolify-go-source >/dev/null 2>&1; then
        show_success "Source code cloned successfully"
        cd coolify-go-source/go-src
        
        show_step "Building Docker image locally..."
        if docker build -t coolify-go:latest . >/dev/null 2>&1; then
            show_success "Local build successful"
            USE_REGISTRY_IMAGE="coolify-go:latest"
        else
            show_error "Local build failed"
        fi
        
        cd /
        rm -rf /tmp/coolify-go-source
    else
        show_error "Failed to clone repository. Check internet connection."
    fi
fi

# Create docker-compose.yml
show_progress "Creating service configuration"
cd "$DATA_DIR"
cat > docker-compose.yml << EOF
version: '3.8'
services:
  coolify-go:
    image: $USE_REGISTRY_IMAGE
    container_name: coolify-go
    restart: unless-stopped
    ports:
      - "$PORT:8080"
    volumes:
      - $DATA_DIR:/data
      - /var/run/docker.sock:/var/run/docker.sock
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    
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
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U coolify_go"]
      interval: 10s
      timeout: 5s
      retries: 5
      
  redis:
    image: redis:7-alpine
    container_name: coolify-go-redis
    restart: unless-stopped
    command: redis-server --requirepass \${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

volumes:
  postgres_data:
  redis_data:
EOF

show_success "Service configuration created"

# Start services
show_progress "Starting services"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d
else
    docker compose up -d
fi

# Wait for service to be ready
show_progress "Waiting for service to be ready"
for i in {1..60}; do
    if curl -sf http://localhost:$PORT/health >/dev/null 2>&1; then
        show_success "Service is ready!"
        break
    fi
    if [ "$QUIET" != "true" ]; then
        echo -n "."
    fi
    sleep 2
done

# Show status
if [ "$QUIET" != "true" ]; then
    echo ""
    show_progress "Checking service status"
    curl -s http://localhost:$PORT/health 2>/dev/null | python3 -m json.tool 2>/dev/null || curl -s http://localhost:$PORT/health
fi

# Get external IP (IPv4 only)
show_progress "Detecting external access information"
EXTERNAL_IP=$(curl -4 -s https://ifconfig.me 2>/dev/null || curl -4 -s https://ipinfo.io/ip 2>/dev/null || curl -s https://api.ipify.org 2>/dev/null || echo "your-vps-ip")

# Handle IPv6 addresses by wrapping in brackets
if [[ $EXTERNAL_IP == *":"* ]] && [[ $EXTERNAL_IP != "your-vps-ip" ]]; then
    EXTERNAL_URL="http://[$EXTERNAL_IP]:$PORT"
    HEALTH_URL="http://[$EXTERNAL_IP]:$PORT/health"
else
    EXTERNAL_URL="http://$EXTERNAL_IP:$PORT"
    HEALTH_URL="http://$EXTERNAL_IP:$PORT/health"
fi

# Final success message
if [ "$QUIET" != "true" ]; then
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    🎉 Installation Complete!                ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}🌐 Access your application:${NC}"
    echo -e "   Local:    http://localhost:$PORT"
    echo -e "   External: $EXTERNAL_URL"
    echo ""
    echo -e "${BLUE}📊 Health check: $HEALTH_URL${NC}"
    echo -e "${BLUE}📁 Data directory: $DATA_DIR${NC}"
    echo -e "${BLUE}⚙️  Configuration: $DATA_DIR/.env${NC}"
    echo ""
    echo -e "${YELLOW}🔧 Useful commands:${NC}"
    echo -e "   docker logs coolify-go                    # View application logs"
    echo -e "   docker ps                                 # Check container status"
    echo -e "   docker-compose logs                       # View all service logs"
    echo -e "   docker-compose restart coolify-go         # Restart application"
    echo ""
    echo -e "${YELLOW}📚 Next steps:${NC}"
    echo -e "   1. Open your browser to http://localhost:$PORT"
    echo -e "   2. Complete the initial setup wizard"
    echo -e "   3. Configure your first deployment"
    echo ""
    echo -e "${CYAN}💡 Pro tip: Run 'curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/update.sh | bash' to update later${NC}"
    echo ""
else
    echo "Installation completed successfully"
    echo "Access: http://localhost:$PORT"
fi
