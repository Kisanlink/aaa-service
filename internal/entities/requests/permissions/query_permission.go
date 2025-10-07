package permissions

// QueryPermissionRequest represents the request to query permissions
// @Description Request payload for querying permissions with filters
type QueryPermissionRequest struct {
	RoleID     *string `json:"role_id,omitempty" form:"role_id" validate:"omitempty,uuid" example:"ROLE_abc123"`
	ResourceID *string `json:"resource_id,omitempty" form:"resource_id" validate:"omitempty,uuid" example:"RES_abc123"`
	ActionID   *string `json:"action_id,omitempty" form:"action_id" validate:"omitempty,uuid" example:"ACT_xyz789"`
	IsActive   *bool   `json:"is_active,omitempty" form:"is_active" example:"true"`
	Search     *string `json:"search,omitempty" form:"search" validate:"omitempty,max=100" example:"manage"`
	Limit      int     `json:"limit" form:"limit" validate:"min=1,max=100" example:"10"`
	Offset     int     `json:"offset" form:"offset" validate:"min=0" example:"0"`
}

// Validate validates the QueryPermissionRequest
func (r *QueryPermissionRequest) Validate() error {
	if r.Limit < 1 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

// GetLimit returns the limit with default
func (r *QueryPermissionRequest) GetLimit() int {
	if r.Limit < 1 {
		return 10
	}
	if r.Limit > 100 {
		return 100
	}
	return r.Limit
}

// GetOffset returns the offset with default
func (r *QueryPermissionRequest) GetOffset() int {
	if r.Offset < 0 {
		return 0
	}
	return r.Offset
}
