# Hierarchy Implementation Quick Reference

## Critical Issues to Fix

### 1. Circular Reference Prevention
**Problem:** Can create org A → B → C → A circular hierarchy
**Solution:** Check ancestors before allowing parent assignment
```go
// Before setting parent, check:
ancestors := GetAllAncestors(newParentID)
for _, ancestor := range ancestors {
    if ancestor.ID == childID {
        return ERROR("Circular reference")
    }
}
```

### 2. Role Inheritance Not Working
**Problem:** Users only get direct roles, not group-inherited roles in JWT
**Current Flow:**
```
LoginV2() → GetUserWithRoles() → GetActiveRolesByUserID() → JWT (direct roles only)
```
**Fixed Flow:**
```
LoginV2() → GetUserWithRoles() → GetActiveRolesByUserID() + GetInheritedRoles() → JWT (all roles)
```

### 3. Inefficient Hierarchy Queries
**Current:** Loop-based traversal in `GetParentHierarchy()`
```go
// BAD: O(n) database calls
for currentID != "" {
    org, err := r.GetByID(ctx, currentID)
    // ... fetch parent ...
}
```

**Solution:** Single recursive CTE query
```sql
WITH RECURSIVE ancestors AS (
    SELECT * FROM organizations WHERE id = $1
    UNION ALL
    SELECT o.* FROM organizations o
    JOIN ancestors a ON o.id = a.parent_id
)
SELECT * FROM ancestors;
```

## Key Files to Modify

### Models
- `/internal/entities/models/organization.go` - Add hierarchy fields
- `/internal/entities/models/group.go` - Add hierarchy fields

### Repositories
- `/internal/repositories/organizations/organization_repository.go` - Add CTE queries
- `/internal/repositories/groups/group_repository.go` - Add CTE queries

### Services
- `/internal/services/organizations/organization_service.go` - Add validation
- `/internal/services/groups/group_service.go` - Expose inherited roles
- `/internal/services/user/additional_methods.go` - Include inherited roles

### Handlers
- `/internal/handlers/auth/auth_handler.go` - Use effective roles in token
- `/internal/handlers/organizations/organization_handler.go` - Add hierarchy endpoints

## Database Changes Required

```sql
-- Organizations
ALTER TABLE organizations ADD COLUMN hierarchy_depth INTEGER DEFAULT 0;
ALTER TABLE organizations ADD COLUMN hierarchy_path TEXT;
ALTER TABLE organizations ADD CONSTRAINT check_org_depth CHECK (hierarchy_depth <= 10);

-- Groups
ALTER TABLE groups ADD COLUMN hierarchy_depth INTEGER DEFAULT 0;
ALTER TABLE groups ADD COLUMN hierarchy_path TEXT;
ALTER TABLE groups ADD CONSTRAINT check_group_depth CHECK (hierarchy_depth <= 8);

-- Critical Indexes
CREATE INDEX idx_org_parent_active ON organizations(parent_id, is_active);
CREATE INDEX idx_group_org_parent ON groups(organization_id, parent_id);
```

## Implementation Checklist

### Phase 1: Security Fixes (CRITICAL)
- [ ] Add `WouldCreateCircularReference()` check
- [ ] Add depth validation (max 10 for orgs, 8 for groups)
- [ ] Add database constraints
- [ ] Update models with hierarchy fields

### Phase 2: Role Inheritance (CRITICAL)
- [ ] Modify `GetUserWithRoles()` to include inherited roles
- [ ] Update `LoginV2()` to use effective roles
- [ ] Test token contains all roles
- [ ] Add `/users/{id}/effective-roles` endpoint

### Phase 3: Performance (HIGH)
- [ ] Replace loop queries with recursive CTEs
- [ ] Add hierarchy path materialization
- [ ] Implement hierarchy caching
- [ ] Add proper indexes

### Phase 4: API Completion (MEDIUM)
- [ ] Add `/organizations/{id}/hierarchy` endpoint
- [ ] Add `/organizations/{id}/ancestors` endpoint
- [ ] Add `/organizations/{id}/descendants` endpoint
- [ ] Add `/organizations/{id}/move` endpoint
- [ ] Repeat for groups

## Common Pitfalls to Avoid

### 1. Forgetting Transaction Boundaries
```go
// BAD: No transaction
org.ParentID = newParentID
r.Update(org)
r.UpdateDescendantPaths() // Could fail, leaving inconsistent state

// GOOD: Use transaction
tx := r.BeginTransaction()
org.ParentID = newParentID
tx.Update(org)
tx.UpdateDescendantPaths()
tx.Commit()
```

### 2. Not Invalidating Cache
```go
// After hierarchy change:
cache.Delete(fmt.Sprintf("org:%s:hierarchy", orgID))
cache.Delete(fmt.Sprintf("org:%s:ancestors", orgID))
cache.Delete(fmt.Sprintf("org:%s:descendants", orgID))
// Also invalidate for old and new parents!
```

### 3. Ignoring Depth on Move
```go
// Must check: newParentDepth + movedSubtreeDepth <= MAX_DEPTH
```

### 4. Not Handling Null Parents
```go
// Root organizations have parent_id = NULL
// Must handle in queries:
WHERE parent_id IS NULL  // for roots
WHERE parent_id = $1      // for children
```

## Testing Scenarios

### Circular Reference Test
```go
// Create: A → B → C
// Try: C.parent = A (should fail)
// Try: B.parent = C (should fail)
// Try: A.parent = A (should fail)
```

### Role Inheritance Test
```go
// Setup:
// - User U in Group G1
// - G1 has Role R1
// - G1 child of G2
// - G2 has Role R2

// Verify:
// - U.GetEffectiveRoles() returns [R1, R2]
// - JWT token contains both roles
```

### Performance Test
```go
// Create hierarchy 10 levels deep, 100 children each level
// Measure:
// - GetHierarchy() < 100ms
// - GetAllDescendants() < 200ms
// - MoveOrganization() < 500ms
```

## SQL Queries for Verification

### Check for Circular References
```sql
WITH RECURSIVE check_circular AS (
    SELECT id, parent_id, id::text as path, 0 as depth
    FROM organizations
    WHERE id = 'START_ORG_ID'

    UNION ALL

    SELECT o.id, o.parent_id,
           c.path || '→' || o.id::text,
           c.depth + 1
    FROM organizations o
    JOIN check_circular c ON o.parent_id = c.id
    WHERE c.depth < 20
)
SELECT * FROM check_circular
WHERE id IN (SELECT unnest(string_to_array(path, '→')))
  AND depth > 0;
```

### Get Effective Roles for User
```sql
-- Direct roles
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = 'USER_ID' AND ur.is_active = true

UNION

-- Inherited from groups
SELECT DISTINCT r.* FROM roles r
JOIN group_roles gr ON r.id = gr.role_id
JOIN group_memberships gm ON gr.group_id = gm.group_id
WHERE gm.principal_id = 'USER_ID'
  AND gm.principal_type = 'user'
  AND gm.is_active = true
  AND gr.is_active = true;
```

### Find Orphaned Groups
```sql
SELECT g.* FROM groups g
LEFT JOIN organizations o ON g.organization_id = o.id
WHERE o.id IS NULL OR o.deleted_at IS NOT NULL;
```

## Monitoring Queries

### Hierarchy Depth Distribution
```sql
SELECT hierarchy_depth, COUNT(*) as count
FROM organizations
WHERE deleted_at IS NULL
GROUP BY hierarchy_depth
ORDER BY hierarchy_depth;
```

### Large Hierarchies
```sql
WITH descendant_counts AS (
    SELECT parent_id, COUNT(*) as child_count
    FROM organizations
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL
    GROUP BY parent_id
)
SELECT o.id, o.name, dc.child_count
FROM organizations o
JOIN descendant_counts dc ON o.id = dc.parent_id
WHERE dc.child_count > 100
ORDER BY dc.child_count DESC;
```

## Emergency Fixes

### Break Circular Reference
```sql
-- If circular reference exists, break it
UPDATE organizations
SET parent_id = NULL,
    hierarchy_depth = 0,
    hierarchy_path = '/' || id || '/'
WHERE id = 'PROBLEMATIC_ORG_ID';
```

### Recalculate All Depths
```sql
-- Reset and recalculate
UPDATE organizations SET hierarchy_depth = 0;

-- Then run application logic to recalculate
```

### Force Cache Clear
```bash
# Clear all hierarchy caches
redis-cli --scan --pattern "org:*:hierarchy*" | xargs redis-cli del
redis-cli --scan --pattern "group:*:hierarchy*" | xargs redis-cli del
```

## Contact for Questions

- Architecture: Review ADR-001 in `.kiro/specs/hierarchy-architecture/`
- Implementation: Follow tasks in `implementation-plan.md`
- Testing: Use scenarios in this quick reference
- Issues: Check logs for "hierarchy" or "circular" errors
