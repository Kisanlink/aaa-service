package organizations

import "time"

// AssignUserToGroupRequest represents the request for assigning a user to a group within an organization
type AssignUserToGroupRequest struct {
	UserID        string     `json:"user_id" validate:"required,uuid4"`
	PrincipalType string     `json:"principal_type" validate:"required,oneof=user service"`
	StartsAt      *time.Time `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at"`
}
