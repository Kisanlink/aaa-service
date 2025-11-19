# Hierarchy Depth Limit Validation Implementation

## Overview

Implemented maximum depth constraints for organization and group hierarchies to prevent DoS attacks through arbitrarily deep hierarchies.

## Constants Defined

```go
// Organization Service
const MaxOrganizationHierarchyDepth = 10

// Group Service
const MaxGroupHierarchyDepth = 8
```

## Implementation Details

### 1. Repository Layer

#### OrganizationRepository

**File:** `/Users/kaushik/aaa-service/internal/repositories/organizations/organization_repository.go`

Added `GetHierarchyDepth` method:

```go
func (r *OrganizationRepository) GetHierarchyDepth(ctx context.Context, orgID string, newParentID string) (int, error)
```

**Logic:**
- Uses PostgreSQL recursive CTEs for efficient depth calculation
- Calculates upward depth: from newParent to root
- Calculates downward depth: from org to its deepest child
- Total depth = parent_depth + 1 + subtree_depth
- Includes cycle prevention with depth limit of 20 in CTEs

**SQL Query Structure:**
```sql
WITH parent_depth AS (
    -- Recursive CTE to traverse upward to root
),
subtree_depth AS (
    -- Recursive CTE to traverse downward to leaves
)
SELECT (parent_depth + 1 + subtree_depth) as total_depth
```

#### GroupRepository

**File:** `/Users/kaushik/aaa-service/internal/repositories/groups/group_repository.go`

Added `GetHierarchyDepth` method with identical logic as OrganizationRepository but operating on `groups` table.

### 2. Interface Layer

**File:** `/Users/kaushik/aaa-service/internal/interfaces/interfaces.go`

Updated interfaces:
```go
type OrganizationRepository interface {
    // ... existing methods
    GetHierarchyDepth(ctx context.Context, orgID string, newParentID string) (int, error)
}

type GroupRepository interface {
    // ... existing methods
    GetHierarchyDepth(ctx context.Context, groupID string, newParentID string) (int, error)
}
```

### 3. Adapter Layer

**File:** `/Users/kaushik/aaa-service/internal/repositories/adapters/group_repository_adapter.go`

Added adapter method:
```go
func (a *GroupRepositoryAdapter) GetHierarchyDepth(ctx context.Context, groupID string, newParentID string) (int, error) {
    return a.repo.GetHierarchyDepth(ctx, groupID, newParentID)
}
```

### 4. Service Layer

#### OrganizationService

**File:** `/Users/kaushik/aaa-service/internal/services/organizations/organization_service.go`

**CreateOrganization - Lines 89-103:**
```go
// Validate hierarchy depth
// For new org creation, orgID is empty string since org doesn't exist yet
depth, err := s.orgRepo.GetHierarchyDepth(ctx, "", *req.ParentID)
if err != nil {
    s.logger.Error("Failed to calculate hierarchy depth", zap.Error(err))
    return nil, errors.NewInternalError(err)
}
if depth > MaxOrganizationHierarchyDepth {
    s.logger.Warn("Organization hierarchy depth limit exceeded",
        zap.String("parent_id", *req.ParentID),
        zap.Int("calculated_depth", depth),
        zap.Int("max_depth", MaxOrganizationHierarchyDepth))
    return nil, errors.NewValidationError(fmt.Sprintf("organization hierarchy depth limit (%d levels) exceeded", MaxOrganizationHierarchyDepth))
}
```

**UpdateOrganization - Lines 232-246:**
```go
// Validate hierarchy depth - when moving an org, check if move would exceed depth limit
depth, err := s.orgRepo.GetHierarchyDepth(ctx, orgID, *req.ParentID)
if err != nil {
    s.logger.Error("Failed to calculate hierarchy depth", zap.Error(err))
    return nil, errors.NewInternalError(err)
}
if depth > MaxOrganizationHierarchyDepth {
    s.logger.Warn("Organization hierarchy depth limit exceeded",
        zap.String("org_id", orgID),
        zap.String("new_parent_id", *req.ParentID),
        zap.Int("calculated_depth", depth),
        zap.Int("max_depth", MaxOrganizationHierarchyDepth))
    return nil, errors.NewValidationError(fmt.Sprintf("organization hierarchy depth limit (%d levels) exceeded", MaxOrganizationHierarchyDepth))
}
```

#### GroupService

**File:** `/Users/kaushik/aaa-service/internal/services/groups/group_service.go`

**CreateGroup - Lines 130-144:**
```go
// Validate hierarchy depth
// For new group creation, groupID is empty string since group doesn't exist yet
depth, err := s.groupRepo.GetHierarchyDepth(ctx, "", *createReq.ParentID)
if err != nil {
    s.logger.Error("Failed to calculate hierarchy depth", zap.Error(err))
    return nil, errors.NewInternalError(err)
}
if depth > MaxGroupHierarchyDepth {
    s.logger.Warn("Group hierarchy depth limit exceeded",
        zap.String("parent_id", *createReq.ParentID),
        zap.Int("calculated_depth", depth),
        zap.Int("max_depth", MaxGroupHierarchyDepth))
    return nil, errors.NewValidationError(fmt.Sprintf("group hierarchy depth limit (%d levels) exceeded", MaxGroupHierarchyDepth))
}
```

**UpdateGroup - Lines 278-292:**
```go
// Validate hierarchy depth - when moving a group, check if move would exceed depth limit
depth, err := s.groupRepo.GetHierarchyDepth(ctx, groupID, *updateReq.ParentID)
if err != nil {
    s.logger.Error("Failed to calculate hierarchy depth", zap.Error(err))
    return nil, errors.NewInternalError(err)
}
if depth > MaxGroupHierarchyDepth {
    s.logger.Warn("Group hierarchy depth limit exceeded",
        zap.String("group_id", groupID),
        zap.String("new_parent_id", *updateReq.ParentID),
        zap.Int("calculated_depth", depth),
        zap.Int("max_depth", MaxGroupHierarchyDepth))
    return nil, errors.NewValidationError(fmt.Sprintf("group hierarchy depth limit (%d levels) exceeded", MaxGroupHierarchyDepth))
}
```

## Error Messages

User-friendly error messages returned:
- **Organizations:** `"organization hierarchy depth limit (10 levels) exceeded"`
- **Groups:** `"group hierarchy depth limit (8 levels) exceeded"`

## Security Benefits

1. **DoS Prevention:** Limits computational complexity of hierarchy traversal operations
2. **Resource Protection:** Prevents unbounded recursion in queries
3. **Performance Optimization:** Recursive CTEs with hard limits prevent runaway queries
4. **Predictable Behavior:** Known maximum depth allows for capacity planning

## Performance Characteristics

- **Query Complexity:** O(depth) for each hierarchy check
- **Database Impact:** Single query using efficient recursive CTE
- **Hard Limits:** CTEs limited to depth of 20 to prevent runaway queries
- **Caching Opportunity:** Results could be cached for frequently accessed hierarchies

## Test Coverage

### Required Test Cases

#### Organization Depth Validation
1. Create org at depth 9 (should succeed)
2. Create org at depth 10 (should fail)
3. Move org to exceed depth limit (should fail)
4. Move org within depth limit (should succeed)

#### Group Depth Validation
1. Create group at depth 7 (should succeed)
2. Create group at depth 8 (should fail)
3. Move group to exceed depth limit (should fail)
4. Move group within depth limit (should succeed)

### Edge Cases
1. New organization with no parent (depth = 0)
2. New group with no parent (depth = 0)
3. Organization with existing deep subtree being moved
4. Group with existing deep subtree being moved

## Files Modified

1. `/Users/kaushik/aaa-service/internal/interfaces/interfaces.go`
2. `/Users/kaushik/aaa-service/internal/repositories/adapters/group_repository_adapter.go`
3. `/Users/kaushik/aaa-service/internal/repositories/groups/group_repository.go`
4. `/Users/kaushik/aaa-service/internal/repositories/organizations/organization_repository.go`
5. `/Users/kaushik/aaa-service/internal/services/groups/group_service.go`
6. `/Users/kaushik/aaa-service/internal/services/organizations/organization_service.go`

## Next Steps

1. Add comprehensive unit tests for depth validation
2. Add integration tests for edge cases
3. Consider adding metrics/monitoring for depth limit violations
4. Document API error responses for depth limit exceeded
5. Consider making depth limits configurable via environment variables

## Related Issues

- Addresses Task 2 from `.kiro/specs/hierarchy-architecture/critical-fixes-code.md`
- Part of hierarchy DoS prevention security fixes
- Complements circular reference prevention (already implemented)

## References

- OWASP ASVS - Resource Exhaustion Prevention
- PostgreSQL Recursive CTE documentation
- Project hierarchy security requirements
