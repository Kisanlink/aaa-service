package actions

import (
	"fmt"
)

// CreateActionRequest represents a request to create a new action
// @Description Define a new action that can be performed on resources
type CreateActionRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100" example:"read"`
	Description string  `json:"description" binding:"max=1000" validate:"max=1000" example:"Read or view data without making changes"`
	Category    string  `json:"category" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"data_access"`
	IsStatic    bool    `json:"is_static" validate:"omitempty" example:"true"`
	ServiceID   *string `json:"service_id" binding:"omitempty,max=255" validate:"omitempty,max=255" example:"aaa-service"`
	Metadata    *string `json:"metadata" binding:"omitempty" validate:"omitempty" example:"{\"http_method\": \"GET\", \"rest_ful\": true}"`
	IsActive    bool    `json:"is_active" validate:"omitempty" example:"true"`
}

// Validate validates the CreateActionRequest
func (r *CreateActionRequest) Validate() error {
	// Validate name
	if r.Name == "" {
		return fmt.Errorf("action name is required")
	}
	if len(r.Name) > 100 {
		return fmt.Errorf("action name must be at most 100 characters")
	}

	// Validate category
	if r.Category == "" {
		return fmt.Errorf("action category is required")
	}
	if len(r.Category) > 50 {
		return fmt.Errorf("action category must be at most 50 characters")
	}

	// Validate description
	if len(r.Description) > 1000 {
		return fmt.Errorf("action description must be at most 1000 characters")
	}

	// Validate service ID if provided
	if r.ServiceID != nil && len(*r.ServiceID) > 255 {
		return fmt.Errorf("service ID must be at most 255 characters")
	}

	return nil
}

// GetType returns the type of request
func (r *CreateActionRequest) GetType() string {
	return "create_action"
}
