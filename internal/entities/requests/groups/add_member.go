package groups

import "time"

// AddMemberRequest represents the request for adding a member to a group
type AddMemberRequest struct {
	GroupID       string     `json:"group_id" validate:"required,uuid4"`
	PrincipalID   string     `json:"principal_id" validate:"required,uuid4"`
	PrincipalType string     `json:"principal_type" validate:"required,oneof=user service"`
	AddedByID     string     `json:"added_by_id" validate:"required,uuid4"`
	StartsAt      *time.Time `json:"starts_at"`
	EndsAt        *time.Time `json:"ends_at"`
}
