package permissions

import (
	"context"
	"fmt"
)

// DeleteWithValidation deletes a permission with validation
func (r *PermissionRepository) DeleteWithValidation(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Check if permission exists
	exists, err := r.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("permission not found with ID: %s", id)
	}

	// Delete the permission
	if err := r.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// SoftDeleteWithValidation soft deletes a permission with validation
func (r *PermissionRepository) SoftDeleteWithValidation(ctx context.Context, id, deletedBy string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	if deletedBy == "" {
		return fmt.Errorf("deletedBy is required for soft delete")
	}

	// Check if permission exists
	exists, err := r.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("permission not found with ID: %s", id)
	}

	// Soft delete the permission
	if err := r.SoftDelete(ctx, id, deletedBy); err != nil {
		return fmt.Errorf("failed to soft delete permission: %w", err)
	}

	return nil
}

// RestoreWithValidation restores a soft-deleted permission with validation
func (r *PermissionRepository) RestoreWithValidation(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Check if permission exists (including deleted)
	exists, err := r.ExistsWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("permission not found with ID: %s", id)
	}

	// Restore the permission
	if err := r.Restore(ctx, id); err != nil {
		return fmt.Errorf("failed to restore permission: %w", err)
	}

	return nil
}

// DeleteBatch deletes multiple permissions by IDs
func (r *PermissionRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}

	for _, id := range ids {
		if err := r.DeleteWithValidation(ctx, id); err != nil {
			return fmt.Errorf("failed to delete permission '%s': %w", id, err)
		}
	}

	return nil
}
