-- Add Foreign Key Relationships

-- Add foreign key constraint for user_profiles.address_id
ALTER TABLE user_profiles
ADD CONSTRAINT fk_user_profiles_address_id
FOREIGN KEY (address_id) REFERENCES addresses(id);

-- Add additional indexes for better performance
CREATE INDEX idx_user_profiles_address_id ON user_profiles(address_id);
CREATE INDEX idx_contacts_address_id ON contacts(address_id);

-- Add constraints for status field
ALTER TABLE users
ADD CONSTRAINT chk_users_status
CHECK (status IN ('pending', 'active', 'suspended', 'blocked'));

-- Add constraints for country code
ALTER TABLE contacts
ADD CONSTRAINT chk_contacts_country_code
CHECK (country_code ~ '^\+[1-9][0-9]{0,3}$');

-- Add constraints for mobile number (basic validation)
ALTER TABLE contacts
ADD CONSTRAINT chk_contacts_mobile_number
CHECK (mobile_number > 1000000000 AND mobile_number < 99999999999);

-- Add constraints for pincode
ALTER TABLE addresses
ADD CONSTRAINT chk_addresses_pincode
CHECK (pincode ~ '^[0-9]{6}$');

-- Add constraints for username (alphanumeric, underscore, hyphen)
ALTER TABLE users
ADD CONSTRAINT chk_users_username
CHECK (username ~ '^[a-zA-Z0-9_-]{3,100}$');

-- Add constraints for permissions effect
ALTER TABLE permissions
ADD CONSTRAINT chk_permissions_effect
CHECK (effect IN ('allow', 'deny', 'override'));

-- Add unique constraint for aadhaar_number if provided
ALTER TABLE user_profiles
ADD CONSTRAINT uq_user_profiles_aadhaar_number
UNIQUE (aadhaar_number);

-- Add partial index for active user roles only
CREATE INDEX idx_user_roles_active ON user_roles(user_id, role_id) WHERE is_active = TRUE;

-- Add partial index for non-deleted records
CREATE INDEX idx_users_active ON users(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_active ON roles(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_active ON permissions(id) WHERE deleted_at IS NULL;

-- Add GIN index for permissions actions array
CREATE INDEX idx_permissions_actions ON permissions USING GIN(actions);
