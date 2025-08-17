package contacts

import (
	"fmt"
)

// CreateContactRequest represents a request to create a new contact
type CreateContactRequest struct {
	UserID      string  `json:"user_id" binding:"required" validate:"required"`
	Type        string  `json:"type" binding:"required,min=1,max=50" validate:"required,min=1,max=50"`
	Value       string  `json:"value" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=1000" validate:"omitempty,max=1000"`
	IsPrimary   bool    `json:"is_primary" validate:"omitempty"`
	IsActive    bool    `json:"is_active" validate:"omitempty"`
	CountryCode *string `json:"country_code" binding:"omitempty,max=10" validate:"omitempty,max=10"`
}

// Validate validates the CreateContactRequest
func (r *CreateContactRequest) Validate() error {
	// Validate user ID
	if r.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate type
	if r.Type == "" {
		return fmt.Errorf("contact type is required")
	}
	if len(r.Type) > 50 {
		return fmt.Errorf("contact type must be at most 50 characters")
	}

	// Validate value
	if r.Value == "" {
		return fmt.Errorf("contact value is required")
	}
	if len(r.Value) > 255 {
		return fmt.Errorf("contact value must be at most 255 characters")
	}

	// Validate description if provided
	if r.Description != nil && len(*r.Description) > 500 {
		return fmt.Errorf("contact description must be at most 500 characters")
	}

	return nil
}

// GetType returns the type of request
func (r *CreateContactRequest) GetType() string {
	return "create_contact"
}
