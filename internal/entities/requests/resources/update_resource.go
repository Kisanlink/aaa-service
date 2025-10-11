package resources

import (
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
)

// UpdateResourceRequest represents the request to update an existing resource
// @Description Request payload for updating a resource
type UpdateResourceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100" example:"Updated User Management"`
	Type        *string `json:"type,omitempty" validate:"omitempty,min=3,max=100" example:"aaa/user"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated description"`
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,uuid" example:"RES_abc123"`
	OwnerID     *string `json:"owner_id,omitempty" validate:"omitempty,uuid" example:"USR_xyz789"`
}

// Validate validates the UpdateResourceRequest
func (r *UpdateResourceRequest) Validate() error {
	if r.Name != nil {
		if len(*r.Name) < 3 {
			return errors.NewValidationError("name must be at least 3 characters")
		}
		if len(*r.Name) > 100 {
			return errors.NewValidationError("name must be at most 100 characters")
		}
	}

	if r.Type != nil {
		if len(*r.Type) < 3 {
			return errors.NewValidationError("type must be at least 3 characters")
		}
		if len(*r.Type) > 100 {
			return errors.NewValidationError("type must be at most 100 characters")
		}
	}

	if r.Description != nil && len(*r.Description) > 500 {
		return errors.NewValidationError("description must be at most 500 characters")
	}

	return nil
}

// HasUpdates checks if the request has any fields to update
func (r *UpdateResourceRequest) HasUpdates() bool {
	return r.Name != nil || r.Type != nil || r.Description != nil ||
		r.IsActive != nil || r.ParentID != nil || r.OwnerID != nil
}
