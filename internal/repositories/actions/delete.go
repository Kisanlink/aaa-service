package actions

import (
	"context"
	"fmt"
)

// Delete deletes an action by ID using the database manager (hard delete)
func (r *ActionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	// Check if action exists
	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action for deletion: %w", err)
	}

	// Prevent deletion of static actions
	if action.IsStatic {
		return fmt.Errorf("cannot delete static action '%s'", action.Name)
	}

	// Use the BaseFilterableRepository which properly handles table names
	if err := r.BaseFilterableRepository.Delete(ctx, id, action); err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}

	return nil
}

// SoftDelete soft deletes an action by ID using the base repository
func (r *ActionRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	if deletedBy == "" {
		return fmt.Errorf("deleted_by is required for soft delete")
	}

	// Check if action exists
	action, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get action for soft deletion: %w", err)
	}

	// Prevent deletion of static actions
	if action.IsStatic {
		return fmt.Errorf("cannot delete static action '%s'", action.Name)
	}

	// Soft delete using base repository
	if err := r.BaseFilterableRepository.SoftDelete(ctx, id, deletedBy); err != nil {
		return fmt.Errorf("failed to soft delete action: %w", err)
	}

	return nil
}

// Restore restores a soft-deleted action using the base repository
func (r *ActionRepository) Restore(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("action ID is required")
	}

	// Check if action exists (including soft-deleted)
	exists, err := r.BaseFilterableRepository.ExistsWithDeleted(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if action exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("action with ID '%s' not found", id)
	}

	// Restore using base repository
	if err := r.BaseFilterableRepository.Restore(ctx, id); err != nil {
		return fmt.Errorf("failed to restore action: %w", err)
	}

	return nil
}

// DeleteBatch deletes multiple actions by IDs
func (r *ActionRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no action IDs provided")
	}

	for _, id := range ids {
		if err := r.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete action '%s': %w", id, err)
		}
	}

	return nil
}

// SoftDeleteBatch soft deletes multiple actions by IDs
func (r *ActionRepository) SoftDeleteBatch(ctx context.Context, ids []string, deletedBy string) error {
	if len(ids) == 0 {
		return fmt.Errorf("no action IDs provided")
	}

	if deletedBy == "" {
		return fmt.Errorf("deleted_by is required for soft delete")
	}

	for _, id := range ids {
		if err := r.SoftDelete(ctx, id, deletedBy); err != nil {
			return fmt.Errorf("failed to soft delete action '%s': %w", id, err)
		}
	}

	return nil
}
