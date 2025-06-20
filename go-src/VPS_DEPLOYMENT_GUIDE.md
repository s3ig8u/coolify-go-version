# 🚀 Coolify Go VPS Deployment Guide

## Option 1: Direct Installation (Recommended)

### Prerequisites
- VPS with Ubuntu 20.04+ / Debian 11+ / CentOS 8+ / Fedora 35+
- Root access (sudo)
- At least 2GB RAM, 20GB storage
- Public IP address

### One-Command Installation
```bash
# SSH into your VPS
ssh root@your-vps-ip

# Run the installation script
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | bash
```

### What This Does:
1. ✅ Detects your OS and architecture
2. ✅ Installs Docker and required packages
3. ✅ Configures Docker daemon properly
4. ✅ Creates `/data/coolify-go/` directory structure
5. ✅ Generates secure random passwords
6. ✅ **Tries to pull from registry first**
7. ✅ **Falls back to building from source if needed**
8. ✅ Deploys full stack: App + PostgreSQL + Redis
9. ✅ Sets up health monitoring and dependency management
10. ✅ Provides external IP access information

### After Installation:
```bash
# Check status
docker ps

# View logs
docker logs coolify-go

# Access the application
curl http://your-vps-ip:8080/health
```

## Option 2: Manual GitHub Repository Installation

### Step 1: Clone Repository
```bash
# SSH into your VPS
ssh root@your-vps-ip

# Clone the repository
git clone https://github.com/s3ig8u/coolify-go-version.git
cd coolify-go-version/go-src
```

### Step 2: Build and Deploy
```bash
# Make scripts executable
chmod +x install.sh

# Run installation
./install.sh
```

## Option 3: Docker-Only Deployment

### Quick Docker Run
```bash
# SSH into your VPS
ssh root@your-vps-ip

# Pull and run (without database)
docker run -d \
  --name coolify-go \
  --restart unless-stopped \
  -p 8080:8080 \
  shrtso.azurecr.io/coolify-go:v1.3.0
```

### With Full Stack (Docker Compose)
```bash
# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  coolify-go:
    image: shrtso.azurecr.io/coolify-go:v1.3.0
    container_name: coolify-go
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PASSWORD=your_secure_password_here
      - REDIS_PASSWORD=your_redis_password_here
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    container_name: coolify-go-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: coolify_go
      POSTGRES_USER: coolify_go
      POSTGRES_PASSWORD: your_secure_password_here
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    container_name: coolify-go-redis
    restart: unless-stopped
    command: redis-server --requirepass your_redis_password_here
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

volumes:
  postgres_data:
  redis_data:
EOF

# Start services
docker-compose up -d
```

## 🔧 Post-Installation Configuration

### 1. Check Services
```bash
# View running containers
docker ps

# Check logs
docker logs coolify-go
docker logs coolify-go-db
docker logs coolify-go-redis

# Test health endpoint
curl http://localhost:8080/health
```

### 2. Access from Internet
```bash
# Test from external
curl http://your-vps-ip:8080/health

# Open browser to
http://your-vps-ip:8080
```

### 3. Configure Firewall (if needed)
```bash
# Ubuntu/Debian
ufw allow 8080

# CentOS/RHEL/Fedora
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --reload
```

## 🔒 Security Recommendations

### 1. Change Default Passwords
```bash
# Edit environment file
nano /data/coolify-go/.env

# Update these values:
DB_PASSWORD=your_new_secure_password
REDIS_PASSWORD=your_new_redis_password
JWT_SECRET=your_new_jwt_secret
```

### 2. Set Up SSL/TLS (Optional)
```bash
# Install Nginx reverse proxy
apt update && apt install nginx certbot python3-certbot-nginx

# Configure Nginx
cat > /etc/nginx/sites-available/coolify-go << 'EOF'
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# Enable site
ln -s /etc/nginx/sites-available/coolify-go /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx

# Get SSL certificate
certbot --nginx -d your-domain.com
```

## 🛠️ Troubleshooting

### Check Service Status
```bash
# Container status
docker ps -a

# Resource usage
docker stats

# Logs for debugging
docker logs coolify-go --tail 100 -f
```

### Common Issues

#### Port 8080 Already in Use
```bash
# Find what's using the port
netstat -tulpn | grep 8080

# Change port in docker-compose.yml
# Change "8080:8080" to "8081:8080" then restart
docker-compose down && docker-compose up -d
```

#### Database Connection Issues
```bash
# Check postgres logs
docker logs coolify-go-db

# Verify database is running
docker exec -it coolify-go-db psql -U coolify_go -d coolify_go -c "\dt"
```

#### Container Won't Start
```bash
# Check detailed logs
docker logs coolify-go

# Restart services
docker-compose restart

# Rebuild if needed
docker-compose down
docker-compose pull
docker-compose up -d
```

## 📊 Monitoring & Maintenance

### View Application Status
```bash
# Health check
curl http://localhost:8080/health | jq

# Container stats
docker stats --no-stream

# Disk usage
docker system df
```

### Backup Database
```bash
# Create backup
docker exec coolify-go-db pg_dump -U coolify_go coolify_go > backup-$(date +%Y%m%d).sql

# Restore backup
cat backup-20250620.sql | docker exec -i coolify-go-db psql -U coolify_go -d coolify_go
```

### Update to Latest Version
```bash
# Pull latest images
docker-compose pull

# Restart with new version
docker-compose down && docker-compose up -d

# Or use the update script (if available)
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/update.sh | bash
```

## 🎯 Quick Start Summary

For the fastest deployment on your VPS:

```bash
# 1. SSH to your VPS
ssh root@your-vps-ip

# 2. Run one command
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | bash

# 3. Wait 2-3 minutes for installation

# 4. Visit your application
http://your-vps-ip:8080
```

That's it! Your Coolify Go instance will be running with PostgreSQL and Redis, ready for development.

## 📞 Support

- **Issues**: Report at GitHub repository
- **Logs**: Always check `docker logs coolify-go` first
- **Health**: Monitor `/health` endpoint
- **Updates**: Re-run install script for updates

---

**Note**: This is a development version. For production use, implement proper SSL, firewalls, and monitoring.
