package permissions

import (
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// UpdatePermissionRequest represents the request to update an existing permission
// @Description Request payload for updating a permission
type UpdatePermissionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100" example:"updated_permission"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated description"`
	ResourceID  *string `json:"resource_id,omitempty" validate:"omitempty,uuid" example:"RES_abc123"`
	ActionID    *string `json:"action_id,omitempty" validate:"omitempty,uuid" example:"ACT_xyz789"`
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`
}

// Validate validates the UpdatePermissionRequest
func (r *UpdatePermissionRequest) Validate() error {
	if r.Name != nil {
		if len(*r.Name) < 3 {
			return errors.NewValidationError("name must be at least 3 characters")
		}
		if len(*r.Name) > 100 {
			return errors.NewValidationError("name must be at most 100 characters")
		}
	}

	if r.Description != nil && len(*r.Description) > 500 {
		return errors.NewValidationError("description must be at most 500 characters")
	}

	return nil
}

// HasUpdates checks if the request has any fields to update
func (r *UpdatePermissionRequest) HasUpdates() bool {
	return r.Name != nil || r.Description != nil ||
		r.ResourceID != nil || r.ActionID != nil || r.IsActive != nil
}
