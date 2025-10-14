package organizations

import (
	"time"

	groupResponses "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/groups"
)

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	ParentID    *string    `json:"parent_id"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// GroupHierarchyNode represents a group with its hierarchy information
type GroupHierarchyNode struct {
	Group    *groupResponses.GroupResponse     `json:"group"`
	Roles    []*groupResponses.GroupRoleDetail `json:"roles"`
	Children []*GroupHierarchyNode             `json:"children"`
}

// OrganizationHierarchyResponse represents the response for organization hierarchy
type OrganizationHierarchyResponse struct {
	Organization *OrganizationResponse   `json:"organization"`
	Parents      []*OrganizationResponse `json:"parents"`
	Children     []*OrganizationResponse `json:"children"`
	Groups       []*GroupHierarchyNode   `json:"groups"`
}

// OrganizationStatsResponse represents statistics about an organization
type OrganizationStatsResponse struct {
	OrganizationID string `json:"organization_id"`
	ChildCount     int64  `json:"child_count"`
	GroupCount     int64  `json:"group_count"`
	UserCount      int64  `json:"user_count"`
}
