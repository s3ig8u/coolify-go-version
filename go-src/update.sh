#!/bin/bash
set -e

# Coolify Go Update Script
REGISTRY="shrtso.azurecr.io/coolify-go"
COMPOSE_FILE="/data/coolify-go/docker-compose.yml"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔄 Coolify Go Update${NC}"

# Check if Coolify Go is installed
if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}❌ Coolify Go not found. Please install first:${NC}"
    echo -e "${YELLOW}curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash${NC}"
    exit 1
fi

# Check current version
echo -e "${BLUE}📊 Checking current version...${NC}"
if docker ps --format "table {{.Names}}\t{{.Image}}" | grep -q coolify-go; then
    CURRENT_IMAGE=$(docker inspect coolify-go --format='{{.Config.Image}}' 2>/dev/null || echo "unknown")
    echo -e "${BLUE}Current image: $CURRENT_IMAGE${NC}"
    
    # Try to get version from running container
    CURRENT_VERSION=$(docker exec coolify-go ./coolify-go --version 2>/dev/null | head -1 || echo "unknown")
    echo -e "${BLUE}Current version: $CURRENT_VERSION${NC}"
else
    echo -e "${YELLOW}⚠️  Coolify Go container not running${NC}"
fi

# Pull latest image
echo -e "${BLUE}📦 Pulling latest image from Azure Container Registry...${NC}"
if docker pull "$REGISTRY:latest" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Successfully pulled latest image${NC}"
else
    echo -e "${RED}❌ Failed to pull latest image from registry${NC}"
    exit 1
fi

# Backup current data
echo -e "${BLUE}💾 Creating backup...${NC}"
BACKUP_DIR="/data/coolify-go/backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup database
if docker ps --format "{{.Names}}" | grep -q coolify-go-db; then
    echo -e "${BLUE}🗄️  Backing up database...${NC}"
    docker exec coolify-go-db pg_dump -U coolify_go coolify_go > "$BACKUP_DIR/database.sql" 2>/dev/null || echo -e "${YELLOW}⚠️  Database backup failed${NC}"
fi

# Backup environment and compose files
cp /data/coolify-go/.env "$BACKUP_DIR/" 2>/dev/null || echo -e "${YELLOW}⚠️  .env backup failed${NC}"
cp /data/coolify-go/docker-compose.yml "$BACKUP_DIR/" 2>/dev/null || echo -e "${YELLOW}⚠️  docker-compose.yml backup failed${NC}"

echo -e "${GREEN}✅ Backup created: $BACKUP_DIR${NC}"

# Stop services
echo -e "${BLUE}🛑 Stopping services...${NC}"
cd /data/coolify-go
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose down >/dev/null 2>&1
else
    docker compose down >/dev/null 2>&1
fi

# Update docker-compose.yml with latest image
echo -e "${BLUE}📝 Updating configuration...${NC}"
sed -i "s|image:.*coolify-go.*|image: $REGISTRY:latest|g" docker-compose.yml

# Start services with new image
echo -e "${BLUE}🚀 Starting updated services...${NC}"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d >/dev/null 2>&1
else
    docker compose up -d >/dev/null 2>&1
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

# Check new version
echo ""
echo -e "${BLUE}📊 Checking updated version...${NC}"
NEW_VERSION=$(docker exec coolify-go ./coolify-go --version 2>/dev/null | head -1 || echo "unknown")
echo -e "${GREEN}New version: $NEW_VERSION${NC}"

# Show status
echo -e "${BLUE}📊 Service Status:${NC}"
curl -s http://localhost:8080/health 2>/dev/null | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health

echo ""
echo -e "${GREEN}🎉 Update completed successfully!${NC}"
echo -e "${BLUE}💾 Backup location: $BACKUP_DIR${NC}"
echo -e "${BLUE}🌐 Access: http://localhost:8080${NC}"
echo -e "${BLUE}📊 Health: http://localhost:8080/health${NC}"

echo ""
echo -e "${YELLOW}📚 Post-update commands:${NC}"
echo -e "   docker logs coolify-go         # View logs"
echo -e "   docker ps                      # Check containers"
echo -e "   ls $BACKUP_DIR                 # View backup files"

# Cleanup old images (optional)
echo ""
echo -e "${BLUE}🧹 Cleaning up old images...${NC}"
docker image prune -f >/dev/null 2>&1 || echo -e "${YELLOW}⚠️  Image cleanup skipped${NC}"

echo -e "${GREEN}✅ Update process complete!${NC}"
