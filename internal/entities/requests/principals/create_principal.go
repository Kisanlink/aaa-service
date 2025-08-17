// Package principals provides request structures for principal-related operations.
package principals

// CreatePrincipalRequest represents the request for creating a new principal.
type CreatePrincipalRequest struct {
	Type           string  `json:"type" validate:"required,oneof=user service"`
	UserID         *string `json:"user_id" validate:"omitempty,uuid4"`
	ServiceID      *string `json:"service_id" validate:"omitempty,uuid4"`
	Name           string  `json:"name" validate:"required,min=1,max=100"`
	OrganizationID *string `json:"organization_id" validate:"omitempty,uuid4"`
	Metadata       *string `json:"metadata"`
}
