# Coolify Go - Customer Distribution & Updates

## 🎯 How Customers Acquire & Update Coolify Go

### 🚀 **Initial Installation**

#### **Method 1: One-Line Install (Recommended)**
```bash
# Install latest version automatically
curl -fsSL https://install.coolify.io | bash
```

#### **Method 2: Docker (Most Popular)**
```bash
# Pull and run latest version
docker run -d \
  --name coolify-go \
  --restart unless-stopped \
  -p 8080:8080 \
  ghcr.io/coolify/coolify-go:latest

# Verify installation
curl http://localhost:8080/health
```

#### **Method 3: Docker Compose (Production)**
```bash
# Download compose file
curl -o docker-compose.yml https://releases.coolify.io/latest/docker-compose.yml

# Start all services
docker-compose up -d
```

#### **Method 4: Binary Download**
```bash
# Linux AMD64
curl -L -o coolify-go https://releases.coolify.io/latest/coolify-go-linux-amd64
chmod +x coolify-go
./coolify-go

# macOS Apple Silicon
curl -L -o coolify-go https://releases.coolify.io/latest/coolify-go-darwin-arm64
chmod +x coolify-go
./coolify-go

# Windows
curl -L -o coolify-go.exe https://releases.coolify.io/latest/coolify-go-windows-amd64.exe
./coolify-go.exe
```

---

### 🔄 **Automatic Updates**

#### **Update Check Command**
```bash
# Check for updates
curl -fsSL https://update.coolify.io | bash
```

#### **How Update Detection Works**
1. **Current Version Check**: Queries `/health` endpoint
2. **Latest Version Check**: GitHub Releases API
3. **Comparison**: Determines if update needed
4. **Method Detection**: Automatically detects installation method
5. **Backup**: Creates backup before updating
6. **Update**: Downloads and deploys new version
7. **Verification**: Confirms successful update

---

### 📦 **Distribution Channels**

#### **1. GitHub Releases** (Primary)
- **URL**: `https://github.com/coolify/coolify-go/releases`
- **Artifacts**: Cross-platform binaries, Docker images, checksums
- **Format**: `coolify-go-v1.1.0-{os}-{arch}`
- **Platforms**: Linux, macOS, Windows (AMD64, ARM64)

#### **2. Container Registry** (Docker)
- **Registry**: `ghcr.io/coolify/coolify-go`
- **Tags**: `latest`, `v1.1.0`, `stable`
- **Multi-arch**: Supports AMD64 and ARM64
- **Automated**: Built via CI/CD pipeline

#### **3. Package Managers** (Future)
```bash
# Homebrew (macOS)
brew install coolify/tap/coolify-go

# Chocolatey (Windows)
choco install coolify-go

# APT (Ubuntu/Debian)
apt install coolify-go

# Snap (Universal Linux)
snap install coolify-go
```

---

### 🔧 **Update Mechanisms**

#### **Docker Updates**
```bash
# Method 1: Pull and restart
docker stop coolify-go
docker rm coolify-go
docker pull ghcr.io/coolify/coolify-go:latest
docker run -d --name coolify-go -p 8080:8080 ghcr.io/coolify/coolify-go:latest

# Method 2: Watchtower (Auto-updates)
docker run -d \
  --name watchtower \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --schedule "0 0 4 * * *" \
  coolify-go
```

#### **Binary Updates**
```bash
# Backup current
cp coolify-go coolify-go.backup

# Download new version
curl -L -o coolify-go https://releases.coolify.io/latest/coolify-go-linux-amd64
chmod +x coolify-go

# Restart service
./coolify-go
```

#### **Docker Compose Updates**
```bash
# Update compose file
curl -o docker-compose.yml https://releases.coolify.io/latest/docker-compose.yml

# Pull new images and restart
docker-compose pull
docker-compose up -d
```

---

### 📊 **Version Management**

#### **Semantic Versioning**
- **Format**: `v{MAJOR}.{MINOR}.{PATCH}`
- **Example**: `v1.2.3`
- **Breaking Changes**: Major version bump
- **Features**: Minor version bump
- **Bug Fixes**: Patch version bump

#### **Release Channels**
- **`latest`**: Stable releases (recommended)
- **`beta`**: Pre-release testing
- **`nightly`**: Daily development builds
- **`v1.1.0`**: Specific version pinning

#### **Health Check & Version API**
```bash
# Check running version
curl http://localhost:8080/health

# Response example
{
  "status": "healthy",
  "version": "v1.1.0",
  "buildTime": "2025-06-19T19:21:25Z",
  "commit": "abc123"
}
```

---

### 🛡️ **Security & Verification**

#### **Checksum Verification**
```bash
# Download checksums
curl -L -o checksums.txt https://releases.coolify.io/v1.1.0/checksums.txt

# Verify binary
sha256sum -c checksums.txt
```

#### **GPG Signature Verification**
```bash
# Download signature
curl -L -o coolify-go.sig https://releases.coolify.io/v1.1.0/coolify-go-linux-amd64.sig

# Verify signature
gpg --verify coolify-go.sig coolify-go
```

#### **Docker Image Verification**
```bash
# Verify image signature (cosign)
cosign verify ghcr.io/coolify/coolify-go:v1.1.0

# Check for vulnerabilities
docker scout quickview ghcr.io/coolify/coolify-go:v1.1.0
```

---

### 🔄 **Migration & Rollback**

#### **Data Migration**
```bash
# Backup data before update
docker exec coolify-go backup-data

# Migrate to new version
coolify-go migrate --from=v1.0.0 --to=v1.1.0
```

#### **Rollback Procedure**
```bash
# Docker rollback
docker stop coolify-go
docker rm coolify-go
docker run -d --name coolify-go -p 8080:8080 ghcr.io/coolify/coolify-go:v1.0.0

# Binary rollback
cp coolify-go.backup coolify-go
./coolify-go
```

---

### 📈 **Update Notifications**

#### **Update Channels**
- **In-App Notifications**: Dashboard banner when update available
- **Email Notifications**: Optional email alerts for new releases
- **Webhook Notifications**: POST to customer webhook URL
- **RSS Feed**: Subscribe to release announcements

#### **Automatic Update Configuration**
```yaml
# coolify-config.yml
updates:
  enabled: true
  channel: "stable"  # latest, beta, nightly
  schedule: "daily"  # daily, weekly, manual
  backup: true
  rollback_on_failure: true
  notification:
    email: admin@company.com
    webhook: https://webhook.company.com/coolify-updates
```

---

### 🏗️ **Enterprise Distribution**

#### **Private Registry**
```bash
# Customer's private registry
docker pull registry.company.com/coolify/coolify-go:v1.1.0
```

#### **Air-Gapped Environments**
```bash
# Download offline bundle
curl -L -o coolify-go-offline.tar.gz https://releases.coolify.io/v1.1.0/offline-bundle.tar.gz

# Transfer and install
tar -xzf coolify-go-offline.tar.gz
./install-offline.sh
```

#### **License Management**
```bash
# Enterprise license activation
coolify-go license activate --key=ENTERPRISE-KEY-123

# License verification
coolify-go license status
```

---

## 📋 **Customer Journey Summary**

1. **Discovery**: Customer finds Coolify Go via website/docs
2. **Installation**: One-line install script or Docker
3. **Configuration**: Setup via web UI or config files
4. **Operation**: Daily usage for application deployment
5. **Updates**: Automatic notifications and easy updates
6. **Support**: Community forum, documentation, enterprise support

This distribution model ensures customers can easily acquire, install, update, and maintain Coolify Go across all deployment scenarios!
