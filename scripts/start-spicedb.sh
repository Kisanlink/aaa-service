#!/bin/bash

# Start SpiceDB locally for development with PostgreSQL backend
echo "Starting SpiceDB locally with PostgreSQL backend..."

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "Loaded environment variables from .env file"
else
    echo "Warning: .env file not found, using default values"
fi

# Set default values if not in .env
DB_POSTGRES_HOST=${DB_POSTGRES_HOST:-localhost}
DB_POSTGRES_PORT=${DB_POSTGRES_PORT:-5432}
DB_POSTGRES_USER=${DB_POSTGRES_USER:-aaa_user}
DB_POSTGRES_PASSWORD=${DB_POSTGRES_PASSWORD:-aaa_password}
DB_POSTGRES_DBNAME=${DB_POSTGRES_DBNAME:-aaa_service}

# Check if SpiceDB container is already running
if docker ps | grep -q "aaa-spicedb"; then
    echo "SpiceDB is already running"
    exit 0
fi

# Start SpiceDB container with PostgreSQL backend
docker run -d \
    --name aaa-spicedb \
    --rm \
    -p 50051:50051 \
    -p 50052:50052 \
    -e SPICEDB_LOG_LEVEL=info \
    -e SPICEDB_GRPC_PRESHARED_KEY=your-secret-key-here \
    -e SPICEDB_DATABASE_ENGINE=postgres \
    -e SPICEDB_DATABASE_CONN_URI=postgres://${DB_POSTGRES_USER}:${DB_POSTGRES_PASSWORD}@${DB_POSTGRES_HOST}:${DB_POSTGRES_PORT}/${DB_POSTGRES_DBNAME}?sslmode=disable \
    --network aaa-network \
    authzed/spicedb:latest

echo "SpiceDB started on localhost:50051 with PostgreSQL backend"
echo "Connection URI: postgres://${DB_POSTGRES_USER}:***@${DB_POSTGRES_HOST}:${DB_POSTGRES_PORT}/${DB_POSTGRES_DBNAME}"
echo "You can stop it with: docker stop aaa-spicedb"
