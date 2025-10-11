package groups

import (
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests"
)

// AssignRoleToGroupRequest represents a request to assign a role to a group within an organization
type AssignRoleToGroupRequest struct {
	*requests.BaseRequest
	RoleID string `json:"role_id" validate:"required"`
}

// NewAssignRoleToGroupRequest creates a new AssignRoleToGroupRequest instance
func NewAssignRoleToGroupRequest(
	roleID string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *AssignRoleToGroupRequest {
	return &AssignRoleToGroupRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"AssignRoleToGroup",
			headers,
			body,
			context,
		),
		RoleID: roleID,
	}
}

// Validate validates the AssignRoleToGroupRequest
func (r *AssignRoleToGroupRequest) Validate() error {
	if r.RoleID == "" {
		return requests.NewValidationError("role_id", "Role ID is required")
	}

	// Validate role ID format (should be a valid UUID or similar)
	if len(r.RoleID) < 1 {
		return requests.NewValidationError("role_id", "Role ID cannot be empty")
	}

	return nil
}

// ValidateWithGroupID validates the request with a group ID from URL path
func (r *AssignRoleToGroupRequest) ValidateWithGroupID(groupID string) error {
	if groupID == "" {
		return requests.NewValidationError("group_id", "Group ID is required")
	}

	if len(groupID) < 1 {
		return requests.NewValidationError("group_id", "Group ID cannot be empty")
	}

	return r.Validate()
}

// ValidateWithOrgAndGroupID validates the request with organization and group IDs from URL path
func (r *AssignRoleToGroupRequest) ValidateWithOrgAndGroupID(orgID, groupID string) error {
	if orgID == "" {
		return requests.NewValidationError("organization_id", "Organization ID is required")
	}

	if len(orgID) < 1 {
		return requests.NewValidationError("organization_id", "Organization ID cannot be empty")
	}

	return r.ValidateWithGroupID(groupID)
}

// GetRoleID returns the role ID
func (r *AssignRoleToGroupRequest) GetRoleID() string {
	return r.RoleID
}
