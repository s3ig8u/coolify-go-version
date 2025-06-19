#!/bin/bash
set -e

# Coolify Go Port - Update Script
# Usage: curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify/v4.x/go-src/update.sh | bash

REPO="s3ig8u/coolify"
REGISTRY="ghcr.io/s3ig8u/coolify-go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Coolify Go Update Script${NC}"
echo ""

# Check current version
echo -e "${BLUE}🔍 Checking current version...${NC}"
CURRENT_VERSION=""
if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    CURRENT_VERSION=$(curl -s http://localhost:8080/health | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✅ Current version: $CURRENT_VERSION${NC}"
else
    echo -e "${YELLOW}⚠️  No running instance detected${NC}"
fi

# Get latest version
echo -e "${BLUE}🔍 Checking for updates...${NC}"
if command -v curl >/dev/null 2>&1; then
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
    LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
else
    echo -e "${RED}❌ Neither curl nor wget found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Latest version: $LATEST_VERSION${NC}"

# Check if update is needed
if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
    echo -e "${GREEN}✅ You're already running the latest version!${NC}"
    exit 0
fi

if [ -n "$CURRENT_VERSION" ]; then
    echo -e "${YELLOW}📦 Update available: $CURRENT_VERSION → $LATEST_VERSION${NC}"
else
    echo -e "${YELLOW}📦 Installing latest version: $LATEST_VERSION${NC}"
fi

# Detect installation method
INSTALLATION_METHOD=""
if docker ps --filter "name=coolify-go" --format "table {{.Names}}" | grep -q coolify-go; then
    INSTALLATION_METHOD="docker"
elif docker-compose ps coolify-go >/dev/null 2>&1 || docker compose ps coolify-go >/dev/null 2>&1; then
    INSTALLATION_METHOD="docker-compose"
elif command -v coolify-go >/dev/null 2>&1; then
    INSTALLATION_METHOD="binary"
else
    echo -e "${YELLOW}⚠️  Could not detect installation method${NC}"
    echo -e "${BLUE}🔧 Please choose update method:${NC}"
    echo "1) Docker"
    echo "2) Binary"
    echo "3) Docker Compose"
    read -p "Enter your choice (1-3): " choice
    case $choice in
        1) INSTALLATION_METHOD="docker" ;;
        2) INSTALLATION_METHOD="binary" ;;
        3) INSTALLATION_METHOD="docker-compose" ;;
        *) echo -e "${RED}❌ Invalid choice${NC}"; exit 1 ;;
    esac
fi

echo -e "${BLUE}🔧 Detected installation method: $INSTALLATION_METHOD${NC}"

# Backup current installation
echo -e "${BLUE}💾 Creating backup...${NC}"
case $INSTALLATION_METHOD in
    docker)
        # Export current container configuration
        if docker ps --filter "name=coolify-go" --format "table {{.Names}}" | grep -q coolify-go; then
            docker inspect coolify-go > coolify-go-backup-$(date +%Y%m%d-%H%M%S).json
            echo -e "${GREEN}✅ Container configuration backed up${NC}"
        fi
        ;;
    binary)
        if command -v coolify-go >/dev/null 2>&1; then
            cp "$(which coolify-go)" "coolify-go-backup-$(date +%Y%m%d-%H%M%S)"
            echo -e "${GREEN}✅ Binary backed up${NC}"
        fi
        ;;
    docker-compose)
        if [ -f docker-compose.yml ]; then
            cp docker-compose.yml "docker-compose-backup-$(date +%Y%m%d-%H%M%S).yml"
            echo -e "${GREEN}✅ Docker Compose configuration backed up${NC}"
        fi
        ;;
esac

# Perform update
echo -e "${BLUE}🚀 Starting update...${NC}"
case $INSTALLATION_METHOD in
    docker)
        echo -e "${BLUE}🐳 Updating Docker container...${NC}"
        
        # Stop current container
        docker stop coolify-go 2>/dev/null || true
        docker rm coolify-go 2>/dev/null || true
        
        # Pull new image
        docker pull "$REGISTRY:$LATEST_VERSION"
        
        # Start new container with same configuration
        docker run -d \
            --name coolify-go \
            --restart unless-stopped \
            -p 8080:8080 \
            "$REGISTRY:$LATEST_VERSION"
        ;;
        
    binary)
        echo -e "${BLUE}📦 Updating binary...${NC}"
        
        # Detect OS and architecture
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            arm64|aarch64) ARCH="arm64" ;;
        esac
        
        # Download new binary
        BINARY_FILE="coolify-go-$LATEST_VERSION-$OS-$ARCH"
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$BINARY_FILE"
        
        TMP_DIR=$(mktemp -d)
        cd "$TMP_DIR"
        
        if command -v curl >/dev/null 2>&1; then
            curl -L -o coolify-go "$DOWNLOAD_URL"
        elif command -v wget >/dev/null 2>&1; then
            wget -O coolify-go "$DOWNLOAD_URL"
        fi
        
        chmod +x coolify-go
        
        # Replace existing binary
        if [ -w "$(which coolify-go)" ]; then
            mv coolify-go "$(which coolify-go)"
        else
            sudo mv coolify-go "$(which coolify-go)"
        fi
        
        cd - >/dev/null
        rm -rf "$TMP_DIR"
        ;;
        
    docker-compose)
        echo -e "${BLUE}🐳 Updating Docker Compose services...${NC}"
        
        # Update image tag in docker-compose.yml
        if [ -f docker-compose.yml ]; then
            sed -i.bak "s|image: $REGISTRY:.*|image: $REGISTRY:$LATEST_VERSION|g" docker-compose.yml
        fi
        
        # Pull new images and restart services
        if command -v docker-compose >/dev/null 2>&1; then
            docker-compose pull
            docker-compose up -d
        else
            docker compose pull
            docker compose up -d
        fi
        ;;
esac

# Wait for service to be ready
echo -e "${BLUE}🔍 Waiting for service to start...${NC}"
for i in {1..30}; do
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is ready!${NC}"
        break
    fi
    echo -n "."
    sleep 2
done

# Verify update
echo ""
echo -e "${BLUE}🔎 Verifying update...${NC}"
NEW_VERSION=$(curl -s http://localhost:8080/health | grep -o '"version":"[^"]*"' | cut -d'"' -f4)

if [ "$NEW_VERSION" = "$LATEST_VERSION" ]; then
    echo -e "${GREEN}✅ Update successful!${NC}"
    echo -e "${GREEN}🎉 Updated from $CURRENT_VERSION to $NEW_VERSION${NC}"
else
    echo -e "${RED}❌ Update failed - version mismatch${NC}"
    echo -e "${YELLOW}Expected: $LATEST_VERSION, Got: $NEW_VERSION${NC}"
    exit 1
fi

# Show new status
echo ""
echo -e "${BLUE}📊 New Service Status:${NC}"
curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health

echo ""
echo -e "${GREEN}🎉 Update completed successfully!${NC}"
echo -e "${BLUE}🌐 Service is available at: http://localhost:8080${NC}"

# Show changelog if available
echo ""
echo -e "${BLUE}📝 What's new in $LATEST_VERSION:${NC}"
if command -v curl >/dev/null 2>&1; then
    curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"body":' | sed 's/.*"body": *"\([^"]*\)".*/\1/' | sed 's/\\n/\n/g' | head -10
fi

echo -e "${BLUE}📚 Full changelog: https://github.com/$REPO/releases/tag/$LATEST_VERSION${NC}"
