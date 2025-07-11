-- Remove Foreign Keys and Constraints

-- Remove GIN index
DROP INDEX IF EXISTS idx_permissions_actions;

-- Remove partial indexes
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_roles_active;
DROP INDEX IF EXISTS idx_permissions_active;
DROP INDEX IF EXISTS idx_user_roles_active;

-- Remove constraints
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS chk_permissions_effect;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_username;
ALTER TABLE addresses DROP CONSTRAINT IF EXISTS chk_addresses_pincode;
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS chk_contacts_mobile_number;
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS chk_contacts_country_code;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_status;
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS uq_user_profiles_aadhaar_number;

-- Remove performance indexes
DROP INDEX IF EXISTS idx_contacts_address_id;
DROP INDEX IF EXISTS idx_user_profiles_address_id;

-- Remove foreign key constraints
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS fk_user_profiles_address_id;
