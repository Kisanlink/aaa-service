# Critical Hierarchy Fixes - Implementation Code

## Fix 1: Circular Reference Prevention

### File: `/Users/kaushik/aaa-service/internal/repositories/organizations/organization_repository.go`

Add this method:
```go
// WouldCreateCircularReference checks if setting newParentID as parent of orgID would create a cycle
func (r *OrganizationRepository) WouldCreateCircularReference(ctx context.Context, orgID, newParentID string) (bool, error) {
    if orgID == newParentID {
        return true, nil // Self-reference
    }

    query := `
        WITH RECURSIVE ancestors AS (
            SELECT id, parent_id
            FROM organizations
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT o.id, o.parent_id
            FROM organizations o
            INNER JOIN ancestors a ON o.id = a.parent_id
            WHERE o.deleted_at IS NULL
        )
        SELECT COUNT(*) > 0 as would_create_cycle
        FROM ancestors
        WHERE id = $2
    `

    var result struct {
        WouldCreateCycle bool `gorm:"column:would_create_cycle"`
    }

    err := r.dbManager.GetDB().Raw(query, newParentID, orgID).Scan(&result).Error
    if err != nil {
        return false, fmt.Errorf("failed to check circular reference: %w", err)
    }

    return result.WouldCreateCycle, nil
}

// ValidateHierarchyDepth checks if the hierarchy would exceed max depth
func (r *OrganizationRepository) ValidateHierarchyDepth(ctx context.Context, orgID, newParentID string) error {
    // Get depth of new parent
    parentDepthQuery := `
        WITH RECURSIVE ancestors AS (
            SELECT id, parent_id, 0 as depth
            FROM organizations
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT o.id, o.parent_id, a.depth + 1
            FROM organizations o
            INNER JOIN ancestors a ON o.id = a.parent_id
            WHERE o.deleted_at IS NULL
        )
        SELECT MAX(depth) as max_depth FROM ancestors
    `

    var parentDepth int
    err := r.dbManager.GetDB().Raw(parentDepthQuery, newParentID).Scan(&parentDepth).Error
    if err != nil {
        return fmt.Errorf("failed to get parent depth: %w", err)
    }

    // Get max depth of org's subtree
    subtreeDepthQuery := `
        WITH RECURSIVE descendants AS (
            SELECT id, parent_id, 0 as depth
            FROM organizations
            WHERE parent_id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT o.id, o.parent_id, d.depth + 1
            FROM organizations o
            INNER JOIN descendants d ON o.parent_id = d.id
            WHERE o.deleted_at IS NULL
        )
        SELECT MAX(depth) as max_depth FROM descendants
    `

    var subtreeDepth int
    err = r.dbManager.GetDB().Raw(subtreeDepthQuery, orgID).Scan(&subtreeDepth).Error
    if err != nil {
        return fmt.Errorf("failed to get subtree depth: %w", err)
    }

    totalDepth := parentDepth + subtreeDepth + 1
    if totalDepth > 10 {
        return fmt.Errorf("hierarchy would exceed maximum depth of 10 (would be %d)", totalDepth)
    }

    return nil
}
```

### File: `/Users/kaushik/aaa-service/internal/services/organizations/organization_service.go`

Update CreateOrganization and UpdateOrganization methods:
```go
// In CreateOrganization method, after line 73 (parent validation):
if req.ParentID != nil && *req.ParentID != "" {
    // Existing parent validation...

    // NEW: Check for circular reference (shouldn't happen on create, but be safe)
    wouldCreateCycle, err := s.orgRepo.WouldCreateCircularReference(ctx, "", *req.ParentID)
    if err != nil {
        s.logger.Error("Failed to check circular reference", zap.Error(err))
        return nil, errors.NewInternalError(err)
    }
    if wouldCreateCycle {
        return nil, errors.NewValidationError("invalid parent: would create circular reference")
    }

    // NEW: Validate depth
    if err := s.orgRepo.ValidateHierarchyDepth(ctx, "", *req.ParentID); err != nil {
        s.logger.Warn("Hierarchy depth validation failed", zap.Error(err))
        return nil, errors.NewValidationError("hierarchy depth limit exceeded")
    }
}

// Add new method for updating organization
func (s *Service) UpdateOrganizationParent(ctx context.Context, orgID string, newParentID *string) error {
    // Validate the organization exists
    org, err := s.orgRepo.GetByID(ctx, orgID)
    if err != nil {
        return errors.NewNotFoundError("organization not found")
    }

    // If setting a parent
    if newParentID != nil && *newParentID != "" {
        // Check circular reference
        wouldCreateCycle, err := s.orgRepo.WouldCreateCircularReference(ctx, orgID, *newParentID)
        if err != nil {
            s.logger.Error("Failed to check circular reference", zap.Error(err))
            return errors.NewInternalError(err)
        }
        if wouldCreateCycle {
            s.logger.Warn("Attempted to create circular reference",
                zap.String("org_id", orgID),
                zap.String("new_parent_id", *newParentID))
            return errors.NewValidationError("would create circular reference in hierarchy")
        }

        // Check depth limit
        if err := s.orgRepo.ValidateHierarchyDepth(ctx, orgID, *newParentID); err != nil {
            s.logger.Warn("Hierarchy depth exceeded",
                zap.String("org_id", orgID),
                zap.String("new_parent_id", *newParentID),
                zap.Error(err))
            return errors.NewValidationError(err.Error())
        }
    }

    // Update the parent
    oldParentID := org.ParentID
    org.ParentID = newParentID

    if err := s.orgRepo.Update(ctx, org); err != nil {
        s.logger.Error("Failed to update organization parent", zap.Error(err))
        return errors.NewInternalError(err)
    }

    // Audit log
    s.auditService.LogOrganizationOperation(ctx, "system",
        models.AuditActionChangeOrganizationHierarchy, orgID,
        "Organization parent changed", true, map[string]interface{}{
            "old_parent": oldParentID,
            "new_parent": newParentID,
        })

    // Invalidate caches
    s.orgCache.InvalidateOrganization(ctx, orgID)
    if oldParentID != nil {
        s.orgCache.InvalidateOrganization(ctx, *oldParentID)
    }
    if newParentID != nil {
        s.orgCache.InvalidateOrganization(ctx, *newParentID)
    }

    return nil
}
```

## Fix 2: Role Inheritance Integration

### File: `/Users/kaushik/aaa-service/internal/services/user/additional_methods.go`

Modify GetUserWithRoles method:
```go
// GetUserWithRoles retrieves user with their roles (including inherited ones)
func (s *UserService) GetUserWithRoles(ctx context.Context, userID string) (*responses.UserResponse, error) {
    // Get the user
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Get direct roles
    directRoles, err := s.getCachedUserRoles(ctx, userID)
    if err != nil {
        s.logger.Error("Failed to get user roles", zap.Error(err))
        return nil, err
    }

    // Convert to response
    userResponse := &responses.UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Roles:     []responses.RoleResponse{},
    }

    // Add direct roles
    roleMap := make(map[string]*responses.RoleResponse)
    for _, role := range directRoles {
        roleResp := responses.RoleResponse{
            ID:          role.ID,
            Name:        role.Name,
            Description: role.Description,
            UserID:      userID,
            IsActive:    role.IsActive,
            IsDirect:    true, // Mark as direct role
        }
        roleMap[role.ID] = &roleResp
    }

    // NEW: Get user's primary organization (you may need to add this method)
    userOrgs, err := s.GetUserOrganizations(ctx, userID, 1, 0)
    if err == nil && len(userOrgs) > 0 {
        primaryOrgID := userOrgs[0].ID

        // Get inherited roles from groups
        inheritedRoles, err := s.groupService.GetEffectiveRoles(ctx, userID, primaryOrgID)
        if err != nil {
            s.logger.Warn("Failed to get inherited roles, using direct roles only",
                zap.String("user_id", userID),
                zap.Error(err))
        } else {
            // Add inherited roles (if not already present as direct)
            for _, inheritedRole := range inheritedRoles {
                if _, exists := roleMap[inheritedRole.ID]; !exists {
                    roleResp := responses.RoleResponse{
                        ID:          inheritedRole.ID,
                        Name:        inheritedRole.Name,
                        Description: inheritedRole.Description,
                        UserID:      userID,
                        IsActive:    true,
                        IsDirect:    false, // Mark as inherited
                        // You might want to add InheritedFrom field to track source
                    }
                    roleMap[inheritedRole.ID] = &roleResp
                }
            }
        }
    }

    // Convert map to slice
    for _, role := range roleMap {
        userResponse.Roles = append(userResponse.Roles, *role)
    }

    return userResponse, nil
}
```

### File: `/Users/kaushik/aaa-service/internal/services/groups/group_service.go`

Add method to expose effective roles:
```go
// GetEffectiveRoles returns all effective roles for a user in an organization (direct + inherited)
func (s *Service) GetEffectiveRoles(ctx context.Context, userID, organizationID string) ([]*models.Role, error) {
    // Use the existing role inheritance engine
    effectiveRoles, err := s.roleInheritanceEngine.CalculateEffectiveRoles(ctx, userID, organizationID)
    if err != nil {
        s.logger.Error("Failed to calculate effective roles",
            zap.String("user_id", userID),
            zap.String("org_id", organizationID),
            zap.Error(err))
        return nil, err
    }

    // Convert to role models
    var roles []*models.Role
    for _, er := range effectiveRoles {
        roles = append(roles, er.Role)
    }

    return roles, nil
}

// GetEffectiveRolesDetailed returns detailed effective roles with inheritance information
func (s *Service) GetEffectiveRolesDetailed(ctx context.Context, userID, organizationID string) ([]EffectiveRoleResponse, error) {
    effectiveRoles, err := s.roleInheritanceEngine.CalculateEffectiveRoles(ctx, userID, organizationID)
    if err != nil {
        return nil, err
    }

    var detailed []EffectiveRoleResponse
    for _, er := range effectiveRoles {
        detailed = append(detailed, EffectiveRoleResponse{
            Role:            er.Role,
            SourceGroupID:   er.GroupID,
            SourceGroupName: er.GroupName,
            InheritancePath: er.InheritancePath,
            Distance:        er.Distance,
            IsDirect:        er.IsDirectRole,
        })
    }

    return detailed, nil
}
```

## Fix 3: Efficient Hierarchy Queries

### File: `/Users/kaushik/aaa-service/internal/repositories/organizations/hierarchy_queries.go` (NEW)

```go
package organizations

import (
    "context"
    "fmt"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// GetOrganizationTree retrieves the complete organization tree using recursive CTE
func (r *OrganizationRepository) GetOrganizationTree(ctx context.Context, rootID string, maxDepth int) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE org_tree AS (
            SELECT o.*,
                   0 as level,
                   o.id::text as path,
                   ARRAY[o.name] as name_path
            FROM organizations o
            WHERE o.id = $1 AND o.deleted_at IS NULL

            UNION ALL

            SELECT o.*,
                   ot.level + 1,
                   ot.path || '/' || o.id::text,
                   ot.name_path || o.name
            FROM organizations o
            INNER JOIN org_tree ot ON o.parent_id = ot.id
            WHERE o.deleted_at IS NULL
              AND ot.level < $2
        )
        SELECT * FROM org_tree
        ORDER BY path;
    `

    var orgs []*models.Organization
    err := r.dbManager.GetDB().Raw(query, rootID, maxDepth).Scan(&orgs).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get organization tree: %w", err)
    }

    return orgs, nil
}

// GetAllAncestors retrieves all ancestors of an organization
func (r *OrganizationRepository) GetAllAncestors(ctx context.Context, orgID string) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE ancestors AS (
            -- Start with the organization's parent
            SELECT o.*, 1 as distance
            FROM organizations o
            WHERE o.id = (
                SELECT parent_id FROM organizations WHERE id = $1
            ) AND o.deleted_at IS NULL

            UNION ALL

            -- Recursively get each parent
            SELECT o.*, a.distance + 1
            FROM organizations o
            INNER JOIN ancestors a ON o.id = (
                SELECT parent_id FROM organizations WHERE id = a.id
            )
            WHERE o.deleted_at IS NULL
        )
        SELECT * FROM ancestors
        ORDER BY distance;
    `

    var ancestors []*models.Organization
    err := r.dbManager.GetDB().Raw(query, orgID).Scan(&ancestors).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get ancestors: %w", err)
    }

    return ancestors, nil
}

// GetAllDescendants retrieves all descendants of an organization
func (r *OrganizationRepository) GetAllDescendants(ctx context.Context, orgID string, maxDepth int) ([]*models.Organization, error) {
    query := `
        WITH RECURSIVE descendants AS (
            -- Start with direct children
            SELECT o.*, 1 as depth
            FROM organizations o
            WHERE o.parent_id = $1 AND o.deleted_at IS NULL

            UNION ALL

            -- Recursively get children of children
            SELECT o.*, d.depth + 1
            FROM organizations o
            INNER JOIN descendants d ON o.parent_id = d.id
            WHERE o.deleted_at IS NULL
              AND d.depth < $2
        )
        SELECT * FROM descendants
        ORDER BY depth, name;
    `

    var descendants []*models.Organization
    err := r.dbManager.GetDB().Raw(query, orgID, maxDepth).Scan(&descendants).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get descendants: %w", err)
    }

    return descendants, nil
}

// GetHierarchyStatistics returns statistics about the organization hierarchy
func (r *OrganizationRepository) GetHierarchyStatistics(ctx context.Context, orgID string) (*HierarchyStats, error) {
    query := `
        WITH RECURSIVE tree_stats AS (
            SELECT id, parent_id, 0 as level
            FROM organizations
            WHERE id = $1 AND deleted_at IS NULL

            UNION ALL

            SELECT o.id, o.parent_id, ts.level + 1
            FROM organizations o
            INNER JOIN tree_stats ts ON o.parent_id = ts.id
            WHERE o.deleted_at IS NULL
        )
        SELECT
            COUNT(*) as total_descendants,
            MAX(level) as max_depth,
            COUNT(DISTINCT level) as unique_levels,
            AVG(level)::numeric(10,2) as avg_depth
        FROM tree_stats
        WHERE level > 0;
    `

    var stats HierarchyStats
    err := r.dbManager.GetDB().Raw(query, orgID).Scan(&stats).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get hierarchy statistics: %w", err)
    }

    return &stats, nil
}

type HierarchyStats struct {
    TotalDescendants int     `json:"total_descendants"`
    MaxDepth         int     `json:"max_depth"`
    UniqueLevels     int     `json:"unique_levels"`
    AvgDepth         float64 `json:"avg_depth"`
}
```

## Fix 4: Database Migration

### File: `/Users/kaushik/aaa-service/migrations/20251117_add_hierarchy_fields.sql`

```sql
-- Add hierarchy tracking fields to organizations
ALTER TABLE organizations
ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS hierarchy_path TEXT;

-- Add hierarchy tracking fields to groups
ALTER TABLE groups
ADD COLUMN IF NOT EXISTS hierarchy_depth INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS hierarchy_path TEXT,
ADD COLUMN IF NOT EXISTS root_group_id VARCHAR(255);

-- Add check constraints for depth limits
ALTER TABLE organizations
ADD CONSTRAINT check_org_hierarchy_depth
CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 10);

ALTER TABLE groups
ADD CONSTRAINT check_group_hierarchy_depth
CHECK (hierarchy_depth >= 0 AND hierarchy_depth <= 8);

-- Create indexes for efficient hierarchy queries
CREATE INDEX IF NOT EXISTS idx_org_hierarchy_path
ON organizations USING btree(hierarchy_path text_pattern_ops)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_org_parent_active
ON organizations(parent_id, is_active)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_org_depth
ON organizations(hierarchy_depth)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_group_hierarchy_path
ON groups USING btree(hierarchy_path text_pattern_ops)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_group_org_parent
ON groups(organization_id, parent_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_group_root
ON groups(root_group_id)
WHERE root_group_id IS NOT NULL AND deleted_at IS NULL;

-- Create function to update hierarchy paths
CREATE OR REPLACE FUNCTION update_hierarchy_path()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.hierarchy_path = '/' || NEW.id || '/';
        NEW.hierarchy_depth = 0;
    ELSE
        SELECT
            hierarchy_path || NEW.id || '/',
            hierarchy_depth + 1
        INTO
            NEW.hierarchy_path,
            NEW.hierarchy_depth
        FROM organizations
        WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic path updates
CREATE TRIGGER trg_update_org_hierarchy_path
BEFORE INSERT OR UPDATE OF parent_id ON organizations
FOR EACH ROW
EXECUTE FUNCTION update_hierarchy_path();

-- Initialize existing data
UPDATE organizations
SET hierarchy_depth = 0,
    hierarchy_path = '/' || id || '/'
WHERE parent_id IS NULL;

-- Update paths for organizations with parents (run multiple times for deep hierarchies)
DO $$
DECLARE
    rows_updated INTEGER;
BEGIN
    LOOP
        WITH updated AS (
            UPDATE organizations o
            SET hierarchy_path = p.hierarchy_path || o.id || '/',
                hierarchy_depth = p.hierarchy_depth + 1
            FROM organizations p
            WHERE o.parent_id = p.id
              AND o.hierarchy_path IS NULL
              AND p.hierarchy_path IS NOT NULL
            RETURNING o.id
        )
        SELECT COUNT(*) INTO rows_updated FROM updated;

        EXIT WHEN rows_updated = 0;
    END LOOP;
END $$;
```

## Testing the Fixes

### Test Circular Reference Prevention
```go
func TestCircularReferencePrevention(t *testing.T) {
    // Setup
    orgA := createOrganization("A", nil)
    orgB := createOrganization("B", &orgA.ID)
    orgC := createOrganization("C", &orgB.ID)

    // Test 1: Try to make C parent of A (would create cycle)
    err := orgService.UpdateOrganizationParent(ctx, orgA.ID, &orgC.ID)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular reference")

    // Test 2: Try self-reference
    err = orgService.UpdateOrganizationParent(ctx, orgA.ID, &orgA.ID)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular reference")
}
```

### Test Role Inheritance
```go
func TestRoleInheritanceInToken(t *testing.T) {
    // Setup user with group membership
    user := createUser("testuser")
    group := createGroup("TestGroup", orgID)
    role := createRole("TestRole")

    // Assign role to group
    groupService.AssignRoleToGroup(ctx, group.ID, role.ID)

    // Add user to group
    groupService.AddUserToGroup(ctx, user.ID, group.ID)

    // Login and get token
    token := authService.Login(ctx, "testuser", "password")

    // Decode token and verify roles
    claims := decodeToken(token)
    assert.Contains(t, claims.Roles, role.ID)
}
```

### Test Efficient Hierarchy Query
```go
func BenchmarkHierarchyQueries(b *testing.B) {
    // Create deep hierarchy (10 levels)
    rootOrg := createOrganization("Root", nil)
    currentParent := rootOrg.ID
    for i := 0; i < 10; i++ {
        child := createOrganization(fmt.Sprintf("Level%d", i), &currentParent)
        currentParent = child.ID
    }

    b.Run("GetOrganizationTree", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            tree, err := orgRepo.GetOrganizationTree(ctx, rootOrg.ID, 10)
            assert.NoError(b, err)
            assert.Equal(b, 11, len(tree)) // root + 10 children
        }
    })

    // Should complete in < 100ms even with deep hierarchy
    assert.Less(b, b.Elapsed(), 100*time.Millisecond)
}
```
