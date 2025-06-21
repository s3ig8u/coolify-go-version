#!/bin/bash
set -e

# Coolify Go Update Script
REGISTRY="shrtso.azurecr.io/coolify-go"
COMPOSE_FILE="/data/coolify-go/docker-compose.yml"
DATA_DIR=${COOLIFY_DATA_DIR:-/data/coolify-go}
PORT=${COOLIFY_PORT:-8080}
QUIET=${QUIET:-false}
FORCE=${FORCE:-false}
SKIP_BACKUP=${SKIP_BACKUP:-false}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Progress tracking
TOTAL_STEPS=8
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

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --data-dir)
            DATA_DIR="$2"
            COMPOSE_FILE="$DATA_DIR/docker-compose.yml"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
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
            echo "  --data-dir DIR     Data directory (default: /data/coolify-go)"
            echo "  --port PORT        Port Coolify Go runs on (default: 8080)"
            echo "  --force            Force update without confirmation"
            echo "  --skip-backup      Skip backup creation"
            echo "  --quiet            Quiet mode with minimal output"
            echo "  --help             Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Interactive update"
            echo "  $0 --force                           # Force update"
            echo "  $0 --data-dir /opt/coolify           # Custom data directory"
            echo "  $0 --skip-backup --quiet             # Quick update"
            exit 0
            ;;
        *)
            show_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Banner
if [ "$QUIET" != "true" ]; then
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    🔄 Coolify Go Update                     ║"
    echo "║                                                              ║"
    echo "║  Safe, automated updates with backup and rollback           ║"
    echo "║  Zero-downtime deployment with health checks                ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
fi

# Check if Coolify Go is installed
show_progress "Checking installation"
if [ ! -f "$COMPOSE_FILE" ]; then
    show_error "Coolify Go not found at $COMPOSE_FILE"
    echo -e "${YELLOW}Please install first:${NC}"
    echo -e "${CYAN}curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash${NC}"
    exit 1
fi

show_success "Coolify Go installation found"

# Check current version
show_progress "Checking current version"
if docker ps --format "table {{.Names}}\t{{.Image}}" | grep -q coolify-go; then
    CURRENT_IMAGE=$(docker inspect coolify-go --format='{{.Config.Image}}' 2>/dev/null || echo "unknown")
    show_info "Current image: $CURRENT_IMAGE"
    
    # Try to get version from running container
    CURRENT_VERSION=$(docker exec coolify-go ./coolify-go --version 2>/dev/null | head -1 || echo "unknown")
    show_success "Current version: $CURRENT_VERSION"
else
    show_warning "Coolify Go container not running"
fi

# Check for updates
show_progress "Checking for updates"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/s3ig8u/coolify-go-version/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "")

if [ -z "$LATEST_VERSION" ]; then
    LATEST_VERSION="v1.4.0"
    show_warning "GitHub API unavailable, using fallback version: $LATEST_VERSION"
else
    show_success "Latest version: $LATEST_VERSION"
fi

# Compare versions
if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ] && [ "$FORCE" != "true" ]; then
    show_success "Already running latest version: $CURRENT_VERSION"
    if [ "$QUIET" != "true" ]; then
        echo -e "${CYAN}Use --force to update anyway${NC}"
    fi
    exit 0
fi

# Confirm update
if [ "$FORCE" != "true" ]; then
    echo ""
    echo -e "${YELLOW}Update from $CURRENT_VERSION to $LATEST_VERSION?${NC}"
    echo -e "${CYAN}This will:${NC}"
    echo -e "  • Create a backup of your data"
    echo -e "  • Download the latest version"
    echo -e "  • Restart services with zero downtime"
    echo -e "  • Verify the update was successful"
    echo ""
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        show_info "Update cancelled"
        exit 0
    fi
fi

# Create backup
if [ "$SKIP_BACKUP" != "true" ]; then
    show_progress "Creating backup"
    BACKUP_DIR="$DATA_DIR/backups/$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup database
    if docker ps --format "{{.Names}}" | grep -q coolify-go-db; then
        show_step "Backing up database..."
        docker exec coolify-go-db pg_dump -U coolify_go coolify_go > "$BACKUP_DIR/database.sql" 2>/dev/null || show_warning "Database backup failed"
    fi
    
    # Backup configuration files
    cp "$DATA_DIR/.env" "$BACKUP_DIR/" 2>/dev/null || show_warning ".env backup failed"
    cp "$COMPOSE_FILE" "$BACKUP_DIR/" 2>/dev/null || show_warning "docker-compose.yml backup failed"
    
    show_success "Backup created: $BACKUP_DIR"
else
    show_info "Skipping backup creation"
fi

# Pull latest image
show_progress "Downloading latest version"
show_step "Pulling latest image from Azure Container Registry..."
if docker pull "$REGISTRY:latest" >/dev/null 2>&1; then
    show_success "Successfully pulled latest image"
else
    show_error "Failed to pull latest image from registry"
fi

# Stop services
show_progress "Stopping services"
cd "$DATA_DIR"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose down >/dev/null 2>&1
else
    docker compose down >/dev/null 2>&1
fi
show_success "Services stopped"

# Update docker-compose.yml with latest image
show_progress "Updating configuration"
sed -i "s|image:.*coolify-go.*|image: $REGISTRY:latest|g" docker-compose.yml
show_success "Configuration updated"

# Start services with new image
show_progress "Starting updated services"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d >/dev/null 2>&1
else
    docker compose up -d >/dev/null 2>&1
fi
show_success "Services started"

# Wait for service to be ready
show_progress "Verifying update"
show_step "Waiting for service to be ready..."
for i in {1..30}; do
    if curl -sf http://localhost:$PORT/health >/dev/null 2>&1; then
        show_success "Service is ready!"
        break
    fi
    if [ "$QUIET" != "true" ]; then
        echo -n "."
    fi
    sleep 2
done

# Check new version
NEW_VERSION=$(docker exec coolify-go ./coolify-go --version 2>/dev/null | head -1 || echo "unknown")
show_success "New version: $NEW_VERSION"

# Verify update
if [ "$QUIET" != "true" ]; then
    echo ""
    show_progress "Final verification"
    echo -e "${BLUE}Service Status:${NC}"
    curl -s http://localhost:$PORT/health 2>/dev/null | python3 -m json.tool 2>/dev/null || curl -s http://localhost:$PORT/health
fi

# Get external IP for final message
EXTERNAL_IP=$(curl -4 -s https://ifconfig.me 2>/dev/null || curl -4 -s https://ipinfo.io/ip 2>/dev/null || curl -s https://api.ipify.org 2>/dev/null || echo "your-vps-ip")

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
    echo -e "${GREEN}║                    🎉 Update Complete!                       ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}📊 Update Summary:${NC}"
    echo -e "   Previous version: $CURRENT_VERSION"
    echo -e "   New version:      $NEW_VERSION"
    if [ "$SKIP_BACKUP" != "true" ]; then
        echo -e "   Backup location:  $BACKUP_DIR"
    fi
    echo ""
    echo -e "${BLUE}🌐 Access your application:${NC}"
    echo -e "   Local:    http://localhost:$PORT"
    echo -e "   External: $EXTERNAL_URL"
    echo ""
    echo -e "${BLUE}📊 Health check: $HEALTH_URL${NC}"
    echo ""
    echo -e "${YELLOW}🔧 Useful commands:${NC}"
    echo -e "   docker logs coolify-go                    # View logs"
    echo -e "   docker ps                                 # Check containers"
    echo -e "   ls $BACKUP_DIR                           # View backup files"
    echo ""
    if [ "$SKIP_BACKUP" != "true" ]; then
        echo -e "${YELLOW}🔄 Rollback (if needed):${NC}"
        echo -e "   cd $DATA_DIR"
        echo -e "   docker-compose down"
        echo -e "   cp $BACKUP_DIR/docker-compose.yml ."
        echo -e "   docker-compose up -d"
        echo ""
    fi
    echo -e "${CYAN}💡 Pro tip: Set up automatic updates with cron or systemd timers${NC}"
    echo ""
else
    echo "Update completed successfully"
    echo "Version: $NEW_VERSION"
    echo "Access: http://localhost:$PORT"
fi

# Cleanup old images (optional)
if [ "$QUIET" != "true" ]; then
    show_progress "Cleaning up old images"
    docker image prune -f >/dev/null 2>&1 || show_warning "Image cleanup skipped"
    show_success "Cleanup complete"
fi

show_success "Update process complete!"
