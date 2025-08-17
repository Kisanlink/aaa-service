// Package organizations provides request structures for organization-related operations.
package organizations

// CreateOrganizationRequest represents the request for creating a new organization.
type CreateOrganizationRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=1000"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid4"`
}
