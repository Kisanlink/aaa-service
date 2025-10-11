package role_assignments

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
)

// GetInheritedRoles retrieves all inherited roles for a role (recursive hierarchy)
func (s *Service) GetInheritedRoles(ctx context.Context, roleID string) ([]*models.Role, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("role:%s:inherited", roleID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if roles, ok := cached.([]*models.Role); ok {
				s.logger.Debug("Inherited roles cache hit", zap.String("role_id", roleID))
				return roles, nil
			}
		}
	}

	// Get the role
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Get hierarchical roles
	visited := make(map[string]bool)
	inheritedRoles := []*models.Role{}
	s.addRoleWithParents(ctx, role, &inheritedRoles, visited)

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, inheritedRoles, 300); err != nil { // 5 minutes TTL
			s.logger.Warn("Failed to cache inherited roles", zap.Error(err))
		}
	}

	return inheritedRoles, nil
}

// GetEffectiveRolesForUser retrieves all effective roles for a user including inherited roles
func (s *Service) GetEffectiveRolesForUser(ctx context.Context, userID, orgID, groupID string) ([]*models.Role, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user:%s:effective_roles:%s:%s", userID, orgID, groupID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if roles, ok := cached.([]*models.Role); ok {
				s.logger.Debug("User effective roles cache hit", zap.String("user_id", userID))
				return roles, nil
			}
		}
	}

	roles := []*models.Role{}
	visited := make(map[string]bool)

	// 1. Get direct user roles
	// TODO: Use UserRoleRepository when available
	// For now, this is a placeholder
	directRoles := []*models.Role{}

	// Add direct roles with hierarchy
	for _, role := range directRoles {
		s.addRoleWithHierarchy(ctx, role, &roles, visited)
	}

	// 2. Get group roles if groupID is provided
	if groupID != "" {
		groupRoles, err := s.getGroupRoles(ctx, groupID)
		if err == nil {
			for _, role := range groupRoles {
				s.addRoleWithHierarchy(ctx, role, &roles, visited)
			}
		}
	}

	// 3. Get organization roles if orgID is provided
	if orgID != "" {
		orgRoles, err := s.getOrganizationRoles(ctx, orgID)
		if err == nil {
			for _, role := range orgRoles {
				s.addRoleWithHierarchy(ctx, role, &roles, visited)
			}
		}
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, roles, 300); err != nil { // 5 minutes TTL
			s.logger.Warn("Failed to cache effective roles", zap.Error(err))
		}
	}

	s.logger.Info("Effective roles computed for user",
		zap.String("user_id", userID),
		zap.Int("total_roles", len(roles)))

	return roles, nil
}

// GetEffectivePermissionsForUser retrieves all effective permissions for a user
func (s *Service) GetEffectivePermissionsForUser(ctx context.Context, userID, orgID, groupID string) ([]*models.Permission, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get effective roles
	roles, err := s.GetEffectiveRolesForUser(ctx, userID, orgID, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective roles: %w", err)
	}

	// Get all permissions from these roles
	permissionMap := make(map[string]*models.Permission)
	for _, role := range roles {
		rolePerms, err := s.GetRolePermissions(ctx, role.ID)
		if err != nil {
			s.logger.Warn("Failed to get permissions for role",
				zap.String("role_id", role.ID),
				zap.Error(err))
			continue
		}

		for _, perm := range rolePerms {
			permissionMap[perm.ID] = perm
		}
	}

	// Convert map to slice
	permissions := make([]*models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	s.logger.Info("Effective permissions computed for user",
		zap.String("user_id", userID),
		zap.Int("total_permissions", len(permissions)))

	return permissions, nil
}

// GetUserAccessToResource retrieves all allowed actions for a user on a specific resource
func (s *Service) GetUserAccessToResource(ctx context.Context, userID, resourceType, resourceID string) ([]string, error) {
	if userID == "" || resourceType == "" || resourceID == "" {
		return nil, fmt.Errorf("user ID, resource type, and resource ID are required")
	}

	// Get effective roles for user
	roles, err := s.GetEffectiveRolesForUser(ctx, userID, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get effective roles: %w", err)
	}

	// Extract role IDs
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	// Get allowed actions from resource permissions
	actions, err := s.resourcePermissionRepo.GetAllowedActionsForRoles(ctx, roleIDs, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed actions: %w", err)
	}

	// TODO: Also check Model 1 permissions
	// This would require matching permissions by resource and action

	s.logger.Info("User access computed for resource",
		zap.String("user_id", userID),
		zap.String("resource_type", resourceType),
		zap.String("resource_id", resourceID),
		zap.Strings("actions", actions))

	return actions, nil
}

// addRoleWithHierarchy adds a role and its parent roles to the result
func (s *Service) addRoleWithHierarchy(ctx context.Context, role *models.Role, result *[]*models.Role, visited map[string]bool) {
	if role == nil || visited[role.ID] {
		return // Prevent cycles
	}

	visited[role.ID] = true
	*result = append(*result, role)

	// Add parent role if exists
	if role.ParentID != nil && *role.ParentID != "" {
		parent, err := s.roleRepo.GetByID(ctx, *role.ParentID, nil)
		if err == nil && parent != nil {
			s.addRoleWithHierarchy(ctx, parent, result, visited)
		}
	}
}

// addRoleWithParents recursively adds a role and its parent roles
func (s *Service) addRoleWithParents(ctx context.Context, role *models.Role, result *[]*models.Role, visited map[string]bool) {
	if role == nil || visited[role.ID] {
		return // Prevent cycles
	}

	visited[role.ID] = true
	*result = append(*result, role)

	// Add parent role if exists
	if role.ParentID != nil && *role.ParentID != "" {
		parent, err := s.roleRepo.GetByID(ctx, *role.ParentID, nil)
		if err == nil && parent != nil {
			s.addRoleWithParents(ctx, parent, result, visited)
		}
	}
}

// getGroupRoles retrieves roles assigned to a group
func (s *Service) getGroupRoles(ctx context.Context, groupID string) ([]*models.Role, error) {
	// TODO: Implement group role retrieval when GroupRoleRepository is available
	// For now, return empty slice
	s.logger.Debug("Group role retrieval not yet implemented", zap.String("group_id", groupID))
	return []*models.Role{}, nil
}

// getOrganizationRoles retrieves roles assigned to an organization
func (s *Service) getOrganizationRoles(ctx context.Context, orgID string) ([]*models.Role, error) {
	// TODO: Implement organization role retrieval
	// For now, return empty slice
	s.logger.Debug("Organization role retrieval not yet implemented", zap.String("org_id", orgID))
	return []*models.Role{}, nil
}

// GetRoleHierarchy retrieves the complete hierarchy tree for a role
func (s *Service) GetRoleHierarchy(ctx context.Context, roleID string) (map[string]interface{}, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	hierarchy := map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
	}

	// Add parent hierarchy
	if role.ParentID != nil && *role.ParentID != "" {
		parentHierarchy, err := s.GetRoleHierarchy(ctx, *role.ParentID)
		if err == nil {
			hierarchy["parent"] = parentHierarchy
		}
	}

	// Add child roles
	// TODO: Implement when role repository has GetChildren method
	hierarchy["children"] = []map[string]interface{}{}

	return hierarchy, nil
}

// HasCircularDependency checks if assigning a parent would create a circular dependency
func (s *Service) HasCircularDependency(ctx context.Context, roleID, parentID string) (bool, error) {
	if roleID == "" || parentID == "" {
		return false, fmt.Errorf("role ID and parent ID are required")
	}

	// Get all ancestors of the parent
	visited := make(map[string]bool)
	current := parentID

	for current != "" {
		if visited[current] {
			return true, nil // Found a cycle
		}

		if current == roleID {
			return true, nil // Would create a cycle
		}

		visited[current] = true

		// Get parent role
		role, err := s.roleRepo.GetByID(ctx, current, nil)
		if err != nil || role == nil || role.ParentID == nil {
			break
		}

		current = *role.ParentID
	}

	return false, nil
}
