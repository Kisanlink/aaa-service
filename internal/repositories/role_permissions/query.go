package role_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// GetByRoleID retrieves all permission assignments for a specific role
func (r *RolePermissionRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.RolePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for role: %w", err)
	}

	return assignments, nil
}

// GetByPermissionID retrieves all role assignments for a specific permission
func (r *RolePermissionRepository) GetByPermissionID(ctx context.Context, permissionID string) ([]*models.RolePermission, error) {
	if permissionID == "" {
		return nil, fmt.Errorf("permission ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("permission_id", base.OpEqual, permissionID).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for permission: %w", err)
	}

	return assignments, nil
}

// GetByRoleAndPermission retrieves a specific role-permission assignment
func (r *RolePermissionRepository) GetByRoleAndPermission(ctx context.Context, roleID, permissionID string) (*models.RolePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return nil, fmt.Errorf("permission ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("permission_id", base.OpEqual, permissionID).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get role-permission assignment: %w", err)
	}

	if len(assignments) == 0 {
		return nil, fmt.Errorf("role-permission assignment not found")
	}

	return assignments[0], nil
}

// Exists checks if a role-permission assignment exists
func (r *RolePermissionRepository) Exists(ctx context.Context, roleID, permissionID string) (bool, error) {
	if roleID == "" {
		return false, fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return false, fmt.Errorf("permission ID is required")
	}

	_, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// CountByRole returns the number of permissions assigned to a role
func (r *RolePermissionRepository) CountByRole(ctx context.Context, roleID string) (int64, error) {
	if roleID == "" {
		return 0, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count permissions for role: %w", err)
	}

	return count, nil
}

// CountByPermission returns the number of roles assigned to a permission
func (r *RolePermissionRepository) CountByPermission(ctx context.Context, permissionID string) (int64, error) {
	if permissionID == "" {
		return 0, fmt.Errorf("permission ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("permission_id", base.OpEqual, permissionID).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles for permission: %w", err)
	}

	return count, nil
}

// GetPermissionsByRoles retrieves all permission assignments for multiple roles
func (r *RolePermissionRepository) GetPermissionsByRoles(ctx context.Context, roleIDs []string) ([]*models.RolePermission, error) {
	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("no role IDs provided")
	}

	filter := base.NewFilterBuilder().
		WhereIn("role_id", convertToInterfaces(roleIDs)).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for roles: %w", err)
	}

	return assignments, nil
}

// GetRolesByPermissions retrieves all role assignments for multiple permissions
func (r *RolePermissionRepository) GetRolesByPermissions(ctx context.Context, permissionIDs []string) ([]*models.RolePermission, error) {
	if len(permissionIDs) == 0 {
		return nil, fmt.Errorf("no permission IDs provided")
	}

	filter := base.NewFilterBuilder().
		WhereIn("permission_id", convertToInterfaces(permissionIDs)).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for permissions: %w", err)
	}

	return assignments, nil
}

// GetActive retrieves all active permission assignments for a role
func (r *RolePermissionRepository) GetActive(ctx context.Context, roleID string) ([]*models.RolePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active permissions for role: %w", err)
	}

	return assignments, nil
}

// GetInactive retrieves all inactive permission assignments for a role
func (r *RolePermissionRepository) GetInactive(ctx context.Context, roleID string) ([]*models.RolePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, false).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive permissions for role: %w", err)
	}

	return assignments, nil
}

// GetActiveByPermission retrieves all active role assignments for a permission
func (r *RolePermissionRepository) GetActiveByPermission(ctx context.Context, permissionID string) ([]*models.RolePermission, error) {
	if permissionID == "" {
		return nil, fmt.Errorf("permission ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("permission_id", base.OpEqual, permissionID).
		Where("is_active", base.OpEqual, true).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active roles for permission: %w", err)
	}

	return assignments, nil
}

// FindByFilter retrieves role-permission assignments using a custom filter
func (r *RolePermissionRepository) FindByFilter(ctx context.Context, filter *base.Filter) ([]*models.RolePermission, error) {
	if filter == nil {
		filter = base.NewFilter()
	}

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find role-permission assignments: %w", err)
	}

	return assignments, nil
}

// Helper function to convert string slice to interface slice for WhereIn
func convertToInterfaces(strs []string) []interface{} {
	interfaces := make([]interface{}, len(strs))
	for i, str := range strs {
		interfaces[i] = str
	}
	return interfaces
}
