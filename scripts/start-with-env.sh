#!/bin/bash

# Set database environment variables
export DB_PRIMARY_BACKEND=gorm
export DB_POSTGRES_HOST=localhost
export DB_POSTGRES_PORT=5432
export DB_POSTGRES_USER=aaa_user
export DB_POSTGRES_PASSWORD=aaa_password
export DB_POSTGRES_DBNAME=aaa_service
export DB_POSTGRES_SSLMODE=disable
export DB_POSTGRES_MAX_CONNS=10
export DB_POSTGRES_IDLE_CONNS=5

# Set server configuration
export PORT=8080
export LOG_LEVEL=info

# Print the environment variables for debugging
echo "Environment variables set:"
echo "DB_PRIMARY_BACKEND=$DB_PRIMARY_BACKEND"
echo "DB_POSTGRES_HOST=$DB_POSTGRES_HOST"
echo "DB_POSTGRES_PORT=$DB_POSTGRES_PORT"
echo "DB_POSTGRES_USER=$DB_POSTGRES_USER"
echo "DB_POSTGRES_DBNAME=$DB_POSTGRES_DBNAME"
echo "DB_POSTGRES_SSLMODE=$DB_POSTGRES_SSLMODE"
echo "PORT=$PORT"
echo "LOG_LEVEL=$LOG_LEVEL"
echo ""

# Build the application
echo "Building application..."
go build -o bin/aaa-server cmd/server/main.go

# Run the application
echo "Starting application..."
./bin/aaa-server
