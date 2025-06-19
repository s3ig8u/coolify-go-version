#!/bin/bash
set -e

# Build script for Coolify Go
# Usage: ./scripts/build.sh [version]

VERSION=${1:-"$(git describe --tags --always --dirty)"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🏗️  Building Coolify Go"
echo "Version: $VERSION"
echo "Build Time: $BUILD_TIME"
echo "Git Commit: $GIT_COMMIT"
echo ""

# Update version.go with build info
cat > version.go << EOF
package main

import "fmt"

const (
Version   = "$VERSION"
BuildTime = "$BUILD_TIME"
GitCommit = "$GIT_COMMIT"
)

func printVersion() {
fmt.Printf("Coolify Go v%s (built %s, commit %s)\n", Version, BuildTime, GitCommit)
}
EOF

# Build for multiple platforms
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

mkdir -p dist

for platform in $PLATFORMS; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    output_name="coolify-go-$VERSION-$GOOS-$GOARCH"
    
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi
    
    echo "📦 Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT" \
        -o "dist/$output_name" .
done

# Build Docker image
echo "🐳 Building Docker image..."
docker build \
    --build-arg VERSION="$VERSION" \
    --build-arg BUILD_TIME="$BUILD_TIME" \
    --build-arg GIT_COMMIT="$GIT_COMMIT" \
    -t "coolify-go:$VERSION" \
    -t "coolify-go:latest" \
    .

echo ""
echo "✅ Build complete!"
echo "📁 Binaries: $(ls -la dist/)"
echo "🐳 Docker: coolify-go:$VERSION"
