package permissions

import (
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
)

// CreatePermissionRequest represents the request to create a new permission
// @Description Request payload for creating a new permission with resource and action
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100" example:"crop_management_create"`
	Description string `json:"description" validate:"max=500" example:"Permission to create and add new crops to the farm inventory"`
	ResourceID  string `json:"resource_id" validate:"required,uuid" example:"RES1760615540005820900"`
	ActionID    string `json:"action_id" validate:"required,uuid" example:"ACT1760615540005820901"`
}

// Validate validates the CreatePermissionRequest
func (r *CreatePermissionRequest) Validate() error {
	if r.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if len(r.Name) < 3 {
		return errors.NewValidationError("name must be at least 3 characters")
	}
	if len(r.Name) > 100 {
		return errors.NewValidationError("name must be at most 100 characters")
	}

	if r.ResourceID == "" {
		return errors.NewValidationError("resource_id is required")
	}

	if r.ActionID == "" {
		return errors.NewValidationError("action_id is required")
	}

	if len(r.Description) > 500 {
		return errors.NewValidationError("description must be at most 500 characters")
	}

	return nil
}

// ToModel converts the request to a Permission model
func (r *CreatePermissionRequest) ToModel() *models.Permission {
	permission := models.NewPermission(r.Name, r.Description)
	permission.ResourceID = &r.ResourceID
	permission.ActionID = &r.ActionID
	return permission
}
