package organizations

import (
	groupResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/groups"
	roleResponses "github.com/Kisanlink/aaa-service/internal/entities/responses/roles"
)

// OrganizationHierarchyGroupNode represents a group node in the organization hierarchy
type OrganizationHierarchyGroupNode struct {
	Group    *OrganizationGroupResponse         `json:"group"`
	Roles    []*groupResponses.GroupRoleDetail  `json:"roles"`
	Members  []*OrganizationGroupMemberResponse `json:"members"`
	Children []*OrganizationHierarchyGroupNode  `json:"children"`
}

// OrganizationCompleteHierarchyResponse represents the complete hierarchy response for an organization
type OrganizationCompleteHierarchyResponse struct {
	Organization *OrganizationResponse             `json:"organization"`
	Parents      []*OrganizationResponse           `json:"parents"`
	Children     []*OrganizationResponse           `json:"children"`
	Groups       []*OrganizationHierarchyGroupNode `json:"groups"`
	Stats        *OrganizationStatsResponse        `json:"stats"`
}

// UserGroupMembershipResponse represents a user's group membership within an organization
type UserGroupMembershipResponse struct {
	GroupID       string `json:"group_id"`
	GroupName     string `json:"group_name"`
	GroupPath     string `json:"group_path"`
	PrincipalType string `json:"principal_type"`
	IsActive      bool   `json:"is_active"`
	IsDirect      bool   `json:"is_direct"`
	Source        string `json:"source"` // "direct", "inherited_up", "inherited_down"
}

// UserOrganizationGroupsResponse represents the response for a user's groups within an organization
type UserOrganizationGroupsResponse struct {
	OrganizationID string                         `json:"organization_id"`
	UserID         string                         `json:"user_id"`
	Groups         []*UserGroupMembershipResponse `json:"groups"`
	TotalCount     int64                          `json:"total_count"`
}

// EffectiveRoleResponse represents an effective role for a user within an organization
type EffectiveRoleResponse struct {
	Role        *roleResponses.RoleResponse `json:"role"`
	Source      string                      `json:"source"` // "direct", "group_direct", "group_inherited"
	SourceGroup *OrganizationGroupResponse  `json:"source_group,omitempty"`
	IsActive    bool                        `json:"is_active"`
}

// UserEffectiveRolesResponse represents the response for a user's effective roles within an organization
type UserEffectiveRolesResponse struct {
	OrganizationID string                   `json:"organization_id"`
	UserID         string                   `json:"user_id"`
	EffectiveRoles []*EffectiveRoleResponse `json:"effective_roles"`
	TotalCount     int64                    `json:"total_count"`
}
