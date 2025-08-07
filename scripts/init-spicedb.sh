#!/bin/bash

# Initialize SpiceDB database
echo "Initializing SpiceDB database..."

# Wait for PostgreSQL to be ready
until PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c '\q' 2>/dev/null; do
    echo "Waiting for PostgreSQL to be ready..."
    sleep 2
done

# Create SpiceDB database if it doesn't exist
echo "Creating SpiceDB database: $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "CREATE DATABASE $SPICEDB_DB_NAME;" 2>/dev/null || echo "Database $SPICEDB_DB_NAME already exists"

# Grant permissions
echo "Granting permissions to $DB_POSTGRES_USER on $SPICEDB_DB_NAME"
PGPASSWORD=$DB_POSTGRES_PASSWORD psql -h $DB_POSTGRES_HOST -p $DB_POSTGRES_PORT -U $DB_POSTGRES_USER -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $SPICEDB_DB_NAME TO $DB_POSTGRES_USER;" 2>/dev/null || echo "Permission grant failed"

echo "SpiceDB database initialization completed!"
