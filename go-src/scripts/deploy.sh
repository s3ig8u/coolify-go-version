#!/bin/bash
set -e

# Deployment script for Coolify Go
# Usage: ./scripts/deploy.sh [environment] [version]

ENVIRONMENT=${1:-"staging"}
VERSION=${2:-"latest"}
SERVICE_NAME="coolify-go"

echo "🚀 Deploying Coolify Go"
echo "Environment: $ENVIRONMENT"
echo "Version: $VERSION"
echo "Service: $SERVICE_NAME"
echo ""

# Check if we're deploying to production
if [ "$ENVIRONMENT" = "production" ]; then
    echo "⚠️  PRODUCTION DEPLOYMENT"
    echo "Are you sure you want to deploy to production? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "❌ Deployment cancelled"
        exit 1
    fi
fi

# Check if image exists locally, if not pull it
echo "🔍 Checking for Docker image..."
if ! docker image inspect "coolify-go:$VERSION" > /dev/null 2>&1; then
    echo "📥 Pulling Docker image..."
    docker pull "coolify-go:$VERSION"
else
    echo "✅ Using local Docker image coolify-go:$VERSION"
fi

# Stop old container if running
echo "🛑 Stopping old container..."
docker stop "$SERVICE_NAME" 2>/dev/null || true
docker rm "$SERVICE_NAME" 2>/dev/null || true

# Start new container
echo "🏃 Starting new container..."
docker run -d \
    --name "$SERVICE_NAME" \
    --restart unless-stopped \
    -p 8080:8080 \
    -e GO_ENV="$ENVIRONMENT" \
    "coolify-go:$VERSION"

# Wait for health check
echo "🔍 Waiting for health check..."
for i in {1..30}; do
    if curl -sf http://localhost:8080/health > /dev/null; then
        echo "✅ Health check passed!"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 2
done

# Verify deployment
echo "🔎 Verifying deployment..."
response=$(curl -s http://localhost:8080/health)
echo "Health check response: $response"

# Show container status
echo ""
echo "📊 Container status:"
docker ps | grep "$SERVICE_NAME"

echo ""
echo "✅ Deployment complete!"
echo "🌐 Service available at: http://localhost:8080"
