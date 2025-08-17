package actions

import (
	"fmt"
)

// CreateActionRequest represents a request to create a new action
type CreateActionRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=1000" validate:"max=1000"`
	Category    string  `json:"category" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	IsStatic    bool    `json:"is_static" validate:"omitempty"`
	ServiceID   *string `json:"service_id" binding:"omitempty,max=255" validate:"omitempty,max=255"`
	Metadata    *string `json:"metadata" binding:"omitempty" validate:"omitempty"`
	IsActive    bool    `json:"is_active" validate:"omitempty"`
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
