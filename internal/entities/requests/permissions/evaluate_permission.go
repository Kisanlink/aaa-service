package permissions

import (
	"github.com/Kisanlink/aaa-service/pkg/errors"
)

// EvaluatePermissionRequest represents the request to evaluate a permission
// @Description Request payload for evaluating if a user has permission
type EvaluatePermissionRequest struct {
	UserID       string                 `json:"user_id" validate:"required,uuid" example:"USR_abc123"`
	ResourceType string                 `json:"resource_type" validate:"required,min=1,max=100" example:"aaa/user"`
	ResourceID   string                 `json:"resource_id" validate:"required" example:"USR_xyz789"`
	Action       string                 `json:"action" validate:"required,min=1,max=50" example:"read"`
	Context      map[string]interface{} `json:"context,omitempty" example:"{\"organization_id\":\"ORG_123\"}"`
}

// Validate validates the EvaluatePermissionRequest
func (r *EvaluatePermissionRequest) Validate() error {
	if r.UserID == "" {
		return errors.NewValidationError("user_id is required")
	}

	if r.ResourceType == "" {
		return errors.NewValidationError("resource_type is required")
	}

	if r.ResourceID == "" {
		return errors.NewValidationError("resource_id is required")
	}

	if r.Action == "" {
		return errors.NewValidationError("action is required")
	}

	return nil
}

// GetOrganizationID extracts organization_id from context
func (r *EvaluatePermissionRequest) GetOrganizationID() string {
	if r.Context == nil {
		return ""
	}
	if orgID, ok := r.Context["organization_id"].(string); ok {
		return orgID
	}
	return ""
}

// GetGroupID extracts group_id from context
func (r *EvaluatePermissionRequest) GetGroupID() string {
	if r.Context == nil {
		return ""
	}
	if groupID, ok := r.Context["group_id"].(string); ok {
		return groupID
	}
	return ""
}
