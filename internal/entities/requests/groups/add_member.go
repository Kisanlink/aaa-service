package groups

import "time"

// AddMemberRequest represents the request for adding a member to a group
// @Description Request body for adding a member to a group
type AddMemberRequest struct {
	GroupID       string     `json:"group_id" validate:"required,group_id" example:"GRP1234567890123456789"` // Group ID
	PrincipalID   string     `json:"principal_id" validate:"required,user_id" example:"USER00000001"`        // Principal ID (user or service)
	PrincipalType string     `json:"principal_type" validate:"required,oneof=user service" example:"user"`   // Principal type: user or service
	AddedByID     string     `json:"added_by_id" validate:"required,user_id" example:"USER00000002"`         // ID of user adding the member
	StartsAt      *time.Time `json:"starts_at,omitempty" example:"2024-01-01T00:00:00Z"`                     // Optional membership start time
	EndsAt        *time.Time `json:"ends_at,omitempty" example:"2024-12-31T23:59:59Z"`                       // Optional membership end time
}
