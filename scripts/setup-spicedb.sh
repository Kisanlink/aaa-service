#!/bin/bash

# Setup script for SpiceDB database in external PostgreSQL
# This script creates the required database for SpiceDB

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
DB_HOST=${DB_POSTGRES_HOST:-localhost}
DB_PORT=${DB_POSTGRES_PORT:-5432}
DB_USER=${DB_POSTGRES_USER:-aaa_user}
DB_PASSWORD=${DB_POSTGRES_PASSWORD:-aaa_password}
SPICEDB_DB_NAME=${SPICEDB_DB_NAME:-spicedb}

echo "Setting up SpiceDB database..."
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "User: $DB_USER"
echo "Database: $SPICEDB_DB_NAME"

# Create SpiceDB database if it doesn't exist
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $SPICEDB_DB_NAME;" 2>/dev/null || echo "Database $SPICEDB_DB_NAME already exists or creation failed"

echo "SpiceDB database setup complete!"
echo ""
echo "Next steps:"
echo "1. Start the SpiceDB container: docker-compose up spicedb"
echo "2. SpiceDB will automatically create its required tables in the $SPICEDB_DB_NAME database"
echo "3. Start the AAA service: make run"
