package groups

// UpdateGroupRequest represents the request for updating an existing group
// @Description Request body for updating a group (all fields optional)
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Team Name"`          // Group name
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000" example:"Updated team description"` // Group description
	ParentID    *string `json:"parent_id,omitempty" validate:"omitempty,group_id" example:"GRP9876543210987654321"`     // Parent group ID
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`                                                     // Whether group is active
}
