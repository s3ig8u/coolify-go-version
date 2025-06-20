# 🎬 Coolify Go Installation Demo

## For VPS Installation

### Step 1: SSH to Your VPS
```bash
ssh root@your-vps-ip
```

### Step 2: Run One Command
```bash
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | bash
```

### Expected Output:
```
🚀 Coolify Go Installation
📋 Platform: ubuntu (amd64)
⚠️  GitHub API unavailable, using fallback version: v1.3.0
📁 Creating directories...
📦 Installing required packages...
🔧 Configuring Docker...
⚙️  Creating configuration...
🚀 Deploying Coolify Go...
📦 Trying to pull from registry: shrtso.azurecr.io/coolify-go:v1.3.0
⚠️  Registry image not available, building from source...
✅ Source code cloned successfully
🔨 Building Docker image locally...
✅ Local build successful
🚀 Starting services...
🔍 Waiting for service to be ready...
✅ Service is ready!

📊 Service Status:
{
    "status": "healthy",
    "version": "v1.3.0",
    "buildTime": "2025-06-20T10:11:00Z",
    "commit": "local-build",
    "timestamp": "2025-06-20T09:13:00Z",
    "database": "connected"
}

🎉 Installation completed successfully!
🌐 Access your application at:
   Local:    http://localhost:8080
   External: http://123.456.789.012:8080

📊 Health check: http://123.456.789.012:8080/health
📁 Data directory: /data/coolify-go
⚙️  Configuration: /data/coolify-go/.env

📚 For troubleshooting:
   docker logs coolify-go
   docker ps
   docker-compose logs
```

### Step 3: Verify Installation
```bash
# Check running containers
docker ps

# Test health endpoint
curl http://localhost:8080/health

# View application in browser
# Navigate to: http://your-vps-ip:8080
```

## How the Installation Works

### 1. Registry-First Approach
- ✅ **Tries to pull from registry**: `shrtso.azurecr.io/coolify-go:v1.3.0`
- ✅ **Fast deployment** if image exists in registry
- ✅ **No build time** required

### 2. Source-Build Fallback
- ✅ **Clones from GitHub**: `https://github.com/s3ig8u/coolify-go-version.git`
- ✅ **Builds locally** using Docker multi-stage build
- ✅ **Always works** even without registry access
- ✅ **Clean source separation** - no code embedded in install script

### 3. Full Stack Deployment
- ✅ **Coolify Go** application (Port 8080)
- ✅ **PostgreSQL 15** database (Port 5432)
- ✅ **Redis 7** cache (Port 6379)
- ✅ **Docker networking** with proper dependencies
- ✅ **Persistent volumes** for data storage

## What Gets Installed

### Services
```bash
CONTAINER ID   IMAGE                    COMMAND                  STATUS
abc123def456   coolify-go:v1.3.0       "./coolify-go"           Up 2 minutes
def456abc789   postgres:15              "docker-entrypoint.s…"   Up 2 minutes
789abc123def   redis:7-alpine           "docker-entrypoint.s…"   Up 2 minutes
```

### Directory Structure
```
/data/coolify-go/
├── .env                    # Environment configuration
├── docker-compose.yml     # Service orchestration
├── source/                 # Application data
├── ssh/                    # SSH keys and configuration
├── applications/           # Future: deployed applications
└── databases/             # Future: database instances
```

### Network Access
- **Internal**: Services communicate via Docker network
- **External**: Application accessible on port 8080
- **Database**: PostgreSQL accessible on port 5432 (for admin)
- **Cache**: Redis accessible on port 6379 (for debugging)

## Troubleshooting Common Issues

### 1. Port Already in Use
```bash
# Check what's using port 8080
sudo netstat -tulpn | grep 8080

# Stop conflicting service or change port in docker-compose.yml
```

### 2. Docker Not Starting
```bash
# Check Docker status
sudo systemctl status docker

# Restart Docker
sudo systemctl restart docker
```

### 3. Build Failures
```bash
# Check Docker logs
docker logs coolify-go

# Rebuild manually
cd /data/coolify-go
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### 4. Database Connection Issues
```bash
# Check database logs
docker logs coolify-go-db

# Verify database is accessible
docker exec -it coolify-go-db psql -U coolify_go -d coolify_go -c "\l"
```

## Next Steps After Installation

### 1. Security Hardening
```bash
# Change default passwords
sudo nano /data/coolify-go/.env

# Set up firewall
sudo ufw allow 8080
sudo ufw enable
```

### 2. Domain Setup (Optional)
```bash
# Install Nginx
sudo apt install nginx

# Configure reverse proxy for your domain
# See VPS_DEPLOYMENT_GUIDE.md for details
```

### 3. Monitoring
```bash
# Set up log rotation
# Monitor resource usage
# Configure backups
```

## Development Workflow

### Local Testing
```bash
# Test locally first
cd go-src
make test-install

# Build and test manually
make build
make quick-test
```

### VPS Deployment
```bash
# Deploy to VPS
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash

# Monitor deployment
docker logs coolify-go -f
```

### Updates
```bash
# Pull latest version
docker-compose pull
docker-compose up -d

# Or re-run install script for major updates
```

---

This installation approach provides:
- ✅ **Reliability**: Registry-first with source fallback
- ✅ **Speed**: Fast deployment when registry available  
- ✅ **Flexibility**: Always buildable from source
- ✅ **Cleanliness**: No embedded code in install script
- ✅ **Production-ready**: Full stack with monitoring
