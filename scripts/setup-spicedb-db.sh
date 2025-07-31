#!/bin/bash

# Setup SpiceDB database on the same PostgreSQL server
echo "Setting up SpiceDB database..."

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
SPICEDB_DB_NAME=${SPICEDB_DB_NAME:-spicedb}

# Create SpiceDB database
echo "Creating SpiceDB database: $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "CREATE DATABASE $SPICEDB_DB_NAME;" 2>/dev/null || echo "Database $SPICEDB_DB_NAME already exists or creation failed"

# Grant permissions
echo "Granting permissions to $DB_POSTGRES_USER on $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $SPICEDB_DB_NAME TO $DB_POSTGRES_USER;" 2>/dev/null || echo "Permission grant failed"

echo "SpiceDB database setup completed!"
echo "Database: $SPICEDB_DB_NAME"
echo "Host: $DB_POSTGRES_HOST:$DB_POSTGRES_PORT"
