package roles

import (
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	requests.Request
	UserID string `json:"user_id" validate:"required"`
	RoleID string `json:"role_id" validate:"required"`
}

// NewAssignRoleRequest creates a new AssignRoleRequest instance
func NewAssignRoleRequest(
	userID string,
	roleID string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *AssignRoleRequest {
	return &AssignRoleRequest{
		Request: requests.Request{
			Protocol:  protocol,
			Operation: operation,
			Version:   version,
			RequestID: requestID,
			Headers:   headers,
			Body:      body,
			Context:   context,
		},
		UserID: userID,
		RoleID: roleID,
	}
}

// Validate validates the AssignRoleRequest
func (r *AssignRoleRequest) Validate() error {
	if r.UserID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if r.RoleID == "" {
		return requests.NewValidationError("role_id", "Role ID is required")
	}

	return nil
}

// GetUserID returns the user ID
func (r *AssignRoleRequest) GetUserID() string {
	return r.UserID
}

// GetRoleID returns the role ID
func (r *AssignRoleRequest) GetRoleID() string {
	return r.RoleID
}
