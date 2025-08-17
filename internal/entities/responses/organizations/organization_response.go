package organizations

import "time"

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ParentID    *string    `json:"parent_id"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// OrganizationHierarchyResponse represents the response for organization hierarchy
type OrganizationHierarchyResponse struct {
	Organization *OrganizationResponse   `json:"organization"`
	Parents      []*OrganizationResponse `json:"parents"`
	Children     []*OrganizationResponse `json:"children"`
}

// OrganizationStatsResponse represents statistics about an organization
type OrganizationStatsResponse struct {
	OrganizationID string `json:"organization_id"`
	ChildCount     int64  `json:"child_count"`
	GroupCount     int64  `json:"group_count"`
	UserCount      int64  `json:"user_count"`
}
