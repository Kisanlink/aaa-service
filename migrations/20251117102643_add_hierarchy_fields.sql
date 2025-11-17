-- Migration: Add hierarchy tracking fields and performance indexes
-- Date: 2025-11-17
-- Purpose: Support efficient hierarchy traversal and role inheritance for organizations and groups
-- Author: DevOps Team

-- ============================================
-- UP MIGRATION
-- ============================================

-- --------------------------------------------
-- 1. ADD HIERARCHY TRACKING FIELDS
-- --------------------------------------------

-- Add hierarchy fields to organizations table
-- hierarchy_depth: tracks nesting level (0 = root, max 10 levels)
-- hierarchy_path: materialized path for efficient ancestor/descendant queries (e.g., '/root/parent/child')
-- version: optimistic locking for concurrent updates
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0 CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 10),
ADD COLUMN IF NOT EXISTS hierarchy_path TEXT,
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1 CHECK (version > 0);

COMMENT ON COLUMN organizations.hierarchy_depth IS 'Depth level in organization hierarchy (0=root, max=10)';
COMMENT ON COLUMN organizations.hierarchy_path IS 'Materialized path for efficient hierarchy queries';
COMMENT ON COLUMN organizations.version IS 'Version number for optimistic locking';

-- Add hierarchy fields to groups table
-- Similar structure but max depth is 8 for groups
ALTER TABLE groups
ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0 CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 8),
ADD COLUMN IF NOT EXISTS hierarchy_path TEXT,
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1 CHECK (version > 0);

COMMENT ON COLUMN groups.hierarchy_depth IS 'Depth level in group hierarchy (0=root, max=8)';
COMMENT ON COLUMN groups.hierarchy_path IS 'Materialized path for efficient hierarchy queries';
COMMENT ON COLUMN groups.version IS 'Version number for optimistic locking';

-- Add version field to group_memberships for tracking changes
ALTER TABLE group_memberships
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1 CHECK (version > 0);

COMMENT ON COLUMN group_memberships.version IS 'Version number for optimistic locking';

-- --------------------------------------------
-- 2. CREATE PERFORMANCE INDEXES
-- --------------------------------------------

-- Organization hierarchy indexes
-- Index for efficient parent-child queries with active filter
CREATE INDEX IF NOT EXISTS idx_org_parent_active
ON organizations(parent_id, is_active)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_org_parent_active IS 'Optimize finding active child organizations';

-- Index for hierarchy path queries (prefix matching)
CREATE INDEX IF NOT EXISTS idx_org_hierarchy_path
ON organizations USING btree(hierarchy_path text_pattern_ops)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_org_hierarchy_path IS 'Optimize hierarchy path prefix searches';

-- Index for filtering by hierarchy depth
CREATE INDEX IF NOT EXISTS idx_org_hierarchy_depth
ON organizations(hierarchy_depth)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_org_hierarchy_depth IS 'Optimize queries filtering by organization depth';

-- Composite index for hierarchy traversal
CREATE INDEX IF NOT EXISTS idx_org_hierarchy_composite
ON organizations(parent_id, hierarchy_depth, is_active)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_org_hierarchy_composite IS 'Optimize complex hierarchy traversal queries';

-- Group hierarchy indexes
-- Index for parent-organization relationship
CREATE INDEX IF NOT EXISTS idx_group_parent_org
ON groups(parent_id, organization_id)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_parent_org IS 'Optimize group hierarchy within organizations';

-- Index for group hierarchy path queries
CREATE INDEX IF NOT EXISTS idx_group_hierarchy_path
ON groups USING btree(hierarchy_path text_pattern_ops)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_hierarchy_path IS 'Optimize group hierarchy path searches';

-- Index for organization-level group queries
CREATE INDEX IF NOT EXISTS idx_group_org_active
ON groups(organization_id, is_active)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_org_active IS 'Optimize finding active groups in organization';

-- Index for group hierarchy depth filtering
CREATE INDEX IF NOT EXISTS idx_group_hierarchy_depth
ON groups(hierarchy_depth)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_hierarchy_depth IS 'Optimize queries filtering by group depth';

-- Composite index for group hierarchy traversal
CREATE INDEX IF NOT EXISTS idx_group_hierarchy_composite
ON groups(organization_id, parent_id, hierarchy_depth, is_active)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_hierarchy_composite IS 'Optimize complex group hierarchy queries';

-- Group membership version index for change tracking
CREATE INDEX IF NOT EXISTS idx_group_membership_version
ON group_memberships(version, updated_at)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_group_membership_version IS 'Optimize version-based change queries';

-- --------------------------------------------
-- 3. POPULATE INITIAL VALUES FOR EXISTING DATA
-- --------------------------------------------

-- Function to calculate hierarchy depth
CREATE OR REPLACE FUNCTION calculate_hierarchy_depth(
    table_name TEXT,
    record_id VARCHAR(255),
    parent_column TEXT DEFAULT 'parent_id',
    max_depth INTEGER DEFAULT 10
) RETURNS INTEGER AS $$
DECLARE
    depth INTEGER := 0;
    current_parent VARCHAR(255);
BEGIN
    IF table_name = 'organizations' THEN
        SELECT parent_id INTO current_parent
        FROM organizations
        WHERE id = record_id AND deleted_at IS NULL;
    ELSIF table_name = 'groups' THEN
        SELECT parent_id INTO current_parent
        FROM groups
        WHERE id = record_id AND deleted_at IS NULL;
    ELSE
        RETURN 0;
    END IF;

    WHILE current_parent IS NOT NULL AND depth < max_depth LOOP
        depth := depth + 1;

        IF table_name = 'organizations' THEN
            SELECT parent_id INTO current_parent
            FROM organizations
            WHERE id = current_parent AND deleted_at IS NULL;
        ELSIF table_name = 'groups' THEN
            SELECT parent_id INTO current_parent
            FROM groups
            WHERE id = current_parent AND deleted_at IS NULL;
        END IF;
    END LOOP;

    RETURN depth;
END;
$$ LANGUAGE plpgsql;

-- Function to build hierarchy path
CREATE OR REPLACE FUNCTION build_hierarchy_path(
    table_name TEXT,
    record_id VARCHAR(255)
) RETURNS TEXT AS $$
DECLARE
    path_array VARCHAR(255)[] := ARRAY[]::VARCHAR(255)[];
    current_id VARCHAR(255) := record_id;
    current_parent VARCHAR(255);
    max_iterations INTEGER := 20;
    iterations INTEGER := 0;
BEGIN
    -- Build path from bottom to top
    WHILE current_id IS NOT NULL AND iterations < max_iterations LOOP
        path_array := array_prepend(current_id, path_array);

        IF table_name = 'organizations' THEN
            SELECT parent_id INTO current_parent
            FROM organizations
            WHERE id = current_id AND deleted_at IS NULL;
        ELSIF table_name = 'groups' THEN
            SELECT parent_id INTO current_parent
            FROM groups
            WHERE id = current_id AND deleted_at IS NULL;
        ELSE
            EXIT;
        END IF;

        current_id := current_parent;
        iterations := iterations + 1;
    END LOOP;

    -- Return path as /id1/id2/id3
    IF array_length(path_array, 1) > 0 THEN
        RETURN '/' || array_to_string(path_array, '/');
    ELSE
        RETURN NULL;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Update organizations with calculated hierarchy values
UPDATE organizations
SET
    hierarchy_depth = calculate_hierarchy_depth('organizations', id),
    hierarchy_path = build_hierarchy_path('organizations', id),
    version = COALESCE(version, 1)
WHERE deleted_at IS NULL;

-- Update groups with calculated hierarchy values
UPDATE groups
SET
    hierarchy_depth = calculate_hierarchy_depth('groups', id),
    hierarchy_path = build_hierarchy_path('groups', id),
    version = COALESCE(version, 1)
WHERE deleted_at IS NULL;

-- Update group_memberships version
UPDATE group_memberships
SET version = COALESCE(version, 1)
WHERE deleted_at IS NULL;

-- --------------------------------------------
-- 4. CREATE TRIGGERS FOR MAINTAINING HIERARCHY
-- --------------------------------------------

-- Trigger function to maintain hierarchy fields on insert/update
CREATE OR REPLACE FUNCTION maintain_hierarchy_fields() RETURNS TRIGGER AS $$
DECLARE
    parent_depth INTEGER;
    parent_path TEXT;
    table_max_depth INTEGER;
BEGIN
    -- Set max depth based on table
    IF TG_TABLE_NAME = 'organizations' THEN
        table_max_depth := 10;
    ELSIF TG_TABLE_NAME = 'groups' THEN
        table_max_depth := 8;
    ELSE
        RETURN NEW;
    END IF;

    -- If parent_id is NULL, this is a root node
    IF NEW.parent_id IS NULL THEN
        NEW.hierarchy_depth := 0;
        NEW.hierarchy_path := '/' || NEW.id;
    ELSE
        -- Get parent's depth and path
        IF TG_TABLE_NAME = 'organizations' THEN
            SELECT hierarchy_depth, hierarchy_path INTO parent_depth, parent_path
            FROM organizations
            WHERE id = NEW.parent_id AND deleted_at IS NULL;
        ELSIF TG_TABLE_NAME = 'groups' THEN
            SELECT hierarchy_depth, hierarchy_path INTO parent_depth, parent_path
            FROM groups
            WHERE id = NEW.parent_id AND deleted_at IS NULL;
        END IF;

        -- Check if parent exists
        IF parent_depth IS NULL THEN
            RAISE EXCEPTION 'Parent record % not found', NEW.parent_id;
        END IF;

        -- Check depth limit
        IF parent_depth >= table_max_depth - 1 THEN
            RAISE EXCEPTION 'Maximum hierarchy depth (%) exceeded', table_max_depth;
        END IF;

        -- Set hierarchy fields
        NEW.hierarchy_depth := parent_depth + 1;
        NEW.hierarchy_path := parent_path || '/' || NEW.id;
    END IF;

    -- Increment version on update
    IF TG_OP = 'UPDATE' AND OLD.version IS NOT NULL THEN
        NEW.version := OLD.version + 1;
    ELSIF NEW.version IS NULL THEN
        NEW.version := 1;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for organizations
DROP TRIGGER IF EXISTS maintain_org_hierarchy ON organizations;
CREATE TRIGGER maintain_org_hierarchy
    BEFORE INSERT OR UPDATE OF parent_id ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION maintain_hierarchy_fields();

-- Create triggers for groups
DROP TRIGGER IF EXISTS maintain_group_hierarchy ON groups;
CREATE TRIGGER maintain_group_hierarchy
    BEFORE INSERT OR UPDATE OF parent_id ON groups
    FOR EACH ROW
    EXECUTE FUNCTION maintain_hierarchy_fields();

-- --------------------------------------------
-- 5. ANALYZE TABLES FOR QUERY OPTIMIZATION
-- --------------------------------------------

ANALYZE organizations;
ANALYZE groups;
ANALYZE group_memberships;

-- ============================================
-- ROLLBACK MIGRATION (DOWN)
-- ============================================

-- To rollback this migration, save and run the following SQL:

/*
-- Drop triggers
DROP TRIGGER IF EXISTS maintain_org_hierarchy ON organizations;
DROP TRIGGER IF EXISTS maintain_group_hierarchy ON groups;

-- Drop functions
DROP FUNCTION IF EXISTS maintain_hierarchy_fields();
DROP FUNCTION IF EXISTS build_hierarchy_path(TEXT, VARCHAR(255));
DROP FUNCTION IF EXISTS calculate_hierarchy_depth(TEXT, VARCHAR(255), TEXT, INTEGER);

-- Drop indexes for group_memberships
DROP INDEX IF EXISTS idx_group_membership_version;

-- Drop indexes for groups
DROP INDEX IF EXISTS idx_group_hierarchy_composite;
DROP INDEX IF EXISTS idx_group_hierarchy_depth;
DROP INDEX IF EXISTS idx_group_org_active;
DROP INDEX IF EXISTS idx_group_hierarchy_path;
DROP INDEX IF EXISTS idx_group_parent_org;

-- Drop indexes for organizations
DROP INDEX IF EXISTS idx_org_hierarchy_composite;
DROP INDEX IF EXISTS idx_org_hierarchy_depth;
DROP INDEX IF EXISTS idx_org_hierarchy_path;
DROP INDEX IF EXISTS idx_org_parent_active;

-- Remove columns from group_memberships
ALTER TABLE group_memberships DROP COLUMN IF EXISTS version;

-- Remove columns from groups
ALTER TABLE groups DROP COLUMN IF EXISTS version;
ALTER TABLE groups DROP COLUMN IF EXISTS hierarchy_path;
ALTER TABLE groups DROP COLUMN IF EXISTS hierarchy_depth;

-- Remove columns from organizations
ALTER TABLE organizations DROP COLUMN IF EXISTS version;
ALTER TABLE organizations DROP COLUMN IF EXISTS hierarchy_path;
ALTER TABLE organizations DROP COLUMN IF EXISTS hierarchy_depth;

-- Re-analyze tables
ANALYZE organizations;
ANALYZE groups;
ANALYZE group_memberships;
*/

-- ============================================
-- VERIFICATION QUERIES
-- ============================================

-- After running this migration, verify with:
/*
-- Check new columns exist
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name IN ('organizations', 'groups', 'group_memberships')
AND column_name IN ('hierarchy_depth', 'hierarchy_path', 'version')
ORDER BY table_name, column_name;

-- Check indexes were created
SELECT schemaname, tablename, indexname, indexdef
FROM pg_indexes
WHERE tablename IN ('organizations', 'groups', 'group_memberships')
AND indexname LIKE 'idx_%hierarchy%' OR indexname LIKE 'idx_%version%'
ORDER BY tablename, indexname;

-- Check constraints
SELECT
    tc.table_name,
    tc.constraint_name,
    tc.constraint_type,
    cc.check_clause
FROM information_schema.table_constraints tc
LEFT JOIN information_schema.check_constraints cc
    ON tc.constraint_name = cc.constraint_name
WHERE tc.table_name IN ('organizations', 'groups', 'group_memberships')
AND tc.constraint_type = 'CHECK'
ORDER BY tc.table_name, tc.constraint_name;

-- Verify data population
SELECT
    'organizations' as table_name,
    COUNT(*) as total_records,
    COUNT(hierarchy_depth) as with_depth,
    COUNT(hierarchy_path) as with_path,
    COUNT(version) as with_version
FROM organizations
WHERE deleted_at IS NULL
UNION ALL
SELECT
    'groups' as table_name,
    COUNT(*) as total_records,
    COUNT(hierarchy_depth) as with_depth,
    COUNT(hierarchy_path) as with_path,
    COUNT(version) as with_version
FROM groups
WHERE deleted_at IS NULL;

-- Sample hierarchy paths
SELECT id, name, parent_id, hierarchy_depth, hierarchy_path
FROM organizations
WHERE deleted_at IS NULL
ORDER BY hierarchy_depth, name
LIMIT 10;

SELECT id, name, parent_id, hierarchy_depth, hierarchy_path
FROM groups
WHERE deleted_at IS NULL
ORDER BY hierarchy_depth, name
LIMIT 10;
*/
