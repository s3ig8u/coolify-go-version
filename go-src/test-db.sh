#!/bin/bash

echo "🔍 Testing Database Connection"

# Check if containers are running
echo "📊 Container Status:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "🗄️  Database Connection Test:"

# Test database connection
if docker exec coolify-go-db psql -U coolify_go -d coolify_go -c "SELECT version();" >/dev/null 2>&1; then
    echo "✅ Database is accessible from inside container"
else
    echo "❌ Database connection failed"
    echo "   Checking database logs:"
    docker logs coolify-go-db --tail 10
fi

echo ""
echo "🌐 Network Test:"
# Test if coolify-go can reach postgres
if docker exec coolify-go ping -c 1 postgres >/dev/null 2>&1; then
    echo "✅ Network connection to postgres container works"
else
    echo "❌ Network connection to postgres failed"
fi

# Check environment variables
echo ""
echo "🔧 Environment Variables:"
docker exec coolify-go env | grep -E "DB_|POSTGRES_" | sort

echo ""
echo "🔄 Restart Database Container:"
echo "docker restart coolify-go-db"

echo ""
echo "🔄 Restart All Services:"
echo "cd /data/coolify-go && docker-compose restart"
