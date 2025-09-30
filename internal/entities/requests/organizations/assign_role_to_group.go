package organizations

import "time"

// AssignRoleToGroupRequest represents the request for assigning a role to a group within an organization
// @Description Request body for assigning a role to a group in an organization
type AssignRoleToGroupRequest struct {
	RoleID   string     `json:"role_id" validate:"required,role_id" example:"ROLE00000001"` // Role ID to assign
	StartsAt *time.Time `json:"starts_at,omitempty" example:"2024-01-01T00:00:00Z"`         // Optional role assignment start time
	EndsAt   *time.Time `json:"ends_at,omitempty" example:"2024-12-31T23:59:59Z"`           // Optional role assignment end time
}
