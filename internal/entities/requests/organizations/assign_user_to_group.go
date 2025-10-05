package organizations

import "time"

// AssignUserToGroupRequest represents the request for assigning a user to a group within an organization
// @Description Request body for adding a user to a group in an organization
type AssignUserToGroupRequest struct {
	PrincipalID   string     `json:"principal_id" validate:"required,user_id" example:"USER00000001"`      // Principal ID (user or service)
	PrincipalType string     `json:"principal_type" validate:"required,oneof=user service" example:"user"` // Principal type: user or service
	StartsAt      *time.Time `json:"starts_at,omitempty" example:"2024-01-01T00:00:00Z"`                   // Optional membership start time
	EndsAt        *time.Time `json:"ends_at,omitempty" example:"2024-12-31T23:59:59Z"`                     // Optional membership end time
}
