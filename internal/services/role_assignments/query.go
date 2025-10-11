package role_assignments

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"go.uber.org/zap"
)

// GetRolePermissions retrieves all permissions assigned to a role (Model 1)
func (s *Service) GetRolePermissions(ctx context.Context, roleID string) ([]*models.Permission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("role:%s:permissions", roleID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if permissions, ok := cached.([]*models.Permission); ok {
				s.logger.Debug("Role permissions cache hit", zap.String("role_id", roleID))
				return permissions, nil
			}
		}
	}

	// Get role-permission assignments
	rolePerms, err := s.rolePermissionRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role permissions",
			zap.String("role_id", roleID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Extract permissions
	permissions := make([]*models.Permission, 0, len(rolePerms))
	for _, rp := range rolePerms {
		if rp.Permission != nil {
			permissions = append(permissions, rp.Permission)
		} else if rp.PermissionID != "" {
			// Lazy load permission
			perm, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
			if err == nil {
				permissions = append(permissions, perm)
			}
		}
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, permissions, 600); err != nil { // 10 minutes TTL
			s.logger.Warn("Failed to cache role permissions", zap.Error(err))
		}
	}

	return permissions, nil
}

// GetRoleResources retrieves all resource permissions for a role (Model 2)
func (s *Service) GetRoleResources(ctx context.Context, roleID string) ([]*models.ResourcePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("role:%s:resources", roleID)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			if resources, ok := cached.([]*models.ResourcePermission); ok {
				s.logger.Debug("Role resources cache hit", zap.String("role_id", roleID))
				return resources, nil
			}
		}
	}

	// Get resource permissions
	resources, err := s.resourcePermissionRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get role resources",
			zap.String("role_id", roleID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get role resources: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(cacheKey, resources, 600); err != nil { // 10 minutes TTL
			s.logger.Warn("Failed to cache role resources", zap.Error(err))
		}
	}

	return resources, nil
}

// GetRolesWithPermission retrieves all roles that have a specific permission
func (s *Service) GetRolesWithPermission(ctx context.Context, permissionID string) ([]*models.Role, error) {
	if permissionID == "" {
		return nil, fmt.Errorf("permission ID is required")
	}

	// Get role-permission assignments
	rolePerms, err := s.rolePermissionRepo.GetByPermissionID(ctx, permissionID)
	if err != nil {
		s.logger.Error("Failed to get roles for permission",
			zap.String("permission_id", permissionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get roles for permission: %w", err)
	}

	// Extract unique roles
	roleMap := make(map[string]*models.Role)
	for _, rp := range rolePerms {
		if rp.Role != nil {
			roleMap[rp.Role.ID] = rp.Role
		} else if rp.RoleID != "" {
			// Lazy load role
			role, err := s.roleRepo.GetByID(ctx, rp.RoleID, nil)
			if err == nil && role != nil {
				roleMap[role.ID] = role
			}
		}
	}

	// Convert map to slice
	roles := make([]*models.Role, 0, len(roleMap))
	for _, role := range roleMap {
		roles = append(roles, role)
	}

	return roles, nil
}

// GetRolesWithResourceAccess retrieves all roles that have access to a specific resource-action
func (s *Service) GetRolesWithResourceAccess(ctx context.Context, resourceType, resourceID, action string) ([]*models.Role, error) {
	if resourceType == "" || resourceID == "" || action == "" {
		return nil, fmt.Errorf("resource type, resource ID, and action are required")
	}

	// Get resource permissions by resource and action
	resourcePerms, err := s.resourcePermissionRepo.GetByResource(ctx, resourceType, resourceID)
	if err != nil {
		s.logger.Error("Failed to get roles for resource",
			zap.String("resource_type", resourceType),
			zap.String("resource_id", resourceID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get roles for resource: %w", err)
	}

	// Filter by action and extract unique roles
	roleMap := make(map[string]*models.Role)
	for _, rp := range resourcePerms {
		if rp.Action == action && rp.RoleID != "" {
			// Lazy load role
			role, err := s.roleRepo.GetByID(ctx, rp.RoleID, nil)
			if err == nil && role != nil {
				roleMap[role.ID] = role
			}
		}
	}

	// Convert map to slice
	roles := make([]*models.Role, 0, len(roleMap))
	for _, role := range roleMap {
		roles = append(roles, role)
	}

	return roles, nil
}

// GetPermissionAssignmentHistory retrieves the assignment history for a role
func (s *Service) GetPermissionAssignmentHistory(ctx context.Context, roleID string, limit, offset int) ([]*AssignmentHistory, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// TODO: Implement assignment history tracking
	// This would require audit log queries or a separate assignment_history table
	s.logger.Warn("GetPermissionAssignmentHistory not yet implemented",
		zap.String("role_id", roleID))

	return []*AssignmentHistory{}, nil
}

// GetRolePermissionStats retrieves statistics about role-permission assignments
func (s *Service) GetRolePermissionStats(ctx context.Context, roleID string) (map[string]interface{}, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	// Get permission count (Model 1)
	permissionCount, err := s.rolePermissionRepo.CountByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to count permissions: %w", err)
	}

	// Get resource permissions (Model 2)
	resourcePerms, err := s.resourcePermissionRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource permissions: %w", err)
	}

	// Count unique resources
	resourceMap := make(map[string]bool)
	actionMap := make(map[string]int)
	for _, rp := range resourcePerms {
		key := fmt.Sprintf("%s:%s", rp.ResourceType, rp.ResourceID)
		resourceMap[key] = true
		actionMap[rp.Action]++
	}

	stats := map[string]interface{}{
		"role_id":                 roleID,
		"named_permissions_count": permissionCount,
		"resource_permissions":    len(resourcePerms),
		"unique_resources":        len(resourceMap),
		"actions_breakdown":       actionMap,
	}

	return stats, nil
}
