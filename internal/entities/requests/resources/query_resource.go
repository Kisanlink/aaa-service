package resources

// QueryResourceRequest represents the request to query resources
// @Description Request payload for querying resources with filters
type QueryResourceRequest struct {
	Type     *string `json:"type,omitempty" form:"type" validate:"omitempty,min=1,max=100" example:"aaa/user"`
	ParentID *string `json:"parent_id,omitempty" form:"parent_id" validate:"omitempty,uuid" example:"RES_abc123"`
	OwnerID  *string `json:"owner_id,omitempty" form:"owner_id" validate:"omitempty,uuid" example:"USR_xyz789"`
	IsActive *bool   `json:"is_active,omitempty" form:"is_active" example:"true"`
	Search   *string `json:"search,omitempty" form:"search" validate:"omitempty,max=100" example:"user"`
	Limit    int     `json:"limit" form:"limit" validate:"min=1,max=100" example:"10"`
	Offset   int     `json:"offset" form:"offset" validate:"min=0" example:"0"`
}

// Validate validates the QueryResourceRequest
func (r *QueryResourceRequest) Validate() error {
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
func (r *QueryResourceRequest) GetLimit() int {
	if r.Limit < 1 {
		return 10
	}
	if r.Limit > 100 {
		return 100
	}
	return r.Limit
}

// GetOffset returns the offset with default
func (r *QueryResourceRequest) GetOffset() int {
	if r.Offset < 0 {
		return 0
	}
	return r.Offset
}
