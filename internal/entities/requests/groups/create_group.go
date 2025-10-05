package groups

// CreateGroupRequest represents the request for creating a new group
// @Description Request body for creating a new group
type CreateGroupRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100" example:"DevOps Team"`                       // Group name
	Description    string  `json:"description" validate:"max=1000" example:"DevOps and infrastructure team"`           // Group description
	OrganizationID string  `json:"organization_id" validate:"required,org_id" example:"ORGN00000001"`                  // Organization ID
	ParentID       *string `json:"parent_id,omitempty" validate:"omitempty,group_id" example:"GRP1234567890123456789"` // Optional parent group ID
}
