# Organization and Group Hierarchy Implementation Design

## Overview

This document details the technical design for implementing complete hierarchy support for organizations and groups in the AAA service.

## Architecture

### Component Layers

```
┌─────────────────────────────────────────────────────┐
│              API/Handler Layer                       │
│  - GET /api/v1/groups/:id/hierarchy                 │
│  - GET /api/v1/organizations/:id/hierarchy          │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│              Service Layer                           │
│  - GroupService.GetGroupHierarchy()                 │
│  - OrganizationService.GetOrganizationHierarchy()   │
│  - Cache management & invalidation                   │
│  - Business validation                               │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│           Repository Layer                           │
│  - GroupRepository (hierarchy methods)               │
│  - OrganizationRepository (optimized CTEs)          │
│  - Raw SQL for recursive queries                     │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│              Database (PostgreSQL)                   │
│  - organizations table with parent_id                │
│  - groups table with parent_id + organization_id     │
│  - Recursive CTE queries                             │
│  - Partial indexes on parent_id                      │
└─────────────────────────────────────────────────────┘
```

## Database Layer

### Schema Review

Both tables correctly have:
- `parent_id` column with self-referencing foreign key
- Partial indexes on `parent_id WHERE deleted_at IS NULL AND parent_id IS NOT NULL`
- Proper cascading behavior

### Recursive CTE Queries

#### Get Parent Hierarchy (Ancestors)

```sql
-- Get all ancestors for a group
WITH RECURSIVE parent_hierarchy AS (
    -- Base case: start with the group itself
    SELECT
        id,
        name,
        description,
        organization_id,
        parent_id,
        is_active,
        created_at,
        updated_at,
        0 AS depth,
        ARRAY[id] AS path  -- Track path to detect cycles
    FROM groups
    WHERE id = $1 AND deleted_at IS NULL

    UNION ALL

    -- Recursive case: get parent of current group
    SELECT
        g.id,
        g.name,
        g.description,
        g.organization_id,
        g.parent_id,
        g.is_active,
        g.created_at,
        g.updated_at,
        ph.depth + 1 AS depth,
        ph.path || g.id AS path  -- Append to path
    FROM groups g
    JOIN parent_hierarchy ph ON g.id = ph.parent_id
    WHERE
        g.deleted_at IS NULL
        AND ph.depth < 10  -- Safety limit
        AND NOT (g.id = ANY(ph.path))  -- Cycle detection
)
SELECT
    id,
    name,
    description,
    organization_id,
    parent_id,
    is_active,
    created_at,
    updated_at,
    depth
FROM parent_hierarchy
WHERE id != $1  -- Exclude the starting group
ORDER BY depth DESC;  -- Root first, then down to immediate parent
```

#### Get Descendant Hierarchy (Children)

```sql
-- Get all descendants for a group
WITH RECURSIVE child_hierarchy AS (
    -- Base case: start with the group itself
    SELECT
        id,
        name,
        description,
        organization_id,
        parent_id,
        is_active,
        created_at,
        updated_at,
        0 AS depth,
        ARRAY[id] AS path
    FROM groups
    WHERE id = $1 AND deleted_at IS NULL

    UNION ALL

    -- Recursive case: get children of current group
    SELECT
        g.id,
        g.name,
        g.description,
        g.organization_id,
        g.parent_id,
        g.is_active,
        g.created_at,
        g.updated_at,
        ch.depth + 1 AS depth,
        ch.path || g.id AS path
    FROM groups g
    JOIN child_hierarchy ch ON g.parent_id = ch.id
    WHERE
        g.deleted_at IS NULL
        AND ch.depth < 10
        AND NOT (g.id = ANY(ch.path))
)
SELECT
    id,
    name,
    description,
    organization_id,
    parent_id,
    is_active,
    created_at,
    updated_at,
    depth
FROM child_hierarchy
WHERE id != $1
ORDER BY depth ASC, name ASC;  -- Breadth-first ordering
```

### Query Performance

**Expected performance with proper indexes:**
- Hierarchy depth 1-3 levels: < 10ms
- Hierarchy depth 4-6 levels: < 30ms
- Hierarchy depth 7-10 levels: < 100ms

**Indexes required:** ✓ Already exist
- `idx_groups_parent` on `(parent_id) WHERE deleted_at IS NULL AND parent_id IS NOT NULL`
- `idx_organizations_parent` on `(parent_id) WHERE deleted_at IS NULL AND parent_id IS NOT NULL`

## Repository Layer Implementation

### GroupRepository New Methods

```go
// GetParentHierarchy retrieves the complete parent hierarchy for a group
// Returns ordered list from root to immediate parent
func (r *GroupRepository) GetParentHierarchy(ctx context.Context, groupID string) ([]*models.Group, error) {
    query := `
        WITH RECURSIVE parent_hierarchy AS (
            SELECT id, name, description, organization_id, parent_id, is_active,
                   created_at, updated_at, created_by, updated_by,
                   0 AS depth, ARRAY[id] AS path
            FROM groups
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT g.id, g.name, g.description, g.organization_id, g.parent_id, g.is_active,
                   g.created_at, g.updated_at, g.created_by, g.updated_by,
                   ph.depth + 1 AS depth, ph.path || g.id AS path
            FROM groups g
            JOIN parent_hierarchy ph ON g.id = ph.parent_id
            WHERE g.deleted_at IS NULL
              AND ph.depth < 10
              AND NOT (g.id = ANY(ph.path))
        )
        SELECT id, name, description, organization_id, parent_id, is_active,
               created_at, updated_at, created_by, updated_by
        FROM parent_hierarchy
        WHERE id != $1
        ORDER BY depth DESC
    `

    // Execute using kisanlink-db
    var groups []*models.Group
    db := r.dbManager.GetDB()
    err := db.WithContext(ctx).Raw(query, groupID).Scan(&groups).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get parent hierarchy: %w", err)
    }

    return groups, nil
}

// GetDescendants retrieves all descendant groups recursively
func (r *GroupRepository) GetDescendants(ctx context.Context, groupID string) ([]*models.Group, error) {
    query := `
        WITH RECURSIVE child_hierarchy AS (
            SELECT id, name, description, organization_id, parent_id, is_active,
                   created_at, updated_at, created_by, updated_by,
                   0 AS depth, ARRAY[id] AS path
            FROM groups
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT g.id, g.name, g.description, g.organization_id, g.parent_id, g.is_active,
                   g.created_at, g.updated_at, g.created_by, g.updated_by,
                   ch.depth + 1 AS depth, ch.path || g.id AS path
            FROM groups g
            JOIN child_hierarchy ch ON g.parent_id = ch.id
            WHERE g.deleted_at IS NULL
              AND ch.depth < 10
              AND NOT (g.id = ANY(ch.path))
        )
        SELECT id, name, description, organization_id, parent_id, is_active,
               created_at, updated_at, created_by, updated_by
        FROM child_hierarchy
        WHERE id != $1
        ORDER BY depth ASC, name ASC
    `

    var groups []*models.Group
    db := r.dbManager.GetDB()
    err := db.WithContext(ctx).Raw(query, groupID).Scan(&groups).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get descendants: %w", err)
    }

    return groups, nil
}

// GetActiveChildren retrieves only active child groups (immediate children)
func (r *GroupRepository) GetActiveChildren(ctx context.Context, parentID string) ([]*models.Group, error) {
    filter := base.NewFilterBuilder().
        Where("parent_id", base.OpEqual, parentID).
        Where("is_active", base.OpEqual, true).
        Build()

    return r.BaseFilterableRepository.Find(ctx, filter)
}
```

### OrganizationRepository Optimization

```go
// GetParentHierarchy - optimized version using CTE
func (r *OrganizationRepository) GetParentHierarchy(ctx context.Context, orgID string) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE parent_hierarchy AS (
            SELECT id, name, type, description, parent_id, is_active,
                   created_at, updated_at, created_by, updated_by,
                   0 AS depth, ARRAY[id] AS path
            FROM organizations
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT o.id, o.name, o.type, o.description, o.parent_id, o.is_active,
                   o.created_at, o.updated_at, o.created_by, o.updated_by,
                   ph.depth + 1 AS depth, ph.path || o.id AS path
            FROM organizations o
            JOIN parent_hierarchy ph ON o.id = ph.parent_id
            WHERE o.deleted_at IS NULL
              AND ph.depth < 10
              AND NOT (o.id = ANY(ph.path))
        )
        SELECT id, name, type, description, parent_id, is_active,
               created_at, updated_at, created_by, updated_by
        FROM parent_hierarchy
        WHERE id != $1
        ORDER BY depth DESC
    `

    var orgs []*models.Organization
    db := r.dbManager.GetDB()
    err := db.WithContext(ctx).Raw(query, orgID).Scan(&orgs).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get parent hierarchy: %w", err)
    }

    return orgs, nil
}

// GetDescendants retrieves all descendant organizations recursively
func (r *OrganizationRepository) GetDescendants(ctx context.Context, orgID string) ([]*models.Organization, error) {
    // Similar CTE implementation as GroupRepository.GetDescendants
    // ... (same pattern)
}
```

## Service Layer Implementation

### GroupService.GetGroupHierarchy

```go
// GetGroupHierarchy retrieves the complete hierarchy for a group including roles
func (s *Service) GetGroupHierarchy(ctx context.Context, groupID string) (*groupResponses.GroupHierarchyResponse, error) {
    s.logger.Info("Retrieving group hierarchy", zap.String("group_id", groupID))

    // Check cache first
    cacheKey := fmt.Sprintf("group:hierarchy:%s", groupID)
    if cached, found := s.groupCache.Get(ctx, cacheKey); found {
        s.logger.Debug("Returning cached group hierarchy")
        return cached.(*groupResponses.GroupHierarchyResponse), nil
    }

    // Get the group
    group, err := s.groupRepo.GetByID(ctx, groupID)
    if err != nil || group == nil {
        return nil, errors.NewNotFoundError("group not found")
    }

    // Get parent hierarchy using CTE
    parents, err := s.groupRepo.GetParentHierarchy(ctx, groupID)
    if err != nil {
        s.logger.Error("Failed to get parent hierarchy", zap.Error(err))
        return nil, errors.NewInternalError(err)
    }

    // Get immediate children
    children, err := s.groupRepo.GetChildren(ctx, groupID)
    if err != nil {
        s.logger.Error("Failed to get children", zap.Error(err))
        return nil, errors.NewInternalError(err)
    }

    // Get roles for this group
    roles, err := s.GetGroupRoles(ctx, groupID)
    if err != nil {
        s.logger.Warn("Failed to get group roles", zap.Error(err))
        roles = []*groupResponses.GroupRoleDetail{}
    }

    // Build response
    response := &groupResponses.GroupHierarchyResponse{
        Group:    convertGroupToResponse(group),
        Parents:  convertGroupsToResponses(parents),
        Children: convertGroupsToResponses(children),
        Roles:    roles.([]*groupResponses.GroupRoleDetail),
    }

    // Cache the response (10 minute TTL)
    s.groupCache.Set(ctx, cacheKey, response, 10*time.Minute)

    return response, nil
}
```

### Cache Invalidation

```go
// InvalidateHierarchyCache invalidates all hierarchy-related caches for a group
func (s *Service) InvalidateHierarchyCache(ctx context.Context, groupID string) {
    // Get the group to find organization
    group, err := s.groupRepo.GetByID(ctx, groupID)
    if err != nil {
        s.logger.Warn("Failed to get group for cache invalidation", zap.Error(err))
        return
    }

    // Invalidate this group's hierarchy cache
    s.groupCache.Delete(ctx, fmt.Sprintf("group:hierarchy:%s", groupID))

    // Invalidate parent's children cache if has parent
    if group.ParentID != nil && *group.ParentID != "" {
        s.groupCache.Delete(ctx, fmt.Sprintf("group:children:%s", *group.ParentID))
    }

    // Invalidate organization hierarchy cache (includes groups)
    s.groupCache.Delete(ctx, fmt.Sprintf("org:hierarchy:%s", group.OrganizationID))

    // Invalidate all ancestors (they show this in their children)
    parents, err := s.groupRepo.GetParentHierarchy(ctx, groupID)
    if err == nil {
        for _, parent := range parents {
            s.groupCache.Delete(ctx, fmt.Sprintf("group:hierarchy:%s", parent.ID))
        }
    }
}
```

## Response Models

### GroupHierarchyResponse

```go
package groups

import "time"

// GroupHierarchyResponse represents a complete hierarchy view of a group
type GroupHierarchyResponse struct {
    Group    *GroupResponse     `json:"group"`
    Parents  []*GroupResponse   `json:"parents"`   // Ordered from root to immediate parent
    Children []*GroupResponse   `json:"children"`  // Immediate children only
    Roles    []*GroupRoleDetail `json:"roles"`     // Roles assigned to this group
}

// GroupWithParentsResponse represents a group with its ancestor chain
type GroupWithParentsResponse struct {
    Group   *GroupResponse   `json:"group"`
    Parents []*GroupResponse `json:"parents"` // Ordered from root to immediate parent
    Depth   int              `json:"depth"`   // Depth in hierarchy (0 = root)
}

// GroupWithChildrenResponse represents a group with its descendant tree
type GroupWithChildrenResponse struct {
    Group    *GroupResponse        `json:"group"`
    Children []*GroupHierarchyNode `json:"children"` // Recursive children tree
    Count    int                   `json:"count"`    // Total descendant count
}

// GroupHierarchyNode represents a node in the group tree
type GroupHierarchyNode struct {
    Group    *GroupResponse        `json:"group"`
    Roles    []*GroupRoleDetail    `json:"roles,omitempty"`
    Children []*GroupHierarchyNode `json:"children,omitempty"`
}
```

## API Routes

### New Routes to Add

```go
// In internal/routes/group_routes.go

// Add under the groups router
groups.GET("/:id/hierarchy", authMiddleware.RequirePermission("group", "read"), groupHandler.GetHierarchy)
groups.GET("/:id/parents", authMiddleware.RequirePermission("group", "read"), groupHandler.GetParents)
groups.GET("/:id/children", authMiddleware.RequirePermission("group", "read"), groupHandler.GetChildren)
```

### Handler Implementation

```go
// internal/handlers/groups/hierarchy.go

package groups

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// GetHierarchy handles GET /api/v1/groups/:id/hierarchy
func (h *GroupHandler) GetHierarchy(c *gin.Context) {
    groupID := c.Param("id")

    hierarchy, err := h.groupService.GetGroupHierarchy(c.Request.Context(), groupID)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, hierarchy)
}

// GetParents handles GET /api/v1/groups/:id/parents
func (h *GroupHandler) GetParents(c *gin.Context) {
    groupID := c.Param("id")

    parents, err := h.groupService.GetGroupWithParents(c.Request.Context(), groupID)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, parents)
}

// GetChildren handles GET /api/v1/groups/:id/children
func (h *GroupHandler) GetChildren(c *gin.Context) {
    groupID := c.Param("id")
    recursive := c.DefaultQuery("recursive", "false") == "true"

    var result interface{}
    var err error

    if recursive {
        result, err = h.groupService.GetGroupWithChildren(c.Request.Context(), groupID)
    } else {
        result, err = h.groupService.GetGroupChildren(c.Request.Context(), groupID)
    }

    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, result)
}
```

## Testing Strategy

### Unit Tests

```go
// internal/repositories/groups/group_repository_test.go

func TestGroupRepository_GetParentHierarchy(t *testing.T) {
    // Test cases:
    // 1. Single level (group with one parent)
    // 2. Multi-level (group with 3+ ancestors)
    // 3. Root group (no parents)
    // 4. Orphaned group (parent deleted)
    // 5. Performance with deep hierarchy (10 levels)
}

func TestGroupRepository_GetDescendants(t *testing.T) {
    // Test cases:
    // 1. Leaf group (no children)
    // 2. Single level children
    // 3. Multi-level descendants (3+ generations)
    // 4. Wide tree (many siblings at each level)
    // 5. Cycle detection (shouldn't happen but test safety)
}
```

### Integration Tests

```go
// internal/services/groups/hierarchy_test.go

func TestGroupService_GetGroupHierarchy(t *testing.T) {
    // Test complete hierarchy retrieval with roles
    // Verify caching works
    // Test cache invalidation on updates
}

func TestGroupService_CircularReferenceDetection(t *testing.T) {
    // Attempt to create circular hierarchy
    // Verify it's rejected
}
```

### Performance Tests

```go
// internal/repositories/groups/performance_test.go

func BenchmarkGetParentHierarchy(b *testing.B) {
    // Benchmark with different hierarchy depths
    // Depths: 3, 5, 10 levels
}

func BenchmarkGetDescendants(b *testing.B) {
    // Benchmark with different tree sizes
    // Sizes: 10, 50, 100, 500 nodes
}
```

## Error Handling

### Circular Reference Detection

```go
func (s *Service) checkCircularReference(ctx context.Context, groupID, newParentID string) error {
    // Get all ancestors of the new parent
    ancestors, err := s.groupRepo.GetParentHierarchy(ctx, newParentID)
    if err != nil {
        return fmt.Errorf("failed to check ancestors: %w", err)
    }

    // Check if groupID appears in ancestors
    for _, ancestor := range ancestors {
        if ancestor.ID == groupID {
            return fmt.Errorf("circular reference: group %s would be its own ancestor", groupID)
        }
    }

    return nil
}
```

### Orphaned Entity Handling

```go
// When getting hierarchy, handle case where parent_id references deleted entity
// The CTE query with "deleted_at IS NULL" will naturally exclude soft-deleted parents
// No special handling needed - just return partial hierarchy
```

## Monitoring and Observability

### Metrics to Track

```go
// Hierarchy query performance
hierarchyQueryDuration.WithLabelValues("group", "parents").Observe(duration)
hierarchyQueryDuration.WithLabelValues("group", "children").Observe(duration)
hierarchyQueryDuration.WithLabelValues("organization", "parents").Observe(duration)

// Cache metrics
cacheHits.WithLabelValues("group_hierarchy").Inc()
cacheMisses.WithLabelValues("group_hierarchy").Inc()

// Hierarchy depth
hierarchyDepth.WithLabelValues("group").Observe(float64(depth))
hierarchyDepth.WithLabelValues("organization").Observe(float64(depth))
```

### Logging

```go
// Log hierarchy operations with context
logger.Info("Retrieved group hierarchy",
    zap.String("group_id", groupID),
    zap.Int("parent_count", len(parents)),
    zap.Int("children_count", len(children)),
    zap.Duration("duration", duration))
```

## Rollout Plan

### Phase 1: Repository Layer (Week 1)
- [ ] Implement GroupRepository.GetParentHierarchy()
- [ ] Implement GroupRepository.GetDescendants()
- [ ] Implement GroupRepository.GetActiveChildren()
- [ ] Optimize OrganizationRepository.GetParentHierarchy() to use CTE
- [ ] Implement OrganizationRepository.GetDescendants()
- [ ] Unit tests for all new methods
- [ ] Performance benchmarks

### Phase 2: Service Layer (Week 1)
- [ ] Implement GroupService.GetGroupHierarchy()
- [ ] Implement GroupService.GetGroupWithParents()
- [ ] Implement GroupService.GetGroupWithChildren()
- [ ] Add cache invalidation logic
- [ ] Integration tests
- [ ] Update circular reference detection to use new methods

### Phase 3: API Layer (Week 2)
- [ ] Add new routes for group hierarchy
- [ ] Implement handlers
- [ ] Add Swagger documentation
- [ ] API integration tests
- [ ] Update existing group update logic to invalidate caches

### Phase 4: Testing & Documentation (Week 2)
- [ ] End-to-end tests with realistic data
- [ ] Performance testing with large hierarchies
- [ ] Load testing
- [ ] Update API documentation
- [ ] Update developer guides

### Phase 5: Deployment (Week 3)
- [ ] Deploy to staging
- [ ] Run migration to validate existing data
- [ ] Performance validation
- [ ] Deploy to production
- [ ] Monitor metrics

## Backward Compatibility

All changes are backward compatible:
- New methods added, no existing methods modified (except optimization)
- New API endpoints added, existing endpoints unchanged
- Database schema unchanged (already correct)
- Response models extended, no breaking changes

## Risks and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance degradation with deep hierarchies | High | CTE queries, caching, depth limits |
| Cache invalidation bugs | Medium | Comprehensive tests, conservative invalidation |
| Circular reference creation | High | Validation before save, CTE cycle detection |
| Memory issues with large trees | Medium | Pagination, depth limits, streaming responses |
| Concurrent modifications | Medium | Database transactions, optimistic locking |

## Success Metrics

After implementation, measure:
- **Query performance:** P95 < 100ms for hierarchies up to 5 levels
- **Cache hit rate:** > 80% for hierarchy queries
- **Error rate:** < 0.1% for hierarchy operations
- **API adoption:** 50%+ of group queries use new hierarchy endpoints
- **User satisfaction:** Positive feedback on hierarchy navigation
