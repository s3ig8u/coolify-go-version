#!/bin/bash
set -e

# Demo Coolify Go Installation Script
# Usage: ./demo-install.sh

REGISTRY="coolify-go"
VERSION="v1.2.0"

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

echo -e "${BLUE}📋 Detected platform: $OS/$ARCH${NC}"
echo -e "${GREEN}✅ Latest version: $VERSION${NC}"

# Installation method selection
echo ""
echo -e "${BLUE}📦 Choose installation method:${NC}"
echo "1) Docker (Recommended)"
echo "2) Binary installation (demo)"
echo ""
read -p "Enter your choice (1-2): " INSTALL_METHOD

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
        
        echo -e "${BLUE}🏃 Starting Coolify Go...${NC}"
        docker run -d \
            --name coolify-go \
            --restart unless-stopped \
            -p 8080:8080 \
            "$REGISTRY:$VERSION"
        
        echo -e "${GREEN}✅ Coolify Go started successfully!${NC}"
        echo -e "${BLUE}🌐 Access it at: http://localhost:8080${NC}"
        ;;
        
    2)
        echo -e "${BLUE}📦 Installing binary (demo)...${NC}"
        
        # Check if we have the binary built
        if [ -f "dist/coolify-go-$VERSION-$OS-$ARCH" ]; then
            cp "dist/coolify-go-$VERSION-$OS-$ARCH" "./coolify-go"
            chmod +x "./coolify-go"
            echo -e "${GREEN}✅ Binary installed to current directory${NC}"
            echo -e "${BLUE}🚀 Run: ./coolify-go${NC}"
        else
            echo -e "${YELLOW}⚠️  Binary not found. Building from source...${NC}"
            go build -o coolify-go .
            echo -e "${GREEN}✅ Built and installed coolify-go${NC}"
            echo -e "${BLUE}🚀 Run: ./coolify-go${NC}"
        fi
        ;;
        
    *)
        echo -e "${RED}❌ Invalid choice${NC}"
        exit 1
        ;;
esac

# Wait for service to be ready
if [ "$INSTALL_METHOD" = "1" ]; then
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
    curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8080/health
fi

echo ""
echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
echo -e "${BLUE}📚 Documentation: https://docs.coolify.io${NC}"
echo -e "${BLUE}🐛 Issues: https://github.com/coolify/coolify-go/issues${NC}"
echo -e "${BLUE}💬 Community: https://discord.coolify.io${NC}"
