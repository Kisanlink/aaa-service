package groups

// UpdateGroupRequest represents the request for updating an existing group
type UpdateGroupRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid4"`
	IsActive    *bool   `json:"is_active"`
}
