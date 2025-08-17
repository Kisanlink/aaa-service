package principals

import "time"

// PrincipalResponse represents the response for principal operations
type PrincipalResponse struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	UserID         *string    `json:"user_id"`
	ServiceID      *string    `json:"service_id"`
	Name           string     `json:"name"`
	OrganizationID *string    `json:"organization_id"`
	IsActive       bool       `json:"is_active"`
	Metadata       *string    `json:"metadata"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// ServiceResponse represents the response for service operations
type ServiceResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	OrganizationID string     `json:"organization_id"`
	IsActive       bool       `json:"is_active"`
	Metadata       *string    `json:"metadata"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}
