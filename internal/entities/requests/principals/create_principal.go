// Package principals provides request structures for principal-related operations.
package principals

// CreatePrincipalRequest represents the request for creating a new principal.
// @Description Request body for creating a new principal (user or service).
type CreatePrincipalRequest struct {
	Type           string  `json:"type" validate:"required,oneof=user service" example:"user"`
	UserID         *string `json:"user_id,omitempty" validate:"omitempty,user_id" example:"USER00000001"`
	ServiceID      *string `json:"service_id,omitempty" example:"service-01"`
	Name           string  `json:"name" validate:"required,min=1,max=100" example:"John Doe"`
	OrganizationID *string `json:"organization_id,omitempty" validate:"omitempty,org_id" example:"ORGN00000001"`
	Metadata       *string `json:"metadata,omitempty" example:"{\"dept\":\"eng\"}"`
}
