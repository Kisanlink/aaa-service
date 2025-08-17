package principals

// CreateServiceRequest represents the request for creating a new service.
type CreateServiceRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100"`
	Description    string  `json:"description" validate:"max=1000"`
	OrganizationID string  `json:"organization_id" validate:"required,uuid4"`
	APIKey         string  `json:"api_key" validate:"required,min=1,max=255"`
	Metadata       *string `json:"metadata"`
}
