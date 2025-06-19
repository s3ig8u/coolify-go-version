#!/bin/bash
set -e

# Release script for Coolify Go
# Usage: ./scripts/release.sh [version]

VERSION=${1:-"$(git describe --tags --always)"}
REGISTRY="ghcr.io/coolify/coolify-go"

echo "🎉 Creating Coolify Go Release v$VERSION"
echo ""

# Build for all platforms
echo "🏗️  Building release artifacts..."
./scripts/build.sh $VERSION

# Push Docker image to registry
echo "🐳 Pushing Docker image to registry..."
docker tag "coolify-go:$VERSION" "$REGISTRY:$VERSION"
docker tag "coolify-go:$VERSION" "$REGISTRY:latest"
docker push "$REGISTRY:$VERSION"
docker push "$REGISTRY:latest"

# Create release notes
cat > "dist/RELEASE_NOTES_$VERSION.md" << EOF
# Coolify Go v$VERSION

## What's New
- Updated to version $VERSION
- Enhanced health monitoring with version info
- Improved deployment pipeline
- Cross-platform binary distribution

## Installation

### Docker (Recommended)
\`\`\`bash
# Pull and run the latest version
docker run -d \\
  --name coolify-go \\
  --restart unless-stopped \\
  -p 8080:8080 \\
  $REGISTRY:$VERSION

# Or use docker-compose
curl -o docker-compose.yml https://releases.coolify.io/v$VERSION/docker-compose.yml
docker-compose up -d
\`\`\`

### Binary Installation
\`\`\`bash
# Linux AMD64
curl -L -o coolify-go https://releases.coolify.io/v$VERSION/coolify-go-$VERSION-linux-amd64
chmod +x coolify-go
./coolify-go

# macOS ARM64 (Apple Silicon)
curl -L -o coolify-go https://releases.coolify.io/v$VERSION/coolify-go-$VERSION-darwin-arm64
chmod +x coolify-go
./coolify-go

# Windows
curl -L -o coolify-go.exe https://releases.coolify.io/v$VERSION/coolify-go-$VERSION-windows-amd64.exe
./coolify-go.exe
\`\`\`

## Update Instructions

### Docker Update
\`\`\`bash
# Stop current version
docker stop coolify-go
docker rm coolify-go

# Pull and run new version
docker run -d \\
  --name coolify-go \\
  --restart unless-stopped \\
  -p 8080:8080 \\
  $REGISTRY:$VERSION
\`\`\`

### Binary Update
\`\`\`bash
# Backup current installation
cp coolify-go coolify-go.backup

# Download new version
curl -L -o coolify-go https://releases.coolify.io/v$VERSION/coolify-go-$VERSION-linux-amd64
chmod +x coolify-go

# Restart service
./coolify-go
\`\`\`

## Verification
\`\`\`bash
# Check version
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","version":"$VERSION","buildTime":"...","commit":"..."}
\`\`\`

## Changelog
- Version bump to $VERSION
- Build time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
- Git commit: $(git rev-parse --short HEAD)
EOF

# Create checksums
echo "🔐 Creating checksums..."
cd dist
sha256sum * > "checksums_$VERSION.txt"
cd ..

# Create GitHub release (if gh CLI is available)
if command -v gh &> /dev/null; then
    echo "📦 Creating GitHub release..."
    gh release create "v$VERSION" \
        --title "Coolify Go v$VERSION" \
        --notes-file "dist/RELEASE_NOTES_$VERSION.md" \
        dist/*
fi

echo ""
echo "✅ Release v$VERSION created successfully!"
echo "📁 Artifacts available in: dist/"
echo "🐳 Docker image: $REGISTRY:$VERSION"
echo "📝 Release notes: dist/RELEASE_NOTES_$VERSION.md"
echo "🔐 Checksums: dist/checksums_$VERSION.txt"
