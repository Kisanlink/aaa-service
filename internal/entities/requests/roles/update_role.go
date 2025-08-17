package roles

import (
	"github.com/Kisanlink/aaa-service/internal/entities/requests"
)

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	*requests.BaseRequest
	RoleID      string   `json:"role_id" validate:"required"`
	Name        *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions" validate:"omitempty"`
}

// NewUpdateRoleRequest creates a new UpdateRoleRequest instance
func NewUpdateRoleRequest(
	roleID string,
	name *string,
	description *string,
	permissions []string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *UpdateRoleRequest {
	return &UpdateRoleRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"UpdateRole",
			headers,
			body,
			context,
		),
		RoleID:      roleID,
		Name:        name,
		Description: description,
		Permissions: permissions,
	}
}

// Validate validates the UpdateRoleRequest
func (r *UpdateRoleRequest) Validate() error {
	if r.RoleID == "" {
		return requests.NewValidationError("role_id", "Role ID is required")
	}

	if r.Name != nil {
		if *r.Name == "" {
			return requests.NewValidationError("name", "Role name cannot be empty")
		}

		if len(*r.Name) < 2 {
			return requests.NewValidationError("name", "Role name must be at least 2 characters long")
		}

		if len(*r.Name) > 100 {
			return requests.NewValidationError("name", "Role name must be at most 100 characters long")
		}
	}

	if r.Description != nil && len(*r.Description) > 500 {
		return requests.NewValidationError("description", "Description must be at most 500 characters long")
	}

	return nil
}

// GetRoleID returns the role ID
func (r *UpdateRoleRequest) GetRoleID() string {
	return r.RoleID
}

// GetName returns the role name
func (r *UpdateRoleRequest) GetName() *string {
	return r.Name
}

// GetDescription returns the description
func (r *UpdateRoleRequest) GetDescription() *string {
	return r.Description
}

// GetPermissions returns the permissions
func (r *UpdateRoleRequest) GetPermissions() []string {
	return r.Permissions
}
