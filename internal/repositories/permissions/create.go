package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CreateWithValidation creates a new permission with validation
func (r *PermissionRepository) CreateWithValidation(ctx context.Context, permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	if permission.Name == "" {
		return fmt.Errorf("permission name is required")
	}

	// Check if permission with same name already exists
	exists, err := r.ExistsByName(ctx, permission.Name)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}

	if exists {
		return fmt.Errorf("permission with name '%s' already exists", permission.Name)
	}

	// Create the permission
	if err := r.Create(ctx, permission); err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// CreateBatch creates multiple permissions in a batch
func (r *PermissionRepository) CreateBatch(ctx context.Context, permissions []*models.Permission) error {
	if len(permissions) == 0 {
		return fmt.Errorf("no permissions to create")
	}

	for _, permission := range permissions {
		if err := r.CreateWithValidation(ctx, permission); err != nil {
			return fmt.Errorf("failed to create permission '%s': %w", permission.Name, err)
		}
	}

	return nil
}

// ExistsByName checks if a permission exists with the given name
func (r *PermissionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	filter := base.NewFilterBuilder().
		Where("name", base.OpEqual, name).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, err
	}

	return len(permissions) > 0, nil
}
