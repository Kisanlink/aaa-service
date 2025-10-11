package role_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// Assign assigns a permission to a role (creates a role-permission relationship)
func (r *RolePermissionRepository) Assign(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Check if assignment already exists
	exists, err := r.Exists(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check if assignment exists: %w", err)
	}

	if exists {
		return fmt.Errorf("permission '%s' is already assigned to role '%s'", permissionID, roleID)
	}

	// Create the assignment
	rolePermission := models.NewRolePermission(roleID, permissionID)

	if err := r.BaseFilterableRepository.Create(ctx, rolePermission); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// AssignBatch assigns multiple permissions to a single role in batch
func (r *RolePermissionRepository) AssignBatch(ctx context.Context, roleID string, permissionIDs []string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if len(permissionIDs) == 0 {
		return fmt.Errorf("no permission IDs provided")
	}

	// Validate and create assignments
	for _, permissionID := range permissionIDs {
		if permissionID == "" {
			continue
		}

		// Check if already assigned
		exists, err := r.Exists(ctx, roleID, permissionID)
		if err != nil {
			return fmt.Errorf("failed to check if assignment exists for permission '%s': %w", permissionID, err)
		}

		if exists {
			// Skip already assigned permissions
			continue
		}

		// Create the assignment
		rolePermission := models.NewRolePermission(roleID, permissionID)

		if err := r.BaseFilterableRepository.Create(ctx, rolePermission); err != nil {
			return fmt.Errorf("failed to assign permission '%s' to role: %w", permissionID, err)
		}
	}

	return nil
}

// AssignMultipleRoles assigns a single permission to multiple roles
func (r *RolePermissionRepository) AssignMultipleRoles(ctx context.Context, roleIDs []string, permissionID string) error {
	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	if len(roleIDs) == 0 {
		return fmt.Errorf("no role IDs provided")
	}

	// Validate and create assignments
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}

		// Check if already assigned
		exists, err := r.Exists(ctx, roleID, permissionID)
		if err != nil {
			return fmt.Errorf("failed to check if assignment exists for role '%s': %w", roleID, err)
		}

		if exists {
			// Skip already assigned roles
			continue
		}

		// Create the assignment
		rolePermission := models.NewRolePermission(roleID, permissionID)

		if err := r.BaseFilterableRepository.Create(ctx, rolePermission); err != nil {
			return fmt.Errorf("failed to assign permission to role '%s': %w", roleID, err)
		}
	}

	return nil
}

// AssignWithStatus assigns a permission to a role with a specific active status
func (r *RolePermissionRepository) AssignWithStatus(ctx context.Context, roleID, permissionID string, isActive bool) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if permissionID == "" {
		return fmt.Errorf("permission ID is required")
	}

	// Check if assignment already exists
	existing, err := r.GetByRoleAndPermission(ctx, roleID, permissionID)
	if err == nil && existing != nil {
		// Update existing assignment status
		existing.IsActive = isActive
		return r.BaseFilterableRepository.Update(ctx, existing)
	}

	// Create new assignment with status
	rolePermission := models.NewRolePermission(roleID, permissionID)
	rolePermission.IsActive = isActive

	if err := r.BaseFilterableRepository.Create(ctx, rolePermission); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}
