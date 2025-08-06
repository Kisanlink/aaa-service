-- Add phone number and mPin support to users table
ALTER TABLE users
ADD COLUMN phone_number VARCHAR(20) UNIQUE,
ADD COLUMN country_code VARCHAR(10) DEFAULT '+91',
ADD COLUMN mpin VARCHAR(255);

-- Create index on phone_number and country_code combination
CREATE INDEX idx_users_phone_country ON users(phone_number, country_code);

-- Make username nullable since phone number will be primary identifier
ALTER TABLE users
ALTER COLUMN username DROP NOT NULL,
ALTER COLUMN username TYPE VARCHAR(100);

-- Update contacts table to remove mobile_number since it will be in users table
-- Note: This is a breaking change - existing data needs to be migrated
-- You may want to handle this differently based on your requirements

-- Add comment to track the migration
COMMENT ON COLUMN users.phone_number IS 'Primary phone number for authentication';
COMMENT ON COLUMN users.country_code IS 'Country code for the phone number';
COMMENT ON COLUMN users.mpin IS 'Hashed mPin for secure refresh token generation';
