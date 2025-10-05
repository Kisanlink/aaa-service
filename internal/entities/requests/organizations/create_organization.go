// Package organizations provides request structures for organization-related operations.
package organizations

// CreateOrganizationRequest represents the request for creating a new organization.
// @Description Request body for creating a new organization
type CreateOrganizationRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100" example:"Acme Corporation"`                  // Organization name
	Description string  `json:"description" validate:"max=1000" example:"Leading provider of innovative solutions"` // Organization description
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,org_id" example:"ORGN00000001"`             // Optional parent organization ID
}
