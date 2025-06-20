#!/bin/bash
# Test script for Coolify Go installation
# This simulates the installation process without requiring root

set -e

echo "🧪 Testing Coolify Go Installation Process"
echo ""

# Test version detection
echo "📋 Testing version detection..."
LATEST_VERSION=$(curl -s "https://api.github.com/repos/s3ig8u/coolify-go-version/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' 2>/dev/null || echo "v1.3.0")
echo "✅ Detected version: $LATEST_VERSION"

# Test OS detection
echo "📋 Testing OS detection..."
OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"' 2>/dev/null || echo "unknown")
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "⚠️  Unsupported architecture: $ARCH" ;;
esac

echo "✅ Platform: $OS_TYPE ($ARCH)"

# Test Docker availability
echo "📋 Testing Docker availability..."
if command -v docker >/dev/null 2>&1; then
    echo "✅ Docker is available"
    docker --version
else
    echo "⚠️  Docker not found"
fi

# Test Docker Compose availability
echo "📋 Testing Docker Compose availability..."
if command -v docker-compose >/dev/null 2>&1; then
    echo "✅ docker-compose is available"
    docker-compose --version
elif docker compose version >/dev/null 2>&1; then
    echo "✅ docker compose is available"
    docker compose version
else
    echo "⚠️  Docker Compose not found"
fi

# Test Docker build
echo "📋 Testing Docker build..."
if docker build -t coolify-go-test:latest . >/dev/null 2>&1; then
    echo "✅ Docker build successful"
    docker images | grep coolify-go-test || echo "⚠️  Image not found"
else
    echo "❌ Docker build failed"
fi

# Test running container
echo "📋 Testing container run..."
if docker run --rm -d --name coolify-go-test-run -p 8082:8080 coolify-go-test:latest >/dev/null 2>&1; then
    echo "✅ Container started successfully"
    sleep 3
    
    # Test health endpoint
    if curl -sf http://localhost:8082/health >/dev/null 2>&1; then
        echo "✅ Health endpoint working"
        echo "📊 Health response:"
        curl -s http://localhost:8082/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8082/health
    else
        echo "❌ Health endpoint failed"
    fi
    
    # Clean up
    docker stop coolify-go-test-run >/dev/null 2>&1 || true
else
    echo "❌ Container failed to start"
fi

# Clean up test image
docker rmi coolify-go-test:latest >/dev/null 2>&1 || true

echo ""
echo "🎉 Installation test completed!"
echo ""
echo "📚 Next steps:"
echo "  1. Run the installation script as root:"
echo "     sudo bash install.sh"
echo "  2. Visit http://localhost:8080 after installation"
echo "  3. Check logs with: docker logs coolify-go"
