package permissions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// DeletePermission deletes a permission with safety checks
func (s *Service) DeletePermission(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Get existing permission
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get permission for deletion",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if permission is assigned to any roles
	inUse, err := s.isPermissionInUse(ctx, id)
	if err != nil {
		s.logger.Error("Failed to check if permission is in use",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to check permission usage: %w", err)
	}

	if inUse {
		// Get count for better error message
		count, _ := s.rolePermissionRepo.CountByPermission(ctx, id)
		return fmt.Errorf("cannot delete permission that is assigned to %d role(s). Revoke all assignments first", count)
	}

	// Perform soft delete
	if err := s.permissionRepo.SoftDelete(ctx, id, "system"); err != nil {
		s.logger.Error("Failed to delete permission",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// Invalidate all related caches
	s.invalidatePermissionRelatedCaches(ctx, permission)

	// Audit log
	if s.audit != nil {
		s.audit.LogPermissionChange(ctx, "", "delete", "", id, permission.Name,
			map[string]interface{}{
				"name":        permission.Name,
				"description": permission.Description,
				"resource_id": permission.ResourceID,
				"action_id":   permission.ActionID,
			})
	}

	s.logger.Info("Permission deleted successfully",
		zap.String("permission_id", id),
		zap.String("name", permission.Name))

	return nil
}

// DeletePermissionWithCascade deletes a permission and all its role assignments
func (s *Service) DeletePermissionWithCascade(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Get existing permission
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get permission for cascade deletion",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Get all role assignments for this permission
	rolePerms, err := s.rolePermissionRepo.GetByPermissionID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get role assignments for permission",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to get role assignments: %w", err)
	}

	// Delete all role-permission assignments
	for _, rp := range rolePerms {
		if err := s.rolePermissionRepo.Revoke(ctx, rp.RoleID, id); err != nil {
			s.logger.Error("Failed to revoke permission from role",
				zap.String("role_id", rp.RoleID),
				zap.String("permission_id", id),
				zap.Error(err))
			// Continue with other deletions even if one fails
		}
	}

	// Delete the permission
	if err := s.permissionRepo.SoftDelete(ctx, id, "system"); err != nil {
		s.logger.Error("Failed to delete permission",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// Invalidate all related caches
	s.invalidatePermissionRelatedCaches(ctx, permission)
	s.invalidateAllRoleCaches(ctx, id)

	// Audit log
	if s.audit != nil {
		s.audit.LogPermissionChange(ctx, "", "cascade_delete", "", id, permission.Name,
			map[string]interface{}{
				"name":                permission.Name,
				"description":         permission.Description,
				"revoked_assignments": len(rolePerms),
			})
	}

	s.logger.Info("Permission deleted with cascade successfully",
		zap.String("permission_id", id),
		zap.String("name", permission.Name),
		zap.Int("revoked_assignments", len(rolePerms)))

	return nil
}
