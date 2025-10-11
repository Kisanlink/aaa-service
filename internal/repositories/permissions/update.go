package permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// UpdateWithValidation updates a permission with validation
func (r *PermissionRepository) UpdateWithValidation(ctx context.Context, permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	if permission.GetID() == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Check if permission exists
	exists, err := r.Exists(ctx, permission.GetID())
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("permission not found with ID: %s", permission.GetID())
	}

	// If name is being updated, check for duplicates
	if permission.Name != "" {
		existing, err := r.GetByID(ctx, permission.GetID())
		if err != nil {
			return fmt.Errorf("failed to get existing permission: %w", err)
		}

		if existing.Name != permission.Name {
			nameExists, err := r.ExistsByName(ctx, permission.Name)
			if err != nil {
				return fmt.Errorf("failed to check name uniqueness: %w", err)
			}

			if nameExists {
				return fmt.Errorf("permission with name '%s' already exists", permission.Name)
			}
		}
	}

	// Update the permission
	if err := r.Update(ctx, permission); err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// ActivatePermission activates a permission
func (r *PermissionRepository) ActivatePermission(ctx context.Context, id string) error {
	permission, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	permission.IsActive = true
	return r.Update(ctx, permission)
}

// DeactivatePermission deactivates a permission
func (r *PermissionRepository) DeactivatePermission(ctx context.Context, id string) error {
	permission, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	permission.IsActive = false
	return r.Update(ctx, permission)
}

// UpdateDescription updates only the description of a permission
func (r *PermissionRepository) UpdateDescription(ctx context.Context, id, description string) error {
	permission, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	permission.Description = description
	return r.Update(ctx, permission)
}
