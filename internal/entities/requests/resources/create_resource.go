package resources

import (
	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// CreateResourceRequest represents the request to create a new resource
// @Description Request payload for creating a new resource
type CreateResourceRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=100" example:"User Management"`
	Type        string  `json:"type" validate:"required,min=3,max=100" example:"aaa/user"`
	Description string  `json:"description" validate:"max=500" example:"Resource for managing users"`
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,uuid" example:"RES_abc123"`
	OwnerID     *string `json:"owner_id,omitempty" validate:"omitempty,uuid" example:"USR_xyz789"`
}

// Validate validates the CreateResourceRequest
func (r *CreateResourceRequest) Validate() error {
	if r.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if len(r.Name) < 3 {
		return errors.NewValidationError("name must be at least 3 characters")
	}
	if len(r.Name) > 100 {
		return errors.NewValidationError("name must be at most 100 characters")
	}

	if r.Type == "" {
		return errors.NewValidationError("type is required")
	}
	if len(r.Type) < 3 {
		return errors.NewValidationError("type must be at least 3 characters")
	}
	if len(r.Type) > 100 {
		return errors.NewValidationError("type must be at most 100 characters")
	}

	if len(r.Description) > 500 {
		return errors.NewValidationError("description must be at most 500 characters")
	}

	return nil
}

// ToModel converts the request to a Resource model
func (r *CreateResourceRequest) ToModel() *models.Resource {
	resource := models.NewResource(r.Name, r.Type, r.Description)
	if r.ParentID != nil {
		resource.ParentID = r.ParentID
	}
	if r.OwnerID != nil {
		resource.OwnerID = r.OwnerID
	}
	return resource
}
