# Organization and Group Hierarchy Fixes

## Problem Statement

The hierarchy implementation for organizations and groups is incomplete. While the data models and database schema have the necessary fields (`parent_id`), the repository and service layers are missing critical hierarchy traversal methods, leading to:

1. **Missing group hierarchy traversal methods** in GroupRepository
2. **Incomplete descendant retrieval** for both organizations and groups
3. **Inconsistent hierarchy API coverage** between organizations and groups
4. **Potential performance issues** with recursive hierarchy queries

## Current State Analysis

### Database Schema (CORRECT)

Both tables have proper schema:

**Organizations Table:**
- `parent_id` VARCHAR(255) references `organizations(id)`
- Index: `idx_organizations_parent` on `parent_id` WHERE deleted_at IS NULL AND parent_id IS NOT NULL
- Foreign key: `fk_organizations_children`

**Groups Table:**
- `parent_id` VARCHAR(255) references `groups(id)`
- `organization_id` VARCHAR(255) references `organizations(id)` (NOT NULL)
- Index: `idx_groups_parent` on `parent_id` WHERE deleted_at IS NULL AND parent_id IS NOT NULL
- Foreign key: `fk_groups_children`

**group_inheritance Table:**
- Exists but appears to be unused
- Has `parent_group_id` and `child_group_id`
- May be redundant with `groups.parent_id`

### Code Issues Identified

#### OrganizationRepository (PARTIALLY COMPLETE)
✓ Has `GetChildren()`
✓ Has `GetActiveChildren()`
✓ Has `GetParentHierarchy()` - **but implementation is inefficient (N+1 query pattern)**
✗ Missing `GetDescendants()` - no method to get all descendants recursively
✗ Missing `GetAncestors()` - alias/alternative to GetParentHierarchy

#### GroupRepository (MISSING CRITICAL METHODS)
✓ Has `GetChildren()` - only gets immediate children
✗ Missing `GetParentHierarchy()` - **CRITICAL MISSING METHOD**
✗ Missing `GetActiveChildren()` - should exist for consistency
✗ Missing `GetDescendants()` - no recursive descendant retrieval
✗ Missing `GetAncestors()` - no upward traversal

#### GroupService (MISSING GROUP HIERARCHY API)
✓ Has `checkCircularReference()` for validation
✗ Missing `GetGroupHierarchy()` - **should exist like OrganizationService**
✗ The service uses `buildGroupHierarchy()` for org-level view but not for individual groups

#### OrganizationService (COMPLETE)
✓ Has `GetOrganizationHierarchy()` with caching
✓ Has `checkCircularReference()`
✓ Uses `buildGroupHierarchy()` to include groups in org hierarchy

## Requirements

### 1. Repository Layer - Critical Missing Methods

#### GroupRepository Must Add:

```go
// GetParentHierarchy retrieves the complete parent hierarchy for a group
// Returns ordered list from root to immediate parent
func (r *GroupRepository) GetParentHierarchy(ctx context.Context, groupID string) ([]*models.Group, error)

// GetActiveChildren retrieves only active child groups
func (r *GroupRepository) GetActiveChildren(ctx context.Context, parentID string) ([]*models.Group, error)

// GetDescendants retrieves all descendant groups recursively
// Includes children, grandchildren, etc.
func (r *GroupRepository) GetDescendants(ctx context.Context, groupID string) ([]*models.Group, error)

// GetDescendantsDepth retrieves descendants up to a specified depth
func (r *GroupRepository) GetDescendantsDepth(ctx context.Context, groupID string, maxDepth int) ([]*models.Group, error)
```

#### OrganizationRepository Must Add:

```go
// GetDescendants retrieves all descendant organizations recursively
func (r *OrganizationRepository) GetDescendants(ctx context.Context, orgID string) ([]*models.Organization, error)

// GetDescendantsDepth retrieves descendants up to a specified depth
func (r *OrganizationRepository) GetDescendantsDepth(ctx context.Context, orgID string, maxDepth int) ([]*models.Organization, error)
```

#### OrganizationRepository Must Fix:

```go
// GetParentHierarchy - optimize to use single recursive CTE query instead of N+1 pattern
```

### 2. Service Layer - Missing Group Hierarchy Operations

#### GroupService Must Add:

```go
// GetGroupHierarchy retrieves the complete hierarchy for a group
// Similar to OrganizationService.GetOrganizationHierarchy
func (s *Service) GetGroupHierarchy(ctx context.Context, groupID string) (*groupResponses.GroupHierarchyResponse, error)

// GetGroupWithParents retrieves a group with its full parent chain
func (s *Service) GetGroupWithParents(ctx context.Context, groupID string) (*groupResponses.GroupWithParentsResponse, error)

// GetGroupWithChildren retrieves a group with its immediate children
func (s *Service) GetGroupWithChildren(ctx context.Context, groupID string) (*groupResponses.GroupWithChildrenResponse, error)
```

### 3. Performance Requirements

- **Use PostgreSQL recursive CTEs** for hierarchy traversal instead of N+1 queries
- **Cache hierarchy results** with appropriate TTL (5-15 minutes)
- **Invalidate caches** when hierarchy changes (parent_id updates)
- **Add indexes** if missing:
  - `idx_organizations_parent` ✓ (exists)
  - `idx_groups_parent` ✓ (exists)
  - `idx_groups_org` ✓ (exists)

### 4. Response Models

Create these if missing:

```go
// GroupHierarchyResponse - complete hierarchy view
type GroupHierarchyResponse struct {
    Group    *GroupResponse   `json:"group"`
    Parents  []*GroupResponse `json:"parents"`   // Ordered from root to immediate parent
    Children []*GroupResponse `json:"children"`  // Immediate children
    Roles    []*GroupRoleDetail `json:"roles"`   // Roles assigned to this group
}

// GroupWithParentsResponse - group with ancestors only
type GroupWithParentsResponse struct {
    Group   *GroupResponse   `json:"group"`
    Parents []*GroupResponse `json:"parents"` // Ordered from root to immediate parent
}

// GroupWithChildrenResponse - group with descendants only
type GroupWithChildrenResponse struct {
    Group    *GroupResponse   `json:"group"`
    Children []*GroupHierarchyNode `json:"children"` // Recursive children tree
}
```

### 5. API Endpoints

Ensure these endpoints exist and work correctly:

```
GET /api/v1/organizations/:id/hierarchy     ✓ (exists)
GET /api/v1/groups/:id/hierarchy             ✗ (MISSING)
GET /api/v1/groups/:id/parents               ✗ (MISSING)
GET /api/v1/groups/:id/children              ✗ (MISSING)
```

### 6. Validation Rules

Both organizations and groups must enforce:

1. **No circular references** - checkCircularReference() must work
2. **Parent must be in same organization** (for groups)
3. **Parent must be active** to create/update child
4. **Cannot delete entity with active children**
5. **Cannot deactivate entity with active children**
6. **Maximum hierarchy depth** (recommended: 10 levels)

### 7. Audit Requirements

All hierarchy changes must be audited:

```go
AuditActionChangeOrganizationHierarchy
AuditActionChangeGroupHierarchy
```

With details:
- Old parent ID
- New parent ID
- Affected entity ID
- Organization context

## Performance Considerations

### Recursive CTE Query Pattern (PostgreSQL)

```sql
-- Get all ancestors (parents)
WITH RECURSIVE hierarchy AS (
    SELECT id, name, parent_id, 0 AS depth
    FROM groups
    WHERE id = $1

    UNION ALL

    SELECT g.id, g.name, g.parent_id, h.depth + 1
    FROM groups g
    JOIN hierarchy h ON g.id = h.parent_id
    WHERE h.depth < 10  -- Prevent infinite loops
)
SELECT * FROM hierarchy WHERE id != $1 ORDER BY depth DESC;

-- Get all descendants (children)
WITH RECURSIVE hierarchy AS (
    SELECT id, name, parent_id, 0 AS depth
    FROM groups
    WHERE id = $1

    UNION ALL

    SELECT g.id, g.name, g.parent_id, h.depth + 1
    FROM groups g
    JOIN hierarchy h ON g.parent_id = h.id
    WHERE h.depth < 10  -- Prevent infinite loops
)
SELECT * FROM hierarchy WHERE id != $1;
```

### Caching Strategy

1. **Cache keys:**
   - `group:hierarchy:{groupID}` - full hierarchy
   - `group:parents:{groupID}` - parent chain
   - `group:children:{groupID}` - immediate children
   - `org:hierarchy:{orgID}` - organization hierarchy

2. **TTL:** 10 minutes for hierarchy data

3. **Invalidation triggers:**
   - Group/org parent_id update
   - Group/org deletion
   - Group/org activation/deactivation

## Edge Cases to Handle

1. **Orphaned groups/orgs** - parent_id references deleted entity
2. **Cross-organization group parents** - must prevent
3. **Circular references** - detect before save
4. **Deep hierarchies** - limit depth, warn on deep structures
5. **Concurrent hierarchy modifications** - use transactions
6. **Soft-deleted parents** - handle gracefully in queries

## Testing Requirements

1. **Unit tests** for each repository method
2. **Integration tests** for hierarchy traversal
3. **Performance tests** for deep hierarchies (100+ nodes)
4. **Circular reference detection tests**
5. **Cache invalidation tests**
6. **Concurrent modification tests**

## Success Criteria

✓ All repository methods implemented and tested
✓ GroupService.GetGroupHierarchy() returns correct hierarchy
✓ OrganizationService.GetOrganizationHierarchy() includes groups
✓ No N+1 query patterns - use CTEs
✓ Cache hit rate > 80% for hierarchy queries
✓ Hierarchy queries < 200ms P95, < 500ms P99
✓ All edge cases handled gracefully
✓ Comprehensive audit logging
✓ API endpoints functional and documented

## Acceptance Criteria

For groups:
- [ ] Can retrieve full hierarchy (parents + children) in single API call
- [ ] Can get only parent chain for breadcrumb display
- [ ] Can get only children for tree expansion
- [ ] Hierarchy changes are cached and invalidated correctly
- [ ] Circular references are prevented
- [ ] Cross-org parent assignments are prevented

For organizations:
- [ ] Can retrieve full hierarchy including sub-organizations
- [ ] Organization hierarchy includes groups within each org
- [ ] GetDescendants works for reporting purposes
- [ ] Performance acceptable for orgs with 100+ sub-orgs

For both:
- [ ] All hierarchy operations complete in < 200ms P95
- [ ] Cache invalidation works correctly
- [ ] Audit logs capture all hierarchy changes
- [ ] Tests cover happy path and edge cases
