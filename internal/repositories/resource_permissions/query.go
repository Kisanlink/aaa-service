package resource_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// GetByRoleID retrieves all resource permissions for a specific role
func (r *ResourcePermissionRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.ResourcePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource permissions for role: %w", err)
	}

	return permissions, nil
}

// GetByResource retrieves all permissions for a specific resource
func (r *ResourcePermissionRepository) GetByResource(ctx context.Context, resourceType, resourceID string) ([]*models.ResourcePermission, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource permissions: %w", err)
	}

	return permissions, nil
}

// GetByResourceType retrieves all permissions for a specific resource type
func (r *ResourcePermissionRepository) GetByResourceType(ctx context.Context, resourceType string) ([]*models.ResourcePermission, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	filter := base.NewFilterBuilder().
		Where("resource_type", base.OpEqual, resourceType).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource permissions by type: %w", err)
	}

	return permissions, nil
}

// GetByAction retrieves all permissions for a specific action
func (r *ResourcePermissionRepository) GetByAction(ctx context.Context, action string) ([]*models.ResourcePermission, error) {
	if action == "" {
		return nil, fmt.Errorf("action is required")
	}

	filter := base.NewFilterBuilder().
		Where("action", base.OpEqual, action).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource permissions by action: %w", err)
	}

	return permissions, nil
}

// FindByFilter retrieves resource permissions using a custom filter
func (r *ResourcePermissionRepository) FindByFilter(ctx context.Context, filter *base.Filter) ([]*models.ResourcePermission, error) {
	if filter == nil {
		filter = base.NewFilter()
	}

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role on a specific resource type
func (r *ResourcePermissionRepository) GetRolePermissions(ctx context.Context, roleID, resourceType string) ([]*models.ResourcePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return permissions, nil
}

// GetActive retrieves all active resource permissions for a role
func (r *ResourcePermissionRepository) GetActive(ctx context.Context, roleID string) ([]*models.ResourcePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active resource permissions: %w", err)
	}

	return permissions, nil
}

// GetInactive retrieves all inactive resource permissions for a role
func (r *ResourcePermissionRepository) GetInactive(ctx context.Context, roleID string) ([]*models.ResourcePermission, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, false).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive resource permissions: %w", err)
	}

	return permissions, nil
}

// Activate activates a resource permission
func (r *ResourcePermissionRepository) Activate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	permission := &models.ResourcePermission{}
	permission, err := r.BaseFilterableRepository.GetByID(ctx, id, permission)
	if err != nil {
		return fmt.Errorf("failed to get resource permission: %w", err)
	}

	permission.IsActive = true
	return r.BaseFilterableRepository.Update(ctx, permission)
}

// Deactivate deactivates a resource permission
func (r *ResourcePermissionRepository) Deactivate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	permission := &models.ResourcePermission{}
	permission, err := r.BaseFilterableRepository.GetByID(ctx, id, permission)
	if err != nil {
		return fmt.Errorf("failed to get resource permission: %w", err)
	}

	permission.IsActive = false
	return r.BaseFilterableRepository.Update(ctx, permission)
}

// CountByRole returns the number of resource permissions assigned to a role
func (r *ResourcePermissionRepository) CountByRole(ctx context.Context, roleID string) (int64, error) {
	if roleID == "" {
		return 0, fmt.Errorf("role ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resource permissions for role: %w", err)
	}

	return count, nil
}

// CountByResource returns the number of permissions for a specific resource
func (r *ResourcePermissionRepository) CountByResource(ctx context.Context, resourceType, resourceID string) (int64, error) {
	if resourceType == "" {
		return 0, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return 0, fmt.Errorf("resource ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Build()

	count, err := r.BaseFilterableRepository.CountWithFilter(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resource permissions: %w", err)
	}

	return count, nil
}
