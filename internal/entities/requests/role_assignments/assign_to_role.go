package role_assignments

import (
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
)

// AssignPermissionsToRoleRequest represents the request to assign permissions to a role
// @Description Request payload for assigning permissions to a role
type AssignPermissionsToRoleRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1,dive,uuid" example:"PERM_abc123,PERM_xyz789"`
}

// Validate validates the AssignPermissionsToRoleRequest
func (r *AssignPermissionsToRoleRequest) Validate() error {
	if len(r.PermissionIDs) == 0 {
		return errors.NewValidationError("at least one permission_id is required")
	}
	return nil
}

// ResourceActionAssignment represents a single resource-action assignment
type ResourceActionAssignment struct {
	ResourceType string   `json:"resource_type" validate:"required,min=1,max=100" example:"aaa/user"`
	ResourceID   string   `json:"resource_id" validate:"required" example:"USR_abc123"`
	Actions      []string `json:"actions" validate:"required,min=1" example:"read,write"`
}

// AssignResourcesToRoleRequest represents the request to assign resources with actions to a role
// @Description Request payload for assigning resource-action combinations to a role
type AssignResourcesToRoleRequest struct {
	Assignments []ResourceActionAssignment `json:"assignments" validate:"required,min=1,dive"`
}

// Validate validates the AssignResourcesToRoleRequest
func (r *AssignResourcesToRoleRequest) Validate() error {
	if len(r.Assignments) == 0 {
		return errors.NewValidationError("at least one assignment is required")
	}

	for i, assignment := range r.Assignments {
		if assignment.ResourceType == "" {
			return errors.NewValidationError("resource_type is required for assignment " + string(rune(i)))
		}
		if assignment.ResourceID == "" {
			return errors.NewValidationError("resource_id is required for assignment " + string(rune(i)))
		}
		if len(assignment.Actions) == 0 {
			return errors.NewValidationError("at least one action is required for assignment " + string(rune(i)))
		}
	}

	return nil
}

// RevokePermissionFromRoleRequest represents the request to revoke a permission from a role
// @Description Request payload for revoking a permission from a role
type RevokePermissionFromRoleRequest struct {
	PermissionID string `json:"permission_id" validate:"required,uuid" example:"PERM_abc123"`
}

// Validate validates the RevokePermissionFromRoleRequest
func (r *RevokePermissionFromRoleRequest) Validate() error {
	if r.PermissionID == "" {
		return errors.NewValidationError("permission_id is required")
	}
	return nil
}

// RevokeResourceFromRoleRequest represents the request to revoke a resource from a role
// @Description Request payload for revoking a resource from a role
type RevokeResourceFromRoleRequest struct {
	ResourceID string `json:"resource_id" validate:"required,uuid" example:"RES_abc123"`
	Action     string `json:"action,omitempty" example:"read"`
}

// Validate validates the RevokeResourceFromRoleRequest
func (r *RevokeResourceFromRoleRequest) Validate() error {
	if r.ResourceID == "" {
		return errors.NewValidationError("resource_id is required")
	}
	return nil
}
