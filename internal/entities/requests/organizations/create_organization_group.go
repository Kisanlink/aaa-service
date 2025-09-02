package organizations

// CreateOrganizationGroupRequest represents the request for creating a group within an organization
type CreateOrganizationGroupRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=1000"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid4"`
}
