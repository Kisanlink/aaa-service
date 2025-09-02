package organizations

// UpdateOrganizationGroupRequest represents the request for updating a group within an organization
type UpdateOrganizationGroupRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid4"`
	IsActive    *bool   `json:"is_active"`
}
