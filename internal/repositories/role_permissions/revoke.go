package role_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Revoke revokes a permission from a role (deletes the role-permission relationship)
func (r *RolePermissionRepository) Revoke(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Find the assignment
	assignment, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("permission assignment not found: %w", err)
	}

	// Delete the assignment
	if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}

	return nil
}

// RevokeBatch revokes multiple permissions from a single role in batch
func (r *RolePermissionRepository) RevokeBatch(ctx context.Context, roleID string, permissionIDs []string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if len(permissionIDs) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}

	// Revoke each permission
	for _, permissionID := range permissionIDs {
		if permissionID == "" {
			continue
		}

		// Find the assignment
		assignment, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
		if err != nil {
			// Skip if assignment doesn't exist
			continue
		}

		// Delete the assignment
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission '%s' from role: %w", permissionID, err)
		}
	}

	return nil
}

// RevokeAll revokes all permissions from a role
func (r *RolePermissionRepository) RevokeAll(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Get all assignments for the role
	assignments, err := r.GetByRoleID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Delete all assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission '%s': %w", assignment.PermissionID, err)
		}
	}

	return nil
}

// RevokeByPermission revokes a specific permission from all roles that have it
func (r *RolePermissionRepository) RevokeByPermission(ctx context.Context, permissionID string) error {
	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Get all assignments for the permission
	assignments, err := r.GetByPermissionID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get permission assignments: %w", err)
	}

	// Delete all assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission from role '%s': %w", assignment.RoleID, err)
		}
	}

	return nil
}

// SoftRevoke marks a role-permission assignment as inactive instead of deleting it
func (r *RolePermissionRepository) SoftRevoke(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Find the assignment
	assignment, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("permission assignment not found: %w", err)
	}

	// Mark as inactive
	assignment.IsActive = false

	if err := r.BaseFilterableRepository.Update(ctx, assignment); err != nil {
		return fmt.Errorf("failed to soft revoke permission from role: %w", err)
	}

	return nil
}

// Activate activates a role-permission assignment
func (r *RolePermissionRepository) Activate(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Find the assignment
	assignment, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("permission assignment not found: %w", err)
	}

	// Mark as active
	assignment.IsActive = true

	if err := r.BaseFilterableRepository.Update(ctx, assignment); err != nil {
		return fmt.Errorf("failed to activate permission assignment: %w", err)
	}

	return nil
}

// Deactivate deactivates a role-permission assignment
func (r *RolePermissionRepository) Deactivate(ctx context.Context, roleID, permissionID string) error {
	return r.SoftRevoke(ctx, roleID, permissionID)
}

// RevokeInactive removes all inactive role-permission assignments for a role
func (r *RolePermissionRepository) RevokeInactive(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Get all assignments for the role
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, false).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get inactive assignments: %w", err)
	}

	// Delete all inactive assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke inactive permission '%s': %w", assignment.PermissionID, err)
		}
	}

	return nil
}
