package principals

// CreateServiceRequest represents the request for creating a new service.
// @Description Request body for creating a new service principal.
type CreateServiceRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100" example:"Payment API"`
	Description    string  `json:"description" validate:"max=1000" example:"Service for payments"`
	OrganizationID string  `json:"organization_id" validate:"required,org_id" example:"ORGN00000001"`
	APIKey         string  `json:"api_key" validate:"required,min=1,max=255" example:"sk_live_abc"`
	Metadata       *string `json:"metadata,omitempty" example:"{\"version\":\"v1\"}"`
}
