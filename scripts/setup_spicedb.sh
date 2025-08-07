#!/bin/bash

# SpiceDB Setup Script for AAA Service
set -e

echo "🚀 Setting up SpiceDB for AAA Service..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if schema file exists
if [ ! -f "./spicedb_schema.zed" ]; then
    echo "❌ Schema file not found: ./spicedb_schema.zed"
    exit 1
fi

# Set default environment variables
export DB_POSTGRES_USER=${DB_POSTGRES_USER:-aaa_user}
export DB_POSTGRES_PASSWORD=${DB_POSTGRES_PASSWORD:-aaa_password}
export DB_POSTGRES_HOST=${DB_POSTGRES_HOST:-host.docker.internal}
export DB_POSTGRES_PORT=${DB_POSTGRES_PORT:-5432}
export SPICEDB_DB_NAME=${SPICEDB_DB_NAME:-spicedb}
export DB_SSL_MODE=${DB_SSL_MODE:-disable}

# Generate a random secret key if not provided
if [ -z "$SPICEDB_SECRET_KEY" ]; then
    export SPICEDB_SECRET_KEY=$(openssl rand -hex 32)
    echo "🔑 Generated SpiceDB secret key: $SPICEDB_SECRET_KEY"
fi

# Update docker-compose file with the secret key
sed -i.bak "s/your-secret-key-here/$SPICEDB_SECRET_KEY/g" docker-compose.yml

echo "📦 Starting SpiceDB with PostgreSQL..."
echo "   Database: $SPICEDB_DB_NAME"
echo "   Host: $DB_POSTGRES_HOST:$DB_POSTGRES_PORT"
echo "   User: $DB_POSTGRES_USER"

docker-compose up -d spicedb

echo "⏳ Waiting for SpiceDB to be ready..."
sleep 15

# Check if SpiceDB is healthy
if docker-compose ps spicedb | grep -q "Up"; then
    echo "✅ SpiceDB is running successfully!"
    echo ""
    echo "📊 SpiceDB Status:"
    echo "   - gRPC Endpoint: localhost:50051"
    echo "   - Database: $SPICEDB_DB_NAME"
    echo "   - Schema: Loaded from spicedb_schema.zed"
    echo ""
    echo "🔑 Secret Key: $SPICEDB_SECRET_KEY"
    echo ""
    echo "📋 Available commands:"
    echo "   - View logs: docker-compose logs -f spicedb"
    echo "   - Stop services: docker-compose down"
    echo "   - Restart services: docker-compose restart spicedb"
    echo ""
    echo "🔍 Test the connection:"
    echo "   grpcurl -plaintext -d '{}' localhost:50051 grpc.health.v1.Health/Check"
    echo ""
    echo "📝 Schema loaded:"
    echo "   - User management with role-based access"
    echo "   - Column-level permissions"
    echo "   - Resource-specific permissions"
    echo "   - Audit logging capabilities"
else
    echo "❌ Failed to start SpiceDB. Check logs:"
    docker-compose logs spicedb
    exit 1
fi
