#!/bin/bash

# Run migrations and setup SpiceDB database
echo "Running migrations and setting up SpiceDB database..."

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
SPICEDB_DB_NAME=${SPICEDB_DB_NAME:-spicedb}
DB_SSL_MODE=${DB_SSL_MODE:-disable}

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c '\q' 2>/dev/null; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

# Run main application migrations
echo "Running main application migrations..."
migrate -path ./migrations -database "postgres://${DB_POSTGRES_USER}:${DB_POSTGRES_PASSWORD}@${DB_POSTGRES_HOST}:${DB_POSTGRES_PORT}/${DB_POSTGRES_DBNAME}?sslmode=${DB_SSL_MODE}" up

# Create SpiceDB database and set permissions
echo "Setting up SpiceDB database: $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "CREATE DATABASE $SPICEDB_DB_NAME;" 2>/dev/null || echo "Database $SPICEDB_DB_NAME already exists"

# Grant permissions for SpiceDB database
echo "Granting permissions to $DB_POSTGRES_USER on $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $SPICEDB_DB_NAME TO $DB_POSTGRES_USER;" 2>/dev/null || echo "Permission grant failed"

echo "Migration and SpiceDB database setup completed!"
echo "SpiceDB will create its own schema tables when it first connects."
