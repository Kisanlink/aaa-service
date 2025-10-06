package role_assignments

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// RevokePermissionFromRole revokes a single permission from a role
func (s *Service) RevokePermissionFromRole(ctx context.Context, roleID, permissionID string, revokedBy string) error {
	if roleID == "" || permissionID == "" {
		return fmt.Errorf("role ID and permission ID are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Verify permission exists
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		s.logger.Error("Permission not found", zap.String("permission_id", permissionID), zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if assignment exists
	exists, err := s.rolePermissionRepo.Exists(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check assignment: %w", err)
	}
	if !exists {
		s.logger.Warn("Permission not assigned to role",
			zap.String("role_id", roleID),
			zap.String("permission_id", permissionID))
		return fmt.Errorf("permission not assigned to role")
	}

	// Revoke permission
	if err := s.rolePermissionRepo.Revoke(ctx, roleID, permissionID); err != nil {
		s.logger.Error("Failed to revoke permission",
			zap.String("role_id", roleID),
			zap.String("permission_id", permissionID),
			zap.Error(err))
		return fmt.Errorf("failed to revoke permission: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, revokedBy, "", roleID, "revoke_permission", true,
			map[string]interface{}{
				"role_name":       role.Name,
				"permission_id":   permissionID,
				"permission_name": permission.Name,
			})
	}

	s.logger.Info("Permission revoked from role",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID),
		zap.String("revoked_by", revokedBy))

	return nil
}

// RevokePermissionsFromRole revokes multiple permissions from a role
func (s *Service) RevokePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string, revokedBy string) error {
	if roleID == "" || len(permissionIDs) == 0 {
		return fmt.Errorf("role ID and permission IDs are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Batch revoke
	if err := s.rolePermissionRepo.RevokeBatch(ctx, roleID, permissionIDs); err != nil {
		s.logger.Error("Failed to revoke permissions",
			zap.String("role_id", roleID),
			zap.Int("count", len(permissionIDs)),
			zap.Error(err))
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, revokedBy, "", roleID, "revoke_permissions_batch", true,
			map[string]interface{}{
				"role_name":        role.Name,
				"permission_count": len(permissionIDs),
				"permission_ids":   permissionIDs,
			})
	}

	s.logger.Info("Permissions revoked from role (batch)",
		zap.String("role_id", roleID),
		zap.Int("count", len(permissionIDs)),
		zap.String("revoked_by", revokedBy))

	return nil
}

// RevokeResourceActionFromRole revokes a resource-action from a role
func (s *Service) RevokeResourceActionFromRole(ctx context.Context, roleID, resourceType, resourceID, action string, revokedBy string) error {
	if roleID == "" || resourceType == "" || resourceID == "" || action == "" {
		return fmt.Errorf("role ID, resource type, resource ID, and action are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Revoke resource-action
	if err := s.resourcePermissionRepo.Revoke(ctx, roleID, resourceType, resourceID, action); err != nil {
		s.logger.Error("Failed to revoke resource-action",
			zap.String("role_id", roleID),
			zap.String("resource", resourceType),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("failed to revoke resource-action: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, revokedBy, "", roleID, "revoke_resource_action", true,
			map[string]interface{}{
				"role_name":     role.Name,
				"resource_type": resourceType,
				"resource_id":   resourceID,
				"action":        action,
			})
	}

	s.logger.Info("Resource-action revoked from role",
		zap.String("role_id", roleID),
		zap.String("resource", resourceType),
		zap.String("action", action),
		zap.String("revoked_by", revokedBy))

	return nil
}

// RevokeAllPermissionsFromRole revokes all permissions from a role
func (s *Service) RevokeAllPermissionsFromRole(ctx context.Context, roleID string, revokedBy string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Get current permission count for audit
	permissionCount, _ := s.rolePermissionRepo.CountByRole(ctx, roleID)

	// Revoke all role-permissions (Model 1)
	if err := s.rolePermissionRepo.RevokeAll(ctx, roleID); err != nil {
		s.logger.Error("Failed to revoke all permissions",
			zap.String("role_id", roleID),
			zap.Error(err))
		return fmt.Errorf("failed to revoke all permissions: %w", err)
	}

	// Revoke all resource-permissions (Model 2)
	if err := s.resourcePermissionRepo.RevokeAllForRole(ctx, roleID); err != nil {
		s.logger.Error("Failed to revoke all resource permissions",
			zap.String("role_id", roleID),
			zap.Error(err))
		return fmt.Errorf("failed to revoke all resource permissions: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, revokedBy, "", roleID, "revoke_all_permissions", true,
			map[string]interface{}{
				"role_name":        role.Name,
				"permission_count": permissionCount,
			})
	}

	s.logger.Info("All permissions revoked from role",
		zap.String("role_id", roleID),
		zap.Int64("count", permissionCount),
		zap.String("revoked_by", revokedBy))

	return nil
}
