package resource_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Revoke revokes a specific action on a resource from a role
func (r *ResourcePermissionRepository) Revoke(ctx context.Context, roleID, resourceType, resourceID, action string) error {
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

	// Find the assignment
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Where("action", base.OpEqual, action).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find resource permission: %w", err)
	}

	if len(assignments) == 0 {
		return fmt.Errorf("resource permission not found")
	}

	// Delete the assignment
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke resource permission: %w", err)
		}
	}

	return nil
}

// RevokeBatch revokes multiple actions on a single resource from a role
func (r *ResourcePermissionRepository) RevokeBatch(ctx context.Context, roleID, resourceType, resourceID string, actions []string) error {
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

	// Revoke each action
	for _, action := range actions {
		if action == "" {
			continue
		}

		// Find and delete the assignment
		if err := r.Revoke(ctx, roleID, resourceType, resourceID, action); err != nil {
			// Log but continue with other actions
			continue
		}
	}

	return nil
}

// RevokeAllForRole revokes all resource permissions for a specific role
func (r *ResourcePermissionRepository) RevokeAllForRole(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Get all permissions for the role
	assignments, err := r.GetByRoleID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Delete all assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}
	}

	return nil
}

// RevokeAllForResource revokes all permissions for a specific resource
func (r *ResourcePermissionRepository) RevokeAllForResource(ctx context.Context, resourceType, resourceID string) error {
	if resourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	// Get all permissions for the resource
	assignments, err := r.GetByResource(ctx, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("failed to get resource permissions: %w", err)
	}

	// Delete all assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}
	}

	return nil
}

// RevokeByResourceType revokes all permissions for a specific resource type
func (r *ResourcePermissionRepository) RevokeByResourceType(ctx context.Context, resourceType string) error {
	if resourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	// Get all permissions for the resource type
	assignments, err := r.GetByResourceType(ctx, resourceType)
	if err != nil {
		return fmt.Errorf("failed to get resource permissions: %w", err)
	}

	// Delete all assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke permission: %w", err)
		}
	}

	return nil
}

// SoftRevoke marks a resource permission as inactive instead of deleting it
func (r *ResourcePermissionRepository) SoftRevoke(ctx context.Context, roleID, resourceType, resourceID, action string) error {
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

	// Find the assignment
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Where("action", base.OpEqual, action).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find resource permission: %w", err)
	}

	if len(assignments) == 0 {
		return fmt.Errorf("resource permission not found")
	}

	// Mark as inactive
	for _, assignment := range assignments {
		assignment.IsActive = false
		if err := r.BaseFilterableRepository.Update(ctx, assignment); err != nil {
			return fmt.Errorf("failed to soft revoke resource permission: %w", err)
		}
	}

	return nil
}

// RevokeInactive removes all inactive resource permissions for a role
func (r *ResourcePermissionRepository) RevokeInactive(ctx context.Context, roleID string) error {
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Get all inactive permissions for the role
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("is_active", base.OpEqual, false).
		Build()

	assignments, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get inactive permissions: %w", err)
	}

	// Delete all inactive assignments
	for _, assignment := range assignments {
		if err := r.BaseFilterableRepository.Delete(ctx, assignment.ID, assignment); err != nil {
			return fmt.Errorf("failed to revoke inactive permission: %w", err)
		}
	}

	return nil
}
