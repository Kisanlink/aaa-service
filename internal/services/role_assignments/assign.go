package role_assignments

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/repositories/resource_permissions"
	"go.uber.org/zap"
)

// AssignPermissionToRole assigns a single permission to a role (Model 1)
func (s *Service) AssignPermissionToRole(ctx context.Context, roleID, permissionID string, assignedBy string) error {
	if roleID == "" || permissionID == "" {
		return fmt.Errorf("role ID and permission ID are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if role is soft-deleted
	if role.DeletedAt != nil {
		s.logger.Error("Cannot assign permission to deleted role",
			zap.String("role_id", roleID),
			zap.String("role_name", role.Name))
		return fmt.Errorf("cannot assign permission to deleted role '%s'", role.Name)
	}

	// Verify permission exists
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		s.logger.Error("Permission not found", zap.String("permission_id", permissionID), zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if already assigned
	exists, err := s.rolePermissionRepo.Exists(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		s.logger.Warn("Permission already assigned to role",
			zap.String("role_id", roleID),
			zap.String("permission_id", permissionID))
		return fmt.Errorf("permission already assigned to role")
	}

	// Assign permission
	if err := s.rolePermissionRepo.Assign(ctx, roleID, permissionID); err != nil {
		s.logger.Error("Failed to assign permission",
			zap.String("role_id", roleID),
			zap.String("permission_id", permissionID),
			zap.Error(err))
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, assignedBy, "", roleID, "assign_permission", true,
			map[string]interface{}{
				"role_name":       role.Name,
				"permission_id":   permissionID,
				"permission_name": permission.Name,
			})
	}

	s.logger.Info("Permission assigned to role",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID),
		zap.String("assigned_by", assignedBy))

	return nil
}

// AssignPermissionsToRole assigns multiple permissions to a role (batch operation)
func (s *Service) AssignPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string, assignedBy string) error {
	if roleID == "" || len(permissionIDs) == 0 {
		return fmt.Errorf("role ID and permission IDs are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if role is soft-deleted
	if role.DeletedAt != nil {
		s.logger.Error("Cannot assign permissions to deleted role",
			zap.String("role_id", roleID),
			zap.String("role_name", role.Name),
			zap.Int("permission_count", len(permissionIDs)))
		return fmt.Errorf("cannot assign permissions to deleted role '%s'", role.Name)
	}

	// Verify all permissions exist
	for _, permID := range permissionIDs {
		if _, err := s.permissionRepo.GetByID(ctx, permID); err != nil {
			s.logger.Error("Permission not found", zap.String("permission_id", permID), zap.Error(err))
			return fmt.Errorf("permission %s not found: %w", permID, err)
		}
	}

	// Batch assign
	if err := s.rolePermissionRepo.AssignBatch(ctx, roleID, permissionIDs); err != nil {
		s.logger.Error("Failed to assign permissions",
			zap.String("role_id", roleID),
			zap.Int("count", len(permissionIDs)),
			zap.Error(err))
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, assignedBy, "", roleID, "assign_permissions_batch", true,
			map[string]interface{}{
				"role_name":        role.Name,
				"permission_count": len(permissionIDs),
				"permission_ids":   permissionIDs,
			})
	}

	s.logger.Info("Permissions assigned to role (batch)",
		zap.String("role_id", roleID),
		zap.Int("count", len(permissionIDs)),
		zap.String("assigned_by", assignedBy))

	return nil
}

// AssignPermissionToMultipleRoles assigns a permission to multiple roles
func (s *Service) AssignPermissionToMultipleRoles(ctx context.Context, permissionID string, roleIDs []string, assignedBy string) error {
	if permissionID == "" || len(roleIDs) == 0 {
		return fmt.Errorf("permission ID and role IDs are required")
	}

	// Verify permission exists
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		s.logger.Error("Permission not found", zap.String("permission_id", permissionID), zap.Error(err))
		return fmt.Errorf("permission not found: %w", err)
	}

	// Assign to each role
	successCount := 0
	failedRoles := []string{}

	for _, roleID := range roleIDs {
		if err := s.rolePermissionRepo.Assign(ctx, roleID, permissionID); err != nil {
			s.logger.Warn("Failed to assign permission to role",
				zap.String("role_id", roleID),
				zap.String("permission_id", permissionID),
				zap.Error(err))
			failedRoles = append(failedRoles, roleID)
			continue
		}

		// Invalidate cache for each role
		s.invalidateRoleCache(ctx, roleID)
		successCount++
	}

	// Audit log
	if s.audit != nil {
		s.audit.LogPermissionChange(ctx, assignedBy, "assign_to_multiple_roles", "", permissionID, permission.Name,
			map[string]interface{}{
				"permission_name": permission.Name,
				"total_roles":     len(roleIDs),
				"success_count":   successCount,
				"failed_count":    len(failedRoles),
				"failed_roles":    failedRoles,
			})
	}

	s.logger.Info("Permission assigned to multiple roles",
		zap.String("permission_id", permissionID),
		zap.Int("total", len(roleIDs)),
		zap.Int("success", successCount),
		zap.Int("failed", len(failedRoles)))

	if len(failedRoles) > 0 {
		return fmt.Errorf("failed to assign permission to %d role(s)", len(failedRoles))
	}

	return nil
}

// AssignResourceActionToRole assigns a resource-action pair to a role (Model 2)
func (s *Service) AssignResourceActionToRole(ctx context.Context, roleID, resourceType, resourceID, action string, assignedBy string) error {
	if roleID == "" || resourceType == "" || resourceID == "" || action == "" {
		return fmt.Errorf("role ID, resource type, resource ID, and action are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if role is soft-deleted
	if role.DeletedAt != nil {
		s.logger.Error("Cannot assign resource-action to deleted role",
			zap.String("role_id", roleID),
			zap.String("role_name", role.Name),
			zap.String("resource_type", resourceType))
		return fmt.Errorf("cannot assign resource-action to deleted role '%s'", role.Name)
	}

	// Assign resource-action
	if err := s.resourcePermissionRepo.Assign(ctx, roleID, resourceType, resourceID, action); err != nil {
		s.logger.Error("Failed to assign resource-action",
			zap.String("role_id", roleID),
			zap.String("resource", resourceType),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("failed to assign resource-action: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, assignedBy, "", roleID, "assign_resource_action", true,
			map[string]interface{}{
				"role_name":     role.Name,
				"resource_type": resourceType,
				"resource_id":   resourceID,
				"action":        action,
			})
	}

	s.logger.Info("Resource-action assigned to role",
		zap.String("role_id", roleID),
		zap.String("resource", resourceType),
		zap.String("action", action),
		zap.String("assigned_by", assignedBy))

	return nil
}

// AssignResourceActionsToRole assigns multiple resource-actions to a role
func (s *Service) AssignResourceActionsToRole(ctx context.Context, roleID string, assignments []ResourceActionAssignment, assignedBy string) error {
	if roleID == "" || len(assignments) == 0 {
		return fmt.Errorf("role ID and assignments are required")
	}

	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID, nil)
	if err != nil {
		s.logger.Error("Role not found", zap.String("role_id", roleID), zap.Error(err))
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if role is soft-deleted
	if role.DeletedAt != nil {
		s.logger.Error("Cannot assign resource-actions to deleted role",
			zap.String("role_id", roleID),
			zap.String("role_name", role.Name),
			zap.Int("assignment_count", len(assignments)))
		return fmt.Errorf("cannot assign resource-actions to deleted role '%s'", role.Name)
	}

	// Build batch assignments
	var batchAssignments []resource_permissions.ResourcePermissionAssignment
	for _, assignment := range assignments {
		for _, action := range assignment.Actions {
			batchAssignments = append(batchAssignments, resource_permissions.ResourcePermissionAssignment{
				RoleID:       roleID,
				ResourceType: assignment.ResourceType,
				ResourceID:   assignment.ResourceID,
				Action:       action,
			})
		}
	}

	// Batch assign
	if err := s.resourcePermissionRepo.AssignBatch(ctx, batchAssignments); err != nil {
		s.logger.Error("Failed to assign resource-actions",
			zap.String("role_id", roleID),
			zap.Int("count", len(batchAssignments)),
			zap.Error(err))
		return fmt.Errorf("failed to assign resource-actions: %w", err)
	}

	// Invalidate caches
	s.invalidateRoleCache(ctx, roleID)

	// Audit log
	if s.audit != nil {
		s.audit.LogRoleOperation(ctx, assignedBy, "", roleID, "assign_resource_actions_batch", true,
			map[string]interface{}{
				"role_name":        role.Name,
				"assignment_count": len(batchAssignments),
				"assignments":      assignments,
			})
	}

	s.logger.Info("Resource-actions assigned to role (batch)",
		zap.String("role_id", roleID),
		zap.Int("count", len(batchAssignments)),
		zap.String("assigned_by", assignedBy))

	return nil
}

// invalidateRoleCache invalidates all cached data for a role
func (s *Service) invalidateRoleCache(ctx context.Context, roleID string) {
	if s.cache == nil {
		return
	}

	cacheKey := fmt.Sprintf("role:%s:permissions", roleID)
	if err := s.cache.Delete(cacheKey); err != nil {
		s.logger.Warn("Failed to invalidate role cache",
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	}

	// Also invalidate resource permissions cache
	resourceCacheKey := fmt.Sprintf("role:%s:resources", roleID)
	if err := s.cache.Delete(resourceCacheKey); err != nil {
		s.logger.Warn("Failed to invalidate role resources cache",
			zap.String("cache_key", resourceCacheKey),
			zap.Error(err))
	}
}
