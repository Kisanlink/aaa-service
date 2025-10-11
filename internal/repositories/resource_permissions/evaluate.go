package resource_permissions

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// HasPermission checks if a role has a specific permission on a resource
// This is the core permission evaluation method for Model 2 (Direct assignment)
func (r *ResourcePermissionRepository) HasPermission(ctx context.Context, roleID, resourceType, resourceID, action string) (bool, error) {
	if roleID == "" {
		return false, fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return false, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return false, fmt.Errorf("resource ID is required")
	}

	if action == "" {
		return false, fmt.Errorf("action is required")
	}

	// Check for exact match
	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Where("action", base.OpEqual, action).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return len(permissions) > 0, nil
}

// CheckMultiplePermissions checks if any of the provided roles has the specified permission
// This is useful for hierarchical role checking
func (r *ResourcePermissionRepository) CheckMultiplePermissions(ctx context.Context, roleIDs []string, resourceType, resourceID, action string) (bool, error) {
	if len(roleIDs) == 0 {
		return false, fmt.Errorf("no role IDs provided")
	}

	if resourceType == "" {
		return false, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return false, fmt.Errorf("resource ID is required")
	}

	if action == "" {
		return false, fmt.Errorf("action is required")
	}

	// Check if any role has the permission
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}

		hasPermission, err := r.HasPermission(ctx, roleID, resourceType, resourceID, action)
		if err != nil {
			continue // Skip errors and check next role
		}

		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// GetAllowedActions retrieves all actions allowed for a role on a specific resource
func (r *ResourcePermissionRepository) GetAllowedActions(ctx context.Context, roleID, resourceType, resourceID string) ([]string, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed actions: %w", err)
	}

	actions := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		actions = append(actions, permission.Action)
	}

	return actions, nil
}

// GetAllowedActionsForRoles retrieves all actions allowed for multiple roles on a resource
func (r *ResourcePermissionRepository) GetAllowedActionsForRoles(ctx context.Context, roleIDs []string, resourceType, resourceID string) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, fmt.Errorf("no role IDs provided")
	}

	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	if resourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}

	// Convert role IDs to interfaces for WhereIn
	roleIDInterfaces := make([]interface{}, len(roleIDs))
	for i, roleID := range roleIDs {
		roleIDInterfaces[i] = roleID
	}

	filter := base.NewFilterBuilder().
		WhereIn("role_id", roleIDInterfaces).
		Where("resource_type", base.OpEqual, resourceType).
		Where("resource_id", base.OpEqual, resourceID).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed actions: %w", err)
	}

	// Use map to deduplicate actions
	actionMap := make(map[string]bool)
	for _, permission := range permissions {
		actionMap[permission.Action] = true
	}

	actions := make([]string, 0, len(actionMap))
	for action := range actionMap {
		actions = append(actions, action)
	}

	return actions, nil
}

// GetResourcesByRoleAndAction retrieves all resources where a role has a specific action
func (r *ResourcePermissionRepository) GetResourcesByRoleAndAction(ctx context.Context, roleID, resourceType, action string) ([]string, error) {
	if roleID == "" {
		return nil, fmt.Errorf("role ID is required")
	}

	if resourceType == "" {
		return nil, fmt.Errorf("resource type is required")
	}

	if action == "" {
		return nil, fmt.Errorf("action is required")
	}

	filter := base.NewFilterBuilder().
		Where("role_id", base.OpEqual, roleID).
		Where("resource_type", base.OpEqual, resourceType).
		Where("action", base.OpEqual, action).
		Where("is_active", base.OpEqual, true).
		Build()

	permissions, err := r.BaseFilterableRepository.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	resourceIDs := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		resourceIDs = append(resourceIDs, permission.ResourceID)
	}

	return resourceIDs, nil
}

// EvaluatePermissionContext evaluates permission with additional context
// This can be extended to support conditions like time-based permissions, IP restrictions, etc.
type PermissionContext struct {
	RoleID       string
	ResourceType string
	ResourceID   string
	Action       string
	Timestamp    int64  // For time-based permissions
	IPAddress    string // For IP-based restrictions
	Extra        map[string]interface{}
}

// EvaluateWithContext evaluates permission with additional context
func (r *ResourcePermissionRepository) EvaluateWithContext(ctx context.Context, permCtx *PermissionContext) (bool, string, error) {
	if permCtx == nil {
		return false, "invalid context", fmt.Errorf("permission context is required")
	}

	// Basic permission check
	hasPermission, err := r.HasPermission(ctx, permCtx.RoleID, permCtx.ResourceType, permCtx.ResourceID, permCtx.Action)
	if err != nil {
		return false, "evaluation error", fmt.Errorf("failed to evaluate permission: %w", err)
	}

	if !hasPermission {
		return false, "permission denied", nil
	}

	// TODO: Add additional context-based checks here:
	// - Time-based permissions (valid from/until)
	// - IP-based restrictions
	// - Resource ownership checks
	// - Quota/rate limiting checks

	return true, "permission granted", nil
}
