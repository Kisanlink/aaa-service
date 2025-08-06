-- Revert phone number and mPin changes
-- Note: This will lose phone number and mPin data
ALTER TABLE users
DROP COLUMN IF EXISTS phone_number,
DROP COLUMN IF EXISTS country_code,
DROP COLUMN IF EXISTS mpin;

-- Make username required again
ALTER TABLE users
ALTER COLUMN username SET NOT NULL;

-- Drop the index
DROP INDEX IF EXISTS idx_users_phone_country;
