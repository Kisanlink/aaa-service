package organizations

import (
	"time"

	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
	userResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/users"
)

// OrganizationGroupResponse represents the response for organization group operations
type OrganizationGroupResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	OrganizationID string     `json:"organization_id"`
	ParentID       *string    `json:"parent_id"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

// OrganizationGroupListResponse represents the response for listing groups within an organization
type OrganizationGroupListResponse struct {
	OrganizationID string                       `json:"organization_id"`
	Groups         []*OrganizationGroupResponse `json:"groups"`
	TotalCount     int64                        `json:"total_count"`
	Page           int                          `json:"page"`
	PageSize       int                          `json:"page_size"`
}

// OrganizationGroupMemberResponse represents the response for group membership within an organization
type OrganizationGroupMemberResponse struct {
	ID            string                      `json:"id"`
	GroupID       string                      `json:"group_id"`
	User          *userResponses.UserResponse `json:"user"`
	PrincipalType string                      `json:"principal_type"`
	StartsAt      *time.Time                  `json:"starts_at"`
	EndsAt        *time.Time                  `json:"ends_at"`
	IsActive      bool                        `json:"is_active"`
	AddedBy       *userResponses.UserResponse `json:"added_by"`
	CreatedAt     *time.Time                  `json:"created_at"`
}

// OrganizationGroupMembersResponse represents the response for listing group members within an organization
type OrganizationGroupMembersResponse struct {
	OrganizationID string                             `json:"organization_id"`
	GroupID        string                             `json:"group_id"`
	Members        []*OrganizationGroupMemberResponse `json:"members"`
	TotalCount     int64                              `json:"total_count"`
	Page           int                                `json:"page"`
	PageSize       int                                `json:"page_size"`
}

// OrganizationGroupRoleResponse represents the response for group role assignments within an organization
type OrganizationGroupRoleResponse struct {
	ID             string                      `json:"id"`
	GroupID        string                      `json:"group_id"`
	RoleID         string                      `json:"role_id"`
	OrganizationID string                      `json:"organization_id"`
	Role           *groupResponses.RoleDetail  `json:"role"`
	AssignedBy     *userResponses.UserResponse `json:"assigned_by"`
	StartsAt       *time.Time                  `json:"starts_at"`
	EndsAt         *time.Time                  `json:"ends_at"`
	IsActive       bool                        `json:"is_active"`
	CreatedAt      *time.Time                  `json:"created_at"`
}

// OrganizationGroupRolesResponse represents the response for listing group roles within an organization
type OrganizationGroupRolesResponse struct {
	OrganizationID string                           `json:"organization_id"`
	GroupID        string                           `json:"group_id"`
	Roles          []*OrganizationGroupRoleResponse `json:"roles"`
	TotalCount     int64                            `json:"total_count"`
	Page           int                              `json:"page"`
	PageSize       int                              `json:"page_size"`
}
