-- Down migration for drawSQL clean schema
-- Reverses 000005_drawsql_schema_clean.up.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_role_id;
DROP INDEX IF EXISTS idx_user_roles_is_active;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_user_roles_deleted_at;
DROP INDEX IF EXISTS idx_contacts_address_id;
DROP INDEX IF EXISTS idx_contacts_mobile_number;
DROP INDEX IF EXISTS idx_contacts_user_id;
DROP INDEX IF EXISTS idx_contacts_deleted_at;
DROP INDEX IF EXISTS idx_user_profiles_address_id;
DROP INDEX IF EXISTS idx_user_profiles_user_id;
DROP INDEX IF EXISTS idx_user_profiles_deleted_at;
DROP INDEX IF EXISTS idx_permissions_resource;
DROP INDEX IF EXISTS idx_permissions_deleted_at;
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_roles_deleted_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_addresses_deleted_at;

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
