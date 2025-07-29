-- Down migration for drawSQL schema with data
-- Reverses 000004_drawsql_schema_with_data.up.sql

-- Drop views
DROP VIEW IF EXISTS user_permissions;
DROP VIEW IF EXISTS active_users_with_profiles;

-- Drop role permissions junction table
DROP TABLE IF EXISTS role_permissions CASCADE;

-- Drop user roles junction table
DROP TABLE IF EXISTS user_roles CASCADE;

-- Drop contacts table
DROP TABLE IF EXISTS contacts CASCADE;

-- Drop user profiles table
DROP TABLE IF EXISTS user_profiles CASCADE;

-- Drop permissions table
DROP TABLE IF EXISTS permissions CASCADE;

-- Drop roles table
DROP TABLE IF EXISTS roles CASCADE;

-- Drop users table
DROP TABLE IF EXISTS users CASCADE;

-- Drop addresses table
DROP TABLE IF EXISTS addresses CASCADE;
