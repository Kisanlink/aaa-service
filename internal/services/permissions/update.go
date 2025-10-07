package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"go.uber.org/zap"
)

// UpdatePermission updates an existing permission with validation
func (s *Service) UpdatePermission(ctx context.Context, permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	if permission.ID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Validate permission fields
	if err := s.validatePermission(permission); err != nil {
		s.logger.Error("Permission validation failed", zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get existing permission
	existing, err := s.permissionRepo.GetByID(ctx, permission.ID)
	if err != nil {
		s.logger.Error("Failed to get existing permission",
			zap.String("id", permission.ID),
			zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if permission is in use before allowing critical changes
	if s.isBreakingChange(existing, permission) {
		inUse, err := s.isPermissionInUse(ctx, permission.ID)
		if err != nil {
			s.logger.Error("Failed to check if permission is in use",
				zap.String("id", permission.ID),
				zap.Error(err))
			return fmt.Errorf("failed to check permission usage: %w", err)
		}

		if inUse {
			return fmt.Errorf("cannot make breaking changes to permission that is in use")
		}
	}

	// Check for duplicate name if name is being changed
	if existing.Name != permission.Name {
		duplicate, err := s.permissionRepo.GetByName(ctx, permission.Name)
		if err == nil && duplicate != nil && duplicate.ID != permission.ID {
			return fmt.Errorf("permission with name '%s' already exists", permission.Name)
		}
	}

	// Verify resource exists if being changed
	if permission.ResourceID != nil && *permission.ResourceID != "" {
		if existing.ResourceID == nil || *existing.ResourceID != *permission.ResourceID {
			if err := s.validateResource(ctx, *permission.ResourceID); err != nil {
				return fmt.Errorf("invalid resource: %w", err)
			}
		}
	}

	// Verify action exists if being changed
	if permission.ActionID != nil && *permission.ActionID != "" {
		if existing.ActionID == nil || *existing.ActionID != *permission.ActionID {
			if err := s.validateAction(ctx, *permission.ActionID); err != nil {
				return fmt.Errorf("invalid action: %w", err)
			}
		}
	}

	// Update permission in database
	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		s.logger.Error("Failed to update permission",
			zap.String("id", permission.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update permission: %w", err)
	}

	// Invalidate all related caches
	s.invalidatePermissionRelatedCaches(ctx, permission)
	s.invalidateAllRoleCaches(ctx, permission.ID)

	// Audit log
	if s.audit != nil {
		s.audit.LogPermissionChange(ctx, "", "update", "", permission.ID, permission.Name,
			map[string]interface{}{
				"old_name":        existing.Name,
				"new_name":        permission.Name,
				"old_description": existing.Description,
				"new_description": permission.Description,
				"old_resource_id": existing.ResourceID,
				"new_resource_id": permission.ResourceID,
				"old_action_id":   existing.ActionID,
				"new_action_id":   permission.ActionID,
			})
	}

	s.logger.Info("Permission updated successfully",
		zap.String("permission_id", permission.ID),
		zap.String("name", permission.Name))

	return nil
}

// isBreakingChange determines if the update would break existing assignments
func (s *Service) isBreakingChange(existing, updated *models.Permission) bool {
	// Name change is considered breaking
	if existing.Name != updated.Name {
		return true
	}

	// Resource ID change is breaking
	if !equalStringPtr(existing.ResourceID, updated.ResourceID) {
		return true
	}

	// Action ID change is breaking
	if !equalStringPtr(existing.ActionID, updated.ActionID) {
		return true
	}

	return false
}

// equalStringPtr compares two string pointers for equality
func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// isPermissionInUse checks if a permission is assigned to any roles
func (s *Service) isPermissionInUse(ctx context.Context, permissionID string) (bool, error) {
	count, err := s.rolePermissionRepo.CountByPermission(ctx, permissionID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// invalidateAllRoleCaches invalidates caches for all roles using this permission
func (s *Service) invalidateAllRoleCaches(ctx context.Context, permissionID string) {
	if s.cache == nil {
		return
	}

	// Get all roles using this permission
	rolePerms, err := s.rolePermissionRepo.GetByPermissionID(ctx, permissionID)
	if err != nil {
		s.logger.Warn("Failed to get roles for permission",
			zap.String("permission_id", permissionID),
			zap.Error(err))
		return
	}

	// Invalidate cache for each role
	for _, rp := range rolePerms {
		cacheKey := fmt.Sprintf("role:%s:permissions", rp.RoleID)
		if err := s.cache.Delete(cacheKey); err != nil {
			s.logger.Warn("Failed to invalidate role cache",
				zap.String("cache_key", cacheKey),
				zap.Error(err))
		}
	}
}
