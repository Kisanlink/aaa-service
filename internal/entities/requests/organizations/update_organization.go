// Package organizations provides request structures for organization-related operations.
package organizations

// UpdateOrganizationRequest represents the request for updating an existing organization.
type UpdateOrganizationRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid4"`
	IsActive    *bool   `json:"is_active"`
}
