# Hierarchy Fields Migration Documentation

## Overview
This migration adds hierarchy tracking fields and performance indexes to support efficient hierarchy traversal and role inheritance for organizations and groups in the AAA service.

## Migration Details

### Migration Files
- **Go Migration**: `migrations/add_hierarchy_fields.go`
- **SQL Migration**: `migrations/20251117102643_add_hierarchy_fields.sql` (for reference/manual execution)

### Database Changes

#### 1. New Fields Added

**Organizations Table**:
- `hierarchy_depth` (INTEGER): Tracks nesting level (0=root, max=10)
- `hierarchy_path` (TEXT): Materialized path for efficient ancestor/descendant queries
- `version` (INTEGER): Optimistic locking for concurrent updates

**Groups Table**:
- `hierarchy_depth` (INTEGER): Tracks nesting level (0=root, max=8)
- `hierarchy_path` (TEXT): Materialized path for efficient hierarchy queries
- `version` (INTEGER): Optimistic locking for concurrent updates

**Group Memberships Table**:
- `version` (INTEGER): Version tracking for change management

#### 2. Performance Indexes Created

**Organization Indexes**:
- `idx_org_parent_active`: Optimize finding active child organizations
- `idx_org_hierarchy_path`: Optimize hierarchy path prefix searches
- `idx_org_hierarchy_depth`: Optimize queries filtering by organization depth
- `idx_org_hierarchy_composite`: Optimize complex hierarchy traversal queries

**Group Indexes**:
- `idx_group_parent_org`: Optimize group hierarchy within organizations
- `idx_group_hierarchy_path`: Optimize group hierarchy path searches
- `idx_group_hierarchy_depth`: Optimize queries filtering by group depth
- `idx_group_hierarchy_composite`: Optimize complex group hierarchy queries

**Membership Indexes**:
- `idx_group_membership_version`: Optimize version-based change queries

#### 3. Helper Functions Created

- `calculate_hierarchy_depth()`: Calculates the depth of a record in the hierarchy
- `build_hierarchy_path()`: Builds the materialized path for a record

### Model Changes

The following model files have been updated with hierarchy fields:

**`internal/entities/models/organization.go`**:
- Added `HierarchyDepth` and `HierarchyPath` fields
- Added hooks to maintain hierarchy fields on create/update

**`internal/entities/models/group.go`**:
- Added `HierarchyDepth` and `HierarchyPath` fields
- Added hooks to maintain hierarchy fields on create/update

### Auto-Migration Integration

The migration is automatically run on server startup through:
1. GORM auto-migration adds the columns based on model definitions
2. `AddHierarchyFields()` function creates indexes and populates existing data
3. `ValidateHierarchyMigration()` verifies the migration was successful

### How It Works

#### Hierarchy Path
The hierarchy path is a materialized path pattern that stores the full path from root to node:
- Example: `/org1/org2/org3` represents org3 is child of org2 which is child of org1
- Enables efficient queries like "find all descendants" using path prefix matching

#### Hierarchy Depth
Stores the nesting level to:
- Enforce maximum depth constraints (10 for orgs, 8 for groups)
- Optimize queries that need to filter by level
- Prevent infinite recursion in hierarchy traversal

#### Version Field
Implements optimistic locking to handle concurrent updates safely:
- Increments on each update
- Prevents lost updates in concurrent scenarios
- Tracks change history for auditing

### Query Performance Improvements

With these indexes and fields, the following queries are optimized:

1. **Find all children of an organization**:
```sql
SELECT * FROM organizations
WHERE parent_id = ? AND deleted_at IS NULL
-- Uses: idx_org_parent_active
```

2. **Find all descendants of an organization**:
```sql
SELECT * FROM organizations
WHERE hierarchy_path LIKE '/org123/%' AND deleted_at IS NULL
-- Uses: idx_org_hierarchy_path
```

3. **Find organizations at specific depth**:
```sql
SELECT * FROM organizations
WHERE hierarchy_depth = 2 AND deleted_at IS NULL
-- Uses: idx_org_hierarchy_depth
```

4. **Complex hierarchy traversal**:
```sql
SELECT * FROM groups
WHERE organization_id = ?
  AND parent_id = ?
  AND hierarchy_depth < 5
  AND is_active = true
  AND deleted_at IS NULL
-- Uses: idx_group_hierarchy_composite
```

### Rollback Plan

If rollback is needed:

1. **Remove indexes**: Run `DropHierarchyFields()` function
2. **Remove columns**: Update models to remove hierarchy fields and run auto-migration
3. **Drop functions**: Execute DROP FUNCTION statements for helper functions

### Testing Checklist

- [ ] Verify columns exist in database
- [ ] Verify indexes are created and being used
- [ ] Test creating new organizations with hierarchy
- [ ] Test creating new groups with hierarchy
- [ ] Verify hierarchy depth constraints are enforced
- [ ] Test hierarchy path updates when parent changes
- [ ] Verify query performance improvements
- [ ] Test concurrent updates with version field

### Monitoring

Monitor the following after deployment:
- Index usage statistics
- Query performance metrics
- Hierarchy depth distribution
- Version conflict frequency

### Notes

- The migration is idempotent and can be run multiple times safely
- Existing data is automatically populated with correct hierarchy values
- The migration runs asynchronously and doesn't block server startup
- Manual execution is possible using the SQL file if needed
