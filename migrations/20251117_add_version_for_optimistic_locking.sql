-- Migration: Add version column for optimistic locking
-- Date: 2025-11-17
-- Description: Adds version field to organizations, groups, and group_memberships tables
--              to support optimistic locking and prevent race conditions in concurrent updates

-- Add version column to organizations table
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Add version column to groups table
ALTER TABLE groups ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Add version column to group_memberships table
ALTER TABLE group_memberships ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Create indexes for optimistic locking queries (optional but improves performance)
-- These indexes help when checking version during concurrent updates
CREATE INDEX IF NOT EXISTS idx_organizations_id_version ON organizations(id, version);
CREATE INDEX IF NOT EXISTS idx_groups_id_version ON groups(id, version);
CREATE INDEX IF NOT EXISTS idx_group_memberships_id_version ON group_memberships(id, version);

-- Add comments to document the purpose of version columns
COMMENT ON COLUMN organizations.version IS 'Optimistic locking version counter, incremented on each update';
COMMENT ON COLUMN groups.version IS 'Optimistic locking version counter, incremented on each update';
COMMENT ON COLUMN group_memberships.version IS 'Optimistic locking version counter, incremented on each update';
