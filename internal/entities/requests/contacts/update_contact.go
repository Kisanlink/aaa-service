package contacts

import (
	"fmt"
)

// UpdateContactRequest represents a request to update an existing contact
type UpdateContactRequest struct {
	Type        *string `json:"type,omitempty" binding:"omitempty,min=1,max=50" validate:"omitempty,min=1,max=50"`
	Value       *string `json:"value,omitempty" binding:"omitempty,min=1,max=255" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000" validate:"omitempty,max=1000"`
	IsPrimary   *bool   `json:"is_primary,omitempty" validate:"omitempty"`
	IsActive    *bool   `json:"is_active,omitempty" validate:"omitempty"`
	CountryCode *string `json:"country_code,omitempty" binding:"omitempty,max=10" validate:"omitempty,max=10"`
	IsVerified  *bool   `json:"is_verified,omitempty" validate:"omitempty"`
	VerifiedBy  *string `json:"verified_by,omitempty" binding:"omitempty,max=255" validate:"omitempty,max=255"`
}

// Validate validates the UpdateContactRequest
func (r *UpdateContactRequest) Validate() error {
	// Validate type if provided
	if r.Type != nil && len(*r.Type) > 50 {
		return fmt.Errorf("contact type must be at most 50 characters")
	}

	// Validate value if provided
	if r.Value != nil && len(*r.Value) > 255 {
		return fmt.Errorf("contact value must be at most 255 characters")
	}

	// Validate description if provided
	if r.Description != nil && len(*r.Description) > 1000 {
		return fmt.Errorf("contact description must be at most 1000 characters")
	}

	// Validate country code if provided
	if r.CountryCode != nil && len(*r.CountryCode) > 10 {
		return fmt.Errorf("country code must be at most 10 characters")
	}

	// Validate verified by if provided
	if r.VerifiedBy != nil && len(*r.VerifiedBy) > 255 {
		return fmt.Errorf("verified by must be at most 255 characters")
	}

	return nil
}

// GetType returns the type of request
func (r *UpdateContactRequest) GetType() string {
	return "update_contact"
}
