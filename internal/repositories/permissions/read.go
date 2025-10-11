package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// GetByName retrieves a permission by name
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}

	if len(permissions) == 0 {
		return nil, fmt.Errorf("permission not found with name: %s", name)
	}

	return permissions[0], nil
}

// GetByResourceID retrieves permissions by resource ID
func (r *PermissionRepository) GetByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("resource_id", base.OpEqual, resourceID).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByActionID retrieves permissions by action ID
func (r *PermissionRepository) GetByActionID(ctx context.Context, actionID string, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("action_id", base.OpEqual, actionID).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// GetByResourceAndAction retrieves a permission by resource and action IDs
func (r *PermissionRepository) GetByResourceAndAction(ctx context.Context, resourceID, actionID string) (*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("resource_id", base.OpEqual, resourceID).
		Where("action_id", base.OpEqual, actionID).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by resource and action: %w", err)
	}

	if len(permissions) == 0 {
		return nil, fmt.Errorf("permission not found for resource %s and action %s", resourceID, actionID)
	}

	return permissions[0], nil
}

// GetActive retrieves all active permissions with pagination
func (r *PermissionRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// Search searches permissions by name using LIKE operator
func (r *PermissionRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpContains, query).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}

// FindByFilter retrieves permissions using a custom filter
func (r *PermissionRepository) FindByFilter(ctx context.Context, filter *base.Filter) ([]*models.Permission, error) {
	return r.BaseFilterableRepository.Find(ctx, filter)
}

// CountByResourceID returns the count of permissions for a specific resource
func (r *PermissionRepository) CountByResourceID(ctx context.Context, resourceID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("resource_id", base.OpEqual, resourceID).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// CountByActionID returns the count of permissions for a specific action
func (r *PermissionRepository) CountByActionID(ctx context.Context, actionID string) (int64, error) {
	filter := base.NewFilterBuilder().
		Where("action_id", base.OpEqual, actionID).
		Build()

	return r.BaseFilterableRepository.CountWithFilter(ctx, filter)
}

// GetActiveByResource retrieves active permissions for a specific resource
func (r *PermissionRepository) GetActiveByResource(ctx context.Context, resourceID string, limit, offset int) ([]*models.Permission, error) {
	filter := base.NewFilterBuilder().
		Where("resource_id", base.OpEqual, resourceID).
		Where("is_active", base.OpEqual, true).
		Limit(limit, offset).
		Build()

	return r.BaseFilterableRepository.Find(ctx, filter)
}
