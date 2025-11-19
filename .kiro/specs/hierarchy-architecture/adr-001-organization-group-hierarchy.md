# ADR-001: Organization and Group Hierarchy Architecture

**Date:** 2025-11-17
**Status:** Proposed
**Author:** SDE-3 Backend Architect

## Context

The AAA service currently has partial hierarchy support for organizations and groups, but the implementation is incomplete and incorrectly configured. The current issues include:

1. **Incomplete Hierarchy Navigation**: Missing methods for traversing organization and group hierarchies efficiently
2. **Inefficient Queries**: No optimized paths for hierarchy queries (missing recursive CTEs, proper indexes)
3. **Missing Circular Reference Protection**: Inadequate validation to prevent circular hierarchies
4. **Incomplete API Surface**: Missing endpoints for hierarchy management and navigation
5. **Role Inheritance Not Integrated**: Group role inheritance engine exists but not integrated with token generation
6. **Performance Issues**: No caching strategy for hierarchy traversal
7. **Missing Depth Limits**: No constraints on hierarchy depth leading to potential performance issues

## Current State Analysis

### Data Models

#### Organization Model (`internal/entities/models/organization.go`)
- **Has**: ParentID field for hierarchy support
- **Has**: Parent and Children relationships defined
- **Missing**: Depth tracking, path materialization, hierarchy validation

#### Group Model (`internal/entities/models/group.go`)
- **Has**: ParentID field for hierarchy within organizations
- **Has**: GroupInheritance table for complex inheritance patterns
- **Has**: Parent and Children relationships
- **Missing**: Effective hierarchy calculation, depth limits

### Repository Layer

#### Organization Repository
- **Has**: Basic GetChildren(), GetParentHierarchy() methods
- **Issues**:
  - GetParentHierarchy() uses inefficient loop-based traversal
  - No recursive CTE queries for efficient hierarchy retrieval
  - Missing batch operations for hierarchy updates

#### Group Repository
- **Has**: Basic GetChildren() method
- **Missing**:
  - GetAncestors() method
  - GetDescendants() method
  - GetHierarchyPath() method
  - Circular reference validation

### Service Layer

#### Organization Service
- **Has**: Basic CRUD operations
- **Missing**:
  - MoveOrganization() for hierarchy changes
  - GetOrganizationTree() for full hierarchy
  - ValidateHierarchyChange() for move operations
  - Hierarchy depth enforcement

#### Group Service
- **Has**: Role assignment and inheritance engine
- **Issues**:
  - Role inheritance not integrated with authentication flow
  - No hierarchy validation on group creation/updates
  - Missing group movement within hierarchy

### Role Inheritance

#### Current Implementation
- **Location**: `internal/services/groups/role_inheritance_engine.go`
- **Type**: Bottom-up inheritance (roles flow from child to parent groups)
- **Status**: Fully implemented but NOT integrated with token generation
- **Issue**: Users only receive direct role assignments in JWT tokens, not inherited roles

## Decision

We will implement a comprehensive hierarchy system with the following components:

### 1. Enhanced Data Model

#### Organization Hierarchy
```go
type Organization struct {
    // Existing fields...

    // Hierarchy fields
    ParentID     *string `gorm:"index:idx_org_parent"`
    HierarchyDepth int   `gorm:"default:0;check:hierarchy_depth >= 0 AND hierarchy_depth <= 10"`
    HierarchyPath  string `gorm:"type:text"` // Materialized path: /root/parent/current/

    // Computed fields
    ChildCount     int       `gorm:"-"`
    DescendantCount int      `gorm:"-"`
    AncestorIDs    []string  `gorm:"-"`
}
```

#### Group Hierarchy
```go
type Group struct {
    // Existing fields...

    // Enhanced hierarchy
    HierarchyDepth int    `gorm:"default:0;check:hierarchy_depth >= 0 AND hierarchy_depth <= 8"`
    HierarchyPath  string `gorm:"type:text"` // Path within organization
    RootGroupID    *string `gorm:"index"` // Denormalized for performance
}
```

### 2. Database Schema Enhancements

#### Indexes
```sql
-- Organization hierarchy indexes
CREATE INDEX idx_org_hierarchy_path ON organizations USING btree(hierarchy_path text_pattern_ops);
CREATE INDEX idx_org_parent_active ON organizations(parent_id, is_active);
CREATE INDEX idx_org_depth ON organizations(hierarchy_depth);

-- Group hierarchy indexes
CREATE INDEX idx_group_hierarchy_path ON groups USING btree(hierarchy_path text_pattern_ops);
CREATE INDEX idx_group_org_parent ON groups(organization_id, parent_id);
CREATE INDEX idx_group_root ON groups(root_group_id) WHERE root_group_id IS NOT NULL;

-- Recursive query optimization
CREATE INDEX idx_org_hierarchy_traverse ON organizations(id, parent_id, is_active);
CREATE INDEX idx_group_hierarchy_traverse ON groups(id, parent_id, organization_id, is_active);
```

#### Constraints
```sql
-- Prevent circular references via triggers
CREATE OR REPLACE FUNCTION check_organization_hierarchy()
RETURNS TRIGGER AS $$
BEGIN
    -- Check for circular reference
    IF NEW.parent_id IS NOT NULL THEN
        IF EXISTS (
            WITH RECURSIVE ancestors AS (
                SELECT id, parent_id FROM organizations WHERE id = NEW.parent_id
                UNION ALL
                SELECT o.id, o.parent_id
                FROM organizations o
                INNER JOIN ancestors a ON o.id = a.parent_id
            )
            SELECT 1 FROM ancestors WHERE id = NEW.id
        ) THEN
            RAISE EXCEPTION 'Circular reference detected in organization hierarchy';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_org_hierarchy
BEFORE INSERT OR UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION check_organization_hierarchy();
```

### 3. Repository Layer Enhancements

#### Efficient Hierarchy Queries
```go
// GetOrganizationHierarchy retrieves complete hierarchy using recursive CTE
func (r *OrganizationRepository) GetOrganizationHierarchy(ctx context.Context, rootID string, maxDepth int) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE org_tree AS (
            SELECT o.*, 0 as level
            FROM organizations o
            WHERE o.id = $1 AND o.deleted_at IS NULL

            UNION ALL

            SELECT o.*, ot.level + 1
            FROM organizations o
            INNER JOIN org_tree ot ON o.parent_id = ot.id
            WHERE o.deleted_at IS NULL AND ot.level < $2
        )
        SELECT * FROM org_tree ORDER BY level, name;
    `

    var orgs []*models.Organization
    err := r.dbManager.GetDB().Raw(query, rootID, maxDepth).Scan(&orgs).Error
    return orgs, err
}

// GetAllAncestors retrieves all ancestors efficiently
func (r *OrganizationRepository) GetAllAncestors(ctx context.Context, orgID string) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE ancestors AS (
            SELECT o.*, 0 as distance
            FROM organizations o
            WHERE o.id = $1

            UNION ALL

            SELECT o.*, a.distance + 1
            FROM organizations o
            INNER JOIN ancestors a ON o.id = a.parent_id
            WHERE o.deleted_at IS NULL
        )
        SELECT * FROM ancestors WHERE distance > 0 ORDER BY distance;
    `

    var ancestors []*models.Organization
    err := r.dbManager.GetDB().Raw(query, orgID).Scan(&ancestors).Error
    return ancestors, err
}

// GetAllDescendants retrieves all descendants with depth limit
func (r *OrganizationRepository) GetAllDescendants(ctx context.Context, orgID string, maxDepth int) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE descendants AS (
            SELECT o.*, 0 as depth
            FROM organizations o
            WHERE o.parent_id = $1 AND o.deleted_at IS NULL

            UNION ALL

            SELECT o.*, d.depth + 1
            FROM organizations o
            INNER JOIN descendants d ON o.parent_id = d.id
            WHERE o.deleted_at IS NULL AND d.depth < $2
        )
        SELECT * FROM descendants ORDER BY depth, name;
    `

    var descendants []*models.Organization
    err := r.dbManager.GetDB().Raw(query, orgID, maxDepth).Scan(&descendants).Error
    return descendants, err
}
```

### 4. Service Layer Enhancements

#### Hierarchy Management Service
```go
type HierarchyService struct {
    orgRepo      interfaces.OrganizationRepository
    groupRepo    interfaces.GroupRepository
    cache        interfaces.CacheService
    auditService interfaces.AuditService
    logger       *zap.Logger
}

// MoveOrganization moves an organization to a new parent with validation
func (s *HierarchyService) MoveOrganization(ctx context.Context, orgID, newParentID string) error {
    // 1. Validate move is allowed
    if err := s.ValidateOrganizationMove(ctx, orgID, newParentID); err != nil {
        return err
    }

    // 2. Begin transaction
    tx := s.orgRepo.BeginTransaction()
    defer tx.Rollback()

    // 3. Update organization parent
    org, err := s.orgRepo.GetByID(ctx, orgID)
    if err != nil {
        return err
    }

    oldParentID := org.ParentID
    org.ParentID = &newParentID

    // 4. Recalculate hierarchy paths for org and all descendants
    if err := s.recalculateHierarchyPaths(ctx, tx, orgID); err != nil {
        return err
    }

    // 5. Update depths
    if err := s.updateHierarchyDepths(ctx, tx, orgID); err != nil {
        return err
    }

    // 6. Invalidate caches
    s.invalidateHierarchyCaches(ctx, orgID, oldParentID, &newParentID)

    // 7. Audit log
    s.auditService.LogOrganizationOperation(ctx, "system",
        models.AuditActionChangeOrganizationHierarchy, orgID,
        "Organization hierarchy changed", true, map[string]interface{}{
            "old_parent": oldParentID,
            "new_parent": newParentID,
        })

    return tx.Commit().Error
}

// ValidateOrganizationMove checks if a move is valid
func (s *HierarchyService) ValidateOrganizationMove(ctx context.Context, orgID, newParentID string) error {
    // 1. Check self-reference
    if orgID == newParentID {
        return errors.NewValidationError("organization cannot be its own parent")
    }

    // 2. Check circular reference
    descendants, err := s.orgRepo.GetAllDescendants(ctx, orgID, 100)
    if err != nil {
        return err
    }

    for _, desc := range descendants {
        if desc.ID == newParentID {
            return errors.NewValidationError("circular reference: new parent is a descendant")
        }
    }

    // 3. Check depth limit
    newParentDepth, err := s.getOrganizationDepth(ctx, newParentID)
    if err != nil {
        return err
    }

    orgSubtreeDepth, err := s.getSubtreeDepth(ctx, orgID)
    if err != nil {
        return err
    }

    if newParentDepth + orgSubtreeDepth > 10 {
        return errors.NewValidationError("move would exceed maximum hierarchy depth of 10")
    }

    return nil
}
```

### 5. API Enhancements

#### New Endpoints
```yaml
# Organization Hierarchy
GET    /api/v1/organizations/{id}/hierarchy      # Full hierarchy tree
GET    /api/v1/organizations/{id}/ancestors      # All ancestors
GET    /api/v1/organizations/{id}/descendants    # All descendants
GET    /api/v1/organizations/{id}/children       # Direct children only
POST   /api/v1/organizations/{id}/move           # Move to new parent
GET    /api/v1/organizations/{id}/path           # Hierarchy path

# Group Hierarchy
GET    /api/v1/groups/{id}/hierarchy             # Full hierarchy tree
GET    /api/v1/groups/{id}/ancestors             # All ancestors
GET    /api/v1/groups/{id}/descendants           # All descendants
POST   /api/v1/groups/{id}/move                  # Move within organization
GET    /api/v1/groups/{id}/effective-roles       # Including inherited roles

# Hierarchy Validation
POST   /api/v1/organizations/validate-hierarchy  # Validate proposed changes
POST   /api/v1/groups/validate-hierarchy         # Validate proposed changes
```

### 6. Role Inheritance Integration

#### Token Generation Enhancement
```go
// GetEffectiveRolesForToken retrieves all roles including inherited ones
func (s *UserService) GetEffectiveRolesForToken(ctx context.Context, userID string, orgID string) ([]models.Role, error) {
    // 1. Get direct user roles
    directRoles, err := s.userRoleRepo.GetActiveRolesByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 2. Get inherited roles from groups
    inheritanceEngine := s.groupService.GetRoleInheritanceEngine()
    effectiveRoles, err := inheritanceEngine.CalculateEffectiveRoles(ctx, userID, orgID)
    if err != nil {
        return nil, err
    }

    // 3. Merge and deduplicate
    roleMap := make(map[string]*models.Role)

    // Add direct roles (highest priority)
    for _, ur := range directRoles {
        if ur.Role != nil {
            roleMap[ur.RoleID] = ur.Role
        }
    }

    // Add inherited roles (if not already present)
    for _, er := range effectiveRoles {
        if _, exists := roleMap[er.Role.ID]; !exists {
            roleMap[er.Role.ID] = er.Role
        }
    }

    // 4. Convert to slice
    var allRoles []models.Role
    for _, role := range roleMap {
        allRoles = append(allRoles, *role)
    }

    return allRoles, nil
}
```

### 7. Caching Strategy

```go
type HierarchyCacheService struct {
    cache  interfaces.CacheService
    logger *zap.Logger
}

// Cache patterns
const (
    OrgHierarchyKey     = "org:%s:hierarchy"       // Full hierarchy
    OrgAncestorsKey     = "org:%s:ancestors"       // Ancestor list
    OrgDescendantsKey   = "org:%s:descendants"     // Descendant list
    OrgDepthKey         = "org:%s:depth"           // Hierarchy depth
    GroupHierarchyKey   = "group:%s:hierarchy"     // Group hierarchy
    GroupEffectiveRoles = "group:%s:user:%s:roles" // Effective roles
)

// GetCachedHierarchy retrieves cached hierarchy with automatic refresh
func (s *HierarchyCacheService) GetCachedHierarchy(ctx context.Context, orgID string,
    loader func() (*OrgHierarchy, error)) (*OrgHierarchy, error) {

    cacheKey := fmt.Sprintf(OrgHierarchyKey, orgID)

    // Try cache first
    if cached, found := s.cache.Get(cacheKey); found {
        if hierarchy, ok := cached.(*OrgHierarchy); ok {
            return hierarchy, nil
        }
    }

    // Load from database
    hierarchy, err := loader()
    if err != nil {
        return nil, err
    }

    // Cache with 15-minute TTL
    s.cache.Set(cacheKey, hierarchy, 900)

    return hierarchy, nil
}

// InvalidateHierarchyCache invalidates all related caches
func (s *HierarchyCacheService) InvalidateHierarchyCache(ctx context.Context, orgID string) {
    patterns := []string{
        fmt.Sprintf(OrgHierarchyKey, orgID),
        fmt.Sprintf(OrgAncestorsKey, "*"),
        fmt.Sprintf(OrgDescendantsKey, "*"),
        fmt.Sprintf("org:*:children:%s", orgID),
    }

    for _, pattern := range patterns {
        s.cache.DeletePattern(pattern)
    }
}
```

### 8. Performance Optimizations

#### Materialized Paths
- Store full hierarchy path for each entity
- Enable fast ancestor/descendant queries without recursion
- Update paths on hierarchy changes

#### Denormalized Counts
- Cache child/descendant counts
- Update via database triggers or async jobs
- Reduce COUNT queries

#### Read Replicas
- Route hierarchy read operations to read replicas
- Keep write operations on primary

## Consequences

### Positive
- **Efficient Hierarchy Traversal**: Recursive CTEs and materialized paths enable fast queries
- **Data Integrity**: Circular reference prevention and depth limits ensure valid hierarchies
- **Scalability**: Caching and indexing strategies support large hierarchies
- **Role Inheritance**: Proper integration ensures users get all entitled permissions
- **Audit Trail**: Complete tracking of hierarchy changes
- **API Completeness**: Full CRUD and navigation operations for hierarchies

### Negative
- **Complexity**: More complex data model and business logic
- **Migration Effort**: Existing data needs migration to add hierarchy fields
- **Cache Invalidation**: Complex cache invalidation patterns
- **Transaction Size**: Hierarchy moves can affect many records

### Neutral
- **Storage Overhead**: Materialized paths and denormalized fields increase storage
- **Maintenance**: Hierarchy paths need recalculation on structure changes

## Implementation Priority

1. **Phase 1: Critical Fixes** (Immediate)
   - Add circular reference validation
   - Implement depth limits
   - Add missing hierarchy navigation methods

2. **Phase 2: Role Integration** (Week 1)
   - Integrate role inheritance with token generation
   - Add effective roles API endpoints
   - Cache effective roles

3. **Phase 3: Performance** (Week 2)
   - Add recursive CTE queries
   - Implement materialized paths
   - Add hierarchy-specific indexes

4. **Phase 4: API Completion** (Week 3)
   - Add hierarchy management endpoints
   - Implement move operations
   - Add validation endpoints

5. **Phase 5: Optimization** (Week 4)
   - Implement comprehensive caching
   - Add denormalized counts
   - Performance testing and tuning

## Security Considerations

1. **Authorization**: Hierarchy changes require elevated permissions
2. **Rate Limiting**: Hierarchy operations should be rate-limited
3. **Audit Logging**: All hierarchy changes must be audited
4. **Input Validation**: Strict validation on depth and parent references
5. **Transaction Isolation**: Use appropriate isolation levels for hierarchy updates

## Monitoring

- Track hierarchy operation latency
- Monitor cache hit rates
- Alert on circular reference attempts
- Track hierarchy depth distribution
- Monitor failed hierarchy validations

## References

- PostgreSQL Recursive CTEs: https://www.postgresql.org/docs/current/queries-with.html
- Materialized Path Pattern: https://docs.mongodb.com/manual/tutorial/model-tree-structures-with-materialized-paths/
- RBAC Best Practices: https://www.osohq.com/academy/role-based-access-control-rbac
