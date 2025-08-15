#!/bin/bash

# Load environment variables from .env file
export $(cat .env | grep -v '^#' | xargs)

# Verify environment variables are loaded
echo "Environment variables loaded:"
echo "DB_PRIMARY_BACKEND: $DB_PRIMARY_BACKEND"
echo "DB_POSTGRES_HOST: $DB_POSTGRES_HOST"
echo "DB_POSTGRES_DBNAME: $DB_POSTGRES_DBNAME"

# Start the service
echo "Starting aaa-service..."
go run cmd/server/main.go
