package permissions

import (
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// AssignPermissionRequest represents the request to assign permissions to a role
// @Description Request payload for assigning permissions to a role
type AssignPermissionRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1,dive,uuid" example:"PERM_abc123,PERM_xyz789"`
}

// Validate validates the AssignPermissionRequest
func (r *AssignPermissionRequest) Validate() error {
	if len(r.PermissionIDs) == 0 {
		return errors.NewValidationError("at least one permission_id is required")
	}
	return nil
}

// RevokePermissionRequest represents the request to revoke a permission from a role
// @Description Request payload for revoking a permission from a role
type RevokePermissionRequest struct {
	PermissionID string `json:"permission_id" validate:"required,uuid" example:"PERM_abc123"`
}

// Validate validates the RevokePermissionRequest
func (r *RevokePermissionRequest) Validate() error {
	if r.PermissionID == "" {
		return errors.NewValidationError("permission_id is required")
	}
	return nil
}
