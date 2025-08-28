package roles

import (
	"github.com/Kisanlink/aaa-service/internal/entities/requests"
)

// AssignRoleRequest represents a request to assign a role to a user
// UserID is expected to come from the URL path parameter
type AssignRoleRequest struct {
	*requests.BaseRequest
	RoleID string `json:"role_id" validate:"required"`
}

// NewAssignRoleRequest creates a new AssignRoleRequest instance
func NewAssignRoleRequest(
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
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"AssignRole",
			headers,
			body,
			context,
		),
		RoleID: roleID,
	}
}

// Validate validates the AssignRoleRequest
func (r *AssignRoleRequest) Validate() error {
	if r.RoleID == "" {
		return requests.NewValidationError("role_id", "Role ID is required")
	}

	// Validate role ID format (should be a valid UUID or similar)
	if len(r.RoleID) < 1 {
		return requests.NewValidationError("role_id", "Role ID cannot be empty")
	}

	return nil
}

// ValidateWithUserID validates the request with a user ID from URL path
func (r *AssignRoleRequest) ValidateWithUserID(userID string) error {
	if userID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if len(userID) < 1 {
		return requests.NewValidationError("user_id", "User ID cannot be empty")
	}

	return r.Validate()
}

// GetRoleID returns the role ID
func (r *AssignRoleRequest) GetRoleID() string {
	return r.RoleID
}
