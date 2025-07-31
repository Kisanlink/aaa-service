-- Migration: Drop SpiceDB database
-- This migration drops the SpiceDB database

-- Terminate all connections to the spicedb database
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'spicedb' AND pid <> pg_backend_pid();

-- Drop the spicedb database
DROP DATABASE IF EXISTS spicedb;
