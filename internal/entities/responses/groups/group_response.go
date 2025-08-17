package groups

import "time"

// GroupResponse represents the response for group operations
type GroupResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	OrganizationID string     `json:"organization_id"`
	ParentID       *string    `json:"parent_id"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// GroupMembershipResponse represents the response for group membership operations
type GroupMembershipResponse struct {
	ID            string     `json:"id"`
	GroupID       string     `json:"group_id"`
	PrincipalID   string     `json:"principal_id"`
	PrincipalType string     `json:"principal_type"`
	StartsAt      *time.Time `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at"`
	IsActive      bool       `json:"is_active"`
	AddedByID     string     `json:"added_by_id"`
	CreatedAt     *time.Time `json:"created_at"`
}
