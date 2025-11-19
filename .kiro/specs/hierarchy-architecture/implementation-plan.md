# Organization and Group Hierarchy Implementation Plan

**Date:** 2025-11-17
**Priority:** CRITICAL
**Estimated Timeline:** 4 weeks

## Current Issues Summary

1. **No Hierarchy Traversal**: Missing efficient methods to navigate organization/group trees
2. **No Circular Reference Protection**: Can create invalid circular hierarchies
3. **No Depth Limits**: Unbounded hierarchy depth can cause performance issues
4. **Role Inheritance Not Used**: Implemented but not integrated with authentication
5. **Inefficient Queries**: Loop-based traversal instead of recursive CTEs
6. **Missing APIs**: No endpoints for hierarchy management
7. **No Caching**: Hierarchy queries hit database every time

## Phase 1: Critical Security & Data Integrity Fixes (3 days)

### Task 1.1: Add Circular Reference Validation
**File:** `/Users/kaushik/aaa-service/internal/repositories/organizations/organization_repository.go`
```go
// Add method to check for circular references
func (r *OrganizationRepository) WouldCreateCircularReference(ctx context.Context, childID, parentID string) (bool, error)
```

**File:** `/Users/kaushik/aaa-service/internal/repositories/groups/group_repository.go`
```go
// Add method to check for circular references
func (r *GroupRepository) WouldCreateCircularReference(ctx context.Context, childID, parentID string) (bool, error)
```

### Task 1.2: Add Hierarchy Depth Validation
**File:** `/Users/kaushik/aaa-service/internal/services/organizations/organization_service.go`
- Add `ValidateHierarchyDepth()` method
- Check depth on create/update operations
- Maximum depth: 10 for organizations, 8 for groups

### Task 1.3: Database Migration for Hierarchy Fields
**File:** `/Users/kaushik/aaa-service/migrations/20251117_add_hierarchy_fields.sql`
```sql
-- Add hierarchy tracking fields
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS hierarchy_path TEXT;

ALTER TABLE groups ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS hierarchy_path TEXT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS root_group_id VARCHAR(255);

-- Add constraints
ALTER TABLE organizations ADD CONSTRAINT check_org_depth CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 10);
ALTER TABLE groups ADD CONSTRAINT check_group_depth CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 8);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_org_hierarchy_path ON organizations USING btree(hierarchy_path text_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_org_parent_active ON organizations(parent_id, is_active);
CREATE INDEX IF NOT EXISTS idx_org_depth ON organizations(hierarchy_depth);

CREATE INDEX IF NOT EXISTS idx_group_hierarchy_path ON groups USING btree(hierarchy_path text_pattern_ops);
CREATE INDEX IF NOT EXISTS idx_group_org_parent ON groups(organization_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_group_root ON groups(root_group_id) WHERE root_group_id IS NOT NULL;
```

### Task 1.4: Update Models with Hierarchy Fields
**File:** `/Users/kaushik/aaa-service/internal/entities/models/organization.go`
- Add `HierarchyDepth`, `HierarchyPath` fields
- Add validation methods

**File:** `/Users/kaushik/aaa-service/internal/entities/models/group.go`
- Add `HierarchyDepth`, `HierarchyPath`, `RootGroupID` fields
- Add validation methods

## Phase 2: Role Inheritance Integration (3 days)

### Task 2.1: Integrate Role Inheritance with Token Generation
**File:** `/Users/kaushik/aaa-service/internal/services/user/additional_methods.go`
```go
// Modify GetUserWithRoles to include inherited roles
func (s *UserService) GetUserWithRoles(ctx context.Context, userID string) (*responses.UserResponse, error) {
    // Existing code for direct roles...

    // NEW: Add inherited roles from groups
    if orgID != "" {
        inheritedRoles, err := s.groupService.GetInheritedRoles(ctx, userID, orgID)
        if err == nil {
            userResponse.Roles = s.mergeRoles(userResponse.Roles, inheritedRoles)
        }
    }
}
```

### Task 2.2: Add Service Method for Effective Roles
**File:** `/Users/kaushik/aaa-service/internal/services/groups/group_service.go`
```go
// Add method to get all effective roles for a user
func (s *Service) GetEffectiveRoles(ctx context.Context, userID, orgID string) ([]*models.Role, error) {
    return s.roleInheritanceEngine.CalculateEffectiveRoles(ctx, userID, orgID)
}
```

### Task 2.3: Update Auth Handler to Use Effective Roles
**File:** `/Users/kaushik/aaa-service/internal/handlers/auth/auth_handler.go`
- Modify `LoginV2()` to call `GetEffectiveRoles()`
- Include inherited roles in JWT token

### Task 2.4: Add Effective Roles Endpoint
**File:** `/Users/kaushik/aaa-service/internal/handlers/users/user_handler.go`
```go
// GET /api/v1/users/{id}/effective-roles
func (h *Handler) GetEffectiveRoles(c *gin.Context)
```

## Phase 3: Efficient Hierarchy Queries (4 days)

### Task 3.1: Implement Recursive CTE Queries for Organizations
**File:** `/Users/kaushik/aaa-service/internal/repositories/organizations/hierarchy_queries.go` (NEW)
```go
package organizations

// GetOrganizationTree - Get full hierarchy tree using recursive CTE
func (r *OrganizationRepository) GetOrganizationTree(ctx context.Context, rootID string, maxDepth int) ([]*models.Organization, error)

// GetAllAncestors - Get all ancestors using recursive CTE
func (r *OrganizationRepository) GetAllAncestors(ctx context.Context, orgID string) ([]*models.Organization, error)

// GetAllDescendants - Get all descendants using recursive CTE
func (r *OrganizationRepository) GetAllDescendants(ctx context.Context, orgID string, maxDepth int) ([]*models.Organization, error)
```

### Task 3.2: Implement Recursive CTE Queries for Groups
**File:** `/Users/kaushik/aaa-service/internal/repositories/groups/hierarchy_queries.go` (NEW)
```go
package groups

// Similar methods for groups
func (r *GroupRepository) GetGroupTree(ctx context.Context, rootID string, maxDepth int) ([]*models.Group, error)
func (r *GroupRepository) GetAllAncestors(ctx context.Context, groupID string) ([]*models.Group, error)
func (r *GroupRepository) GetAllDescendants(ctx context.Context, groupID string, maxDepth int) ([]*models.Group, error)
```

### Task 3.3: Add Materialized Path Management
**File:** `/Users/kaushik/aaa-service/internal/services/organizations/hierarchy_manager.go` (NEW)
```go
package organizations

type HierarchyManager struct {
    orgRepo interfaces.OrganizationRepository
    logger  *zap.Logger
}

// UpdateHierarchyPath - Recalculate and update hierarchy paths
func (m *HierarchyManager) UpdateHierarchyPath(ctx context.Context, orgID string) error

// UpdateDescendantPaths - Update paths for all descendants when parent moves
func (m *HierarchyManager) UpdateDescendantPaths(ctx context.Context, orgID string, newBasePath string) error
```

### Task 3.4: Create Hierarchy Service
**File:** `/Users/kaushik/aaa-service/internal/services/hierarchy/service.go` (NEW)
- Centralized service for hierarchy operations
- Handles both organizations and groups
- Manages cache invalidation

## Phase 4: API Endpoints (3 days)

### Task 4.1: Organization Hierarchy Endpoints
**File:** `/Users/kaushik/aaa-service/internal/handlers/organizations/hierarchy_handler.go` (NEW)
```go
// GET /api/v1/organizations/{id}/hierarchy
func (h *Handler) GetOrganizationHierarchy(c *gin.Context)

// GET /api/v1/organizations/{id}/ancestors
func (h *Handler) GetOrganizationAncestors(c *gin.Context)

// GET /api/v1/organizations/{id}/descendants
func (h *Handler) GetOrganizationDescendants(c *gin.Context)

// POST /api/v1/organizations/{id}/move
func (h *Handler) MoveOrganization(c *gin.Context)
```

### Task 4.2: Group Hierarchy Endpoints
**File:** `/Users/kaushik/aaa-service/internal/handlers/groups/hierarchy_handler.go` (NEW)
```go
// Similar endpoints for groups
func (h *Handler) GetGroupHierarchy(c *gin.Context)
func (h *Handler) GetGroupAncestors(c *gin.Context)
func (h *Handler) GetGroupDescendants(c *gin.Context)
func (h *Handler) MoveGroup(c *gin.Context)
```

### Task 4.3: Update Routes
**File:** `/Users/kaushik/aaa-service/internal/routes/organization_routes.go`
- Add new hierarchy endpoints

**File:** `/Users/kaushik/aaa-service/internal/routes/group_routes.go`
- Add new hierarchy endpoints

### Task 4.4: OpenAPI Documentation
**File:** `/Users/kaushik/aaa-service/docs/swagger.yaml`
- Document all new endpoints
- Add request/response schemas

## Phase 5: Caching & Performance (3 days)

### Task 5.1: Implement Hierarchy Cache Service
**File:** `/Users/kaushik/aaa-service/internal/services/cache/hierarchy_cache.go` (NEW)
```go
package cache

type HierarchyCacheService struct {
    cache  interfaces.CacheService
    logger *zap.Logger
}

// Cache hierarchy queries with TTL
func (s *HierarchyCacheService) GetCachedHierarchy(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error)

// Invalidate related caches on hierarchy change
func (s *HierarchyCacheService) InvalidateHierarchyCache(ctx context.Context, entityID string, entityType string)
```

### Task 5.2: Add Cache Warming
**File:** `/Users/kaushik/aaa-service/internal/services/warmup/hierarchy_warmup.go` (NEW)
- Preload frequently accessed hierarchies on startup
- Background job to refresh cache

### Task 5.3: Performance Monitoring
**File:** `/Users/kaushik/aaa-service/internal/metrics/hierarchy_metrics.go` (NEW)
- Track hierarchy query latency
- Monitor cache hit rates
- Alert on slow queries

## Phase 6: Testing & Validation (4 days)

### Task 6.1: Unit Tests for Hierarchy Operations
- Test circular reference detection
- Test depth limit enforcement
- Test hierarchy traversal methods
- Test cache invalidation

### Task 6.2: Integration Tests
- Test hierarchy moves with large trees
- Test role inheritance with complex hierarchies
- Test concurrent hierarchy modifications
- Test transaction rollback scenarios

### Task 6.3: Performance Tests
- Benchmark recursive CTE queries
- Test with deep hierarchies (10 levels)
- Test with wide hierarchies (1000+ children)
- Measure cache impact

### Task 6.4: End-to-End Tests
- Test complete user journey with hierarchy
- Test token generation with inherited roles
- Test API endpoints with various scenarios

## Implementation Order

### Week 1: Foundation
1. Database migration (Task 1.3)
2. Model updates (Task 1.4)
3. Circular reference validation (Task 1.1)
4. Depth validation (Task 1.2)
5. Role inheritance integration (Tasks 2.1-2.3)

### Week 2: Core Functionality
1. Recursive CTE queries (Tasks 3.1-3.2)
2. Materialized path management (Task 3.3)
3. Hierarchy service (Task 3.4)
4. Effective roles endpoint (Task 2.4)

### Week 3: API & Features
1. Organization hierarchy endpoints (Task 4.1)
2. Group hierarchy endpoints (Task 4.2)
3. Route updates (Task 4.3)
4. API documentation (Task 4.4)

### Week 4: Optimization & Testing
1. Cache implementation (Tasks 5.1-5.2)
2. Performance monitoring (Task 5.3)
3. Comprehensive testing (Tasks 6.1-6.4)
4. Performance tuning based on tests

## Rollback Plan

### Database Rollback
```sql
-- Remove hierarchy fields if needed
ALTER TABLE organizations DROP COLUMN IF EXISTS hierarchy_depth;
ALTER TABLE organizations DROP COLUMN IF EXISTS hierarchy_path;

ALTER TABLE groups DROP COLUMN IF EXISTS hierarchy_depth;
ALTER TABLE groups DROP COLUMN IF EXISTS hierarchy_path;
ALTER TABLE groups DROP COLUMN IF EXISTS root_group_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_org_hierarchy_path;
DROP INDEX IF EXISTS idx_org_parent_active;
DROP INDEX IF EXISTS idx_org_depth;

DROP INDEX IF EXISTS idx_group_hierarchy_path;
DROP INDEX IF EXISTS idx_group_org_parent;
DROP INDEX IF EXISTS idx_group_root;
```

### Code Rollback
- Git revert for each phase
- Feature flags for gradual rollout
- Keep old code paths during transition

## Success Metrics

1. **Performance**
   - Hierarchy queries < 100ms for 95th percentile
   - Cache hit rate > 80%
   - No timeout errors on hierarchy operations

2. **Reliability**
   - Zero circular references created
   - Zero hierarchy depth violations
   - 100% successful hierarchy validations

3. **Functionality**
   - Users receive all inherited roles in tokens
   - All hierarchy navigation working correctly
   - Hierarchy moves complete successfully

4. **Scale**
   - Support hierarchies up to 10 levels deep
   - Support organizations with 10,000+ children
   - Handle 100+ concurrent hierarchy operations

## Risk Mitigation

1. **Data Corruption**: Use transactions, validate before commit
2. **Performance Degradation**: Cache aggressively, use read replicas
3. **Breaking Changes**: Version APIs, maintain backward compatibility
4. **Complex Migrations**: Test on staging, have rollback ready
5. **Cache Inconsistency**: Use cache versioning, TTL limits

## Dependencies

- PostgreSQL 12+ (for recursive CTEs)
- Redis for caching
- kisanlink-db for database operations
- Existing role inheritance engine

## Next Steps

1. Review and approve implementation plan
2. Set up development branch
3. Begin Phase 1 implementation
4. Daily progress updates
5. Phase completion reviews
