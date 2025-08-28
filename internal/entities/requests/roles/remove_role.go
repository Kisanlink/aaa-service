package roles

import (
	"github.com/Kisanlink/aaa-service/internal/entities/requests"
)

// RemoveRoleRequest represents a request to remove a role from a user
// UserID and RoleID are expected to come from the URL path parameters
type RemoveRoleRequest struct {
	*requests.BaseRequest
}

// NewRemoveRoleRequest creates a new RemoveRoleRequest instance
func NewRemoveRoleRequest(
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *RemoveRoleRequest {
	return &RemoveRoleRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"RemoveRole",
			headers,
			body,
			context,
		),
	}
}

// Validate validates the RemoveRoleRequest with user ID and role ID from URL path
func (r *RemoveRoleRequest) ValidateWithIDs(userID, roleID string) error {
	if userID == "" {
		return requests.NewValidationError("user_id", "User ID is required")
	}

	if len(userID) < 1 {
		return requests.NewValidationError("user_id", "User ID cannot be empty")
	}

	if roleID == "" {
		return requests.NewValidationError("role_id", "Role ID is required")
	}

	if len(roleID) < 1 {
		return requests.NewValidationError("role_id", "Role ID cannot be empty")
	}

	return nil
}
