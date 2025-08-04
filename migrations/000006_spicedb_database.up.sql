-- Migration: Create SpiceDB database
-- This migration creates a separate database for SpiceDB

-- Create SpiceDB database if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'spicedb') THEN
        CREATE DATABASE spicedb;
    END IF;
END
$$;

-- Grant all privileges on the spicedb database to the current user
GRANT ALL PRIVILEGES ON DATABASE spicedb TO CURRENT_USER;

-- Connect to the spicedb database to set up additional permissions
\c spicedb;

-- Grant usage on schema
GRANT USAGE ON SCHEMA public TO CURRENT_USER;

-- Grant all privileges on all tables in the spicedb database
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO CURRENT_USER;

-- Grant all privileges on all sequences in the spicedb database
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO CURRENT_USER;

-- Grant all privileges on all functions in the spicedb database
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO CURRENT_USER;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO CURRENT_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO CURRENT_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO CURRENT_USER;
