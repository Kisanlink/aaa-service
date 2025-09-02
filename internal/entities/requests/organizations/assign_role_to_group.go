package organizations

import "time"

// AssignRoleToGroupRequest represents the request for assigning a role to a group within an organization
type AssignRoleToGroupRequest struct {
	RoleID   string     `json:"role_id" validate:"required,uuid4"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}
