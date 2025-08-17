package groups

// CreateGroupRequest represents the request for creating a new group
type CreateGroupRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100"`
	Description    string  `json:"description" validate:"max=1000"`
	OrganizationID string  `json:"organization_id" validate:"required,uuid4"`
	ParentID       *string `json:"parent_id" validate:"omitempty,uuid4"`
}
