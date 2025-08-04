#!/bin/bash

# SpiceDB Test Script for AAA Service
set -e

echo "🧪 Testing SpiceDB setup..."

# Check if SpiceDB is running
if ! docker ps | grep -q "aaa-spicedb"; then
    echo "❌ SpiceDB container is not running. Start it first with:"
    echo "   ./scripts/setup_spicedb.sh"
    exit 1
fi

echo "✅ SpiceDB container is running"

# Test gRPC health check
echo "🔍 Testing gRPC health check..."
if command -v grpcurl > /dev/null 2>&1; then
    if grpcurl -plaintext -d '{}' localhost:50051 grpc.health.v1.Health/Check > /dev/null 2>&1; then
        echo "✅ gRPC health check passed"
    else
        echo "❌ gRPC health check failed"
        exit 1
    fi
else
    echo "⚠️  grpcurl not installed, skipping gRPC test"
fi

# Test schema loading
echo "📋 Testing schema loading..."
if docker exec aaa-spicedb spicedb schema read 2>&1 | grep -q "aaa/user\|aaa/role\|aaa/permission"; then
    echo "✅ Schema loaded successfully"
else
    echo "⚠️  Schema not loaded. Load it with:"
    echo "   ./scripts/load_schema.sh"
    echo "   or"
    echo "   zed schema write spicedb_schema.zed"
fi

# Test database connection
echo "🗄️  Testing database connection..."
if docker logs aaa-spicedb 2>&1 | grep -q "connected to postgres\|database connection established"; then
    echo "✅ Database connection established"
else
    echo "⚠️  Database connection status unclear, check logs:"
    echo "   docker logs aaa-spicedb"
fi

echo ""
echo "🎉 SpiceDB setup test completed!"
echo ""
echo "📊 Current status:"
echo "   - Container: $(docker ps --filter name=aaa-spicedb --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}')"
echo ""
echo "📋 Useful commands:"
echo "   - View logs: docker logs -f aaa-spicedb"
echo "   - Stop: docker stop aaa-spicedb"
echo "   - Restart: docker restart aaa-spicedb"
echo "   - Shell access: docker exec -it aaa-spicedb sh"
