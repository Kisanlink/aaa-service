-- Rollback V1/V2 Compatible Schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_v1_updated_at ON users_v1;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS update_memberships_updated_at ON memberships;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;
DROP TRIGGER IF EXISTS update_contacts_updated_at ON contacts;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS update_role_permissions_updated_at ON role_permissions;
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;

-- Drop V2 tables (in reverse dependency order)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS users;

-- Drop V1 tables (in reverse dependency order)
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users_v1;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();
