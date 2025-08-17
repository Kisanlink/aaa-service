package principals

// UpdatePrincipalRequest represents the request for updating an existing principal.
type UpdatePrincipalRequest struct {
	Name           *string `json:"name" validate:"omitempty,min=1,max=100"`
	OrganizationID *string `json:"organization_id" validate:"omitempty,uuid4"`
	IsActive       *bool   `json:"is_active"`
	Metadata       *string `json:"metadata"`
}
