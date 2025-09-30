package principals

// UpdatePrincipalRequest represents the request for updating an existing principal.
// @Description Request body for updating a principal (all fields optional).
type UpdatePrincipalRequest struct {
	Name           *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated"`
	OrganizationID *string `json:"organization_id,omitempty" validate:"omitempty,org_id" example:"ORGN00000002"`
	IsActive       *bool   `json:"is_active,omitempty" example:"true"`
	Metadata       *string `json:"metadata,omitempty" example:"{\"key\":\"val\"}"`
}
