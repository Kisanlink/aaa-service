-- Migration: Add user_id column to addresses table
-- Date: 2025-11-05
-- Description: Adds user_id field to addresses table to establish user-address relationship

-- Add user_id column to addresses table
ALTER TABLE addresses ADD COLUMN IF NOT EXISTS user_id VARCHAR(255);

-- Create index on user_id for efficient lookups
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);

-- Create composite index for user_id + pincode (common query pattern)
CREATE INDEX IF NOT EXISTS idx_addresses_user_pincode ON addresses(user_id, pincode);

-- Create composite index for district + state (for location-based queries)
CREATE INDEX IF NOT EXISTS idx_addresses_district_state ON addresses(district, state);

-- Create index on pincode for postal code searches
CREATE INDEX IF NOT EXISTS idx_addresses_pincode ON addresses(pincode);

-- Add check constraint for pincode format (6-digit Indian postal code)
ALTER TABLE addresses ADD CONSTRAINT chk_pincode_format
    CHECK (pincode IS NULL OR pincode ~ '^[1-9][0-9]{5}$');

-- Note: user_id is initially nullable to allow migration of existing data
-- After data migration, run: ALTER TABLE addresses ALTER COLUMN user_id SET NOT NULL;
