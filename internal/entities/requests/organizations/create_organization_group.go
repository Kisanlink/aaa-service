package organizations

// CreateOrganizationGroupRequest represents the request for creating a group within an organization
// @Description Request body for creating a group in an organization
type CreateOrganizationGroupRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100" example:"Engineering Team"`                  // Group name
	Description string  `json:"description" validate:"max=1000" example:"Software engineering team members"`        // Group description
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,group_id" example:"GRP1234567890123456789"` // Optional parent group ID
}
