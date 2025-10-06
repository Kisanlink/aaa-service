package resource_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// Assign assigns a specific action on a resource to a role (Model 2: Direct assignment)
func (r *ResourcePermissionRepository) Assign(ctx context.Context, roleID, resourceType, resourceID, action string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if action == "" {
		return fmt.Errorf("action is required")
	}

	// Check if assignment already exists
	exists, err := r.HasPermission(ctx, roleID, resourceType, resourceID, action)
	if err != nil {
		return fmt.Errorf("failed to check if permission exists: %w", err)
	}

	if exists {
		return fmt.Errorf("permission already exists for role '%s' on resource '%s/%s' with action '%s'",
			roleID, resourceType, resourceID, action)
	}

	// Create the assignment
	resourcePermission := models.NewResourcePermission(resourceID, resourceType, roleID, action)

	if err := r.BaseFilterableRepository.Create(ctx, resourcePermission); err != nil {
		return fmt.Errorf("failed to assign resource permission: %w", err)
	}

	return nil
}

// AssignBatch assigns multiple resource-action permissions in batch
func (r *ResourcePermissionRepository) AssignBatch(ctx context.Context, assignments []ResourcePermissionAssignment) error {
	if len(assignments) == 0 {
		return fmt.Errorf("no assignments provided")
	}

	for i, assignment := range assignments {
		if assignment.RoleID == "" {
			return fmt.Errorf("role ID is required for assignment %d", i)
		}

		if assignment.ResourceType == "" {
			return fmt.Errorf("resource type is required for assignment %d", i)
		}

		if assignment.ResourceID == "" {
			return fmt.Errorf("resource ID is required for assignment %d", i)
		}

		if assignment.Action == "" {
			return fmt.Errorf("action is required for assignment %d", i)
		}

		// Check if already assigned
		exists, err := r.HasPermission(ctx, assignment.RoleID, assignment.ResourceType,
			assignment.ResourceID, assignment.Action)
		if err != nil {
			return fmt.Errorf("failed to check permission for assignment %d: %w", i, err)
		}

		if exists {
			// Skip already assigned permissions
			continue
		}

		// Create the assignment
		resourcePermission := models.NewResourcePermission(assignment.ResourceID, assignment.ResourceType,
			assignment.RoleID, assignment.Action)

		if err := r.BaseFilterableRepository.Create(ctx, resourcePermission); err != nil {
			return fmt.Errorf("failed to create assignment %d: %w", i, err)
		}
	}

	return nil
}

// AssignActions assigns multiple actions on a single resource to a role
func (r *ResourcePermissionRepository) AssignActions(ctx context.Context, roleID, resourceType, resourceID string, actions []string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if len(actions) == 0 {
		return fmt.Errorf("no actions provided")
	}

	// Create assignments for each action
	for _, action := range actions {
		if action == "" {
			continue
		}

		// Check if already assigned
		exists, err := r.HasPermission(ctx, roleID, resourceType, resourceID, action)
		if err != nil {
			return fmt.Errorf("failed to check permission for action '%s': %w", action, err)
		}

		if exists {
			// Skip already assigned actions
			continue
		}

		// Create the assignment
		resourcePermission := models.NewResourcePermission(resourceID, resourceType, roleID, action)

		if err := r.BaseFilterableRepository.Create(ctx, resourcePermission); err != nil {
			return fmt.Errorf("failed to assign action '%s': %w", action, err)
		}
	}

	return nil
}

// AssignWithStatus assigns a resource permission with a specific active status
func (r *ResourcePermissionRepository) AssignWithStatus(ctx context.Context, roleID, resourceType, resourceID, action string, isActive bool) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if action == "" {
		return fmt.Errorf("action is required")
	}

	// Create the assignment with status
	resourcePermission := models.NewResourcePermission(resourceID, resourceType, roleID, action)
	resourcePermission.IsActive = isActive

	if err := r.BaseFilterableRepository.Create(ctx, resourcePermission); err != nil {
		return fmt.Errorf("failed to assign resource permission: %w", err)
	}

	return nil
}
