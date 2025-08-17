package actions

import (
	"fmt"
)

// UpdateActionRequest represents a request to update an existing action
type UpdateActionRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" validate:"omitempty,max=1000"`
	Category    *string `json:"category,omitempty" binding:"omitempty,min=1,max=50" validate:"omitempty,min=1,max=50"`
	IsStatic    *bool   `json:"is_static,omitempty" validate:"omitempty"`
	ServiceID   *string `json:"service_id,omitempty" binding:"omitempty,max=255" validate:"omitempty,max=255"`
	Metadata    *string `json:"metadata,omitempty" binding:"omitempty" validate:"omitempty"`
	IsActive    *bool   `json:"is_active,omitempty" validate:"omitempty"`
}

// Validate validates the UpdateActionRequest
func (r *UpdateActionRequest) Validate() error {
	// Validate name if provided
	if r.Name != nil {
		if *r.Name == "" {
			return fmt.Errorf("action name cannot be empty")
		}
		if len(*r.Name) > 100 {
			return fmt.Errorf("action name must be at most 100 characters")
		}
	}

	// Validate category if provided
	if r.Category != nil {
		if *r.Category == "" {
			return fmt.Errorf("action category cannot be empty")
		}
		if len(*r.Category) > 50 {
			return fmt.Errorf("action category must be at most 50 characters")
		}
	}

	// Validate description if provided
	if r.Description != nil && len(*r.Description) > 1000 {
		return fmt.Errorf("action description must be at most 1000 characters")
	}

	// Validate service ID if provided
	if r.ServiceID != nil && len(*r.ServiceID) > 255 {
		return fmt.Errorf("service ID must be at most 255 characters")
	}

	return nil
}

// GetType returns the type of request
func (r *UpdateActionRequest) GetType() string {
	return "update_action"
}
