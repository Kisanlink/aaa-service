package roles

import (
	"github.com/Kisanlink/aaa-service/v2/internal/entities/requests"
)

// CreateRoleRequest represents a request to create a role
// @Description Create a new role with name, description, and optional permissions
type CreateRoleRequest struct {
	*requests.BaseRequest
	Name        string   `json:"name" validate:"required,min=2,max=100" example:"farm_manager"`
	Description *string  `json:"description" validate:"omitempty,max=500" example:"Manager role for farm operations and crop management"`
	Permissions []string `json:"permissions" validate:"omitempty" example:"PERM00000001,PERM00000002"`
}

// NewCreateRoleRequest creates a new CreateRoleRequest instance
func NewCreateRoleRequest(
	name string,
	description *string,
	permissions []string,
	protocol string,
	operation string,
	version string,
	requestID string,
	headers map[string][]string,
	body interface{},
	context map[string]interface{},
) *CreateRoleRequest {
	return &CreateRoleRequest{
		BaseRequest: requests.NewBaseRequest(
			protocol,
			operation,
			version,
			requestID,
			"CreateRole",
			headers,
			body,
			context,
		),
		Name:        name,
		Description: description,
		Permissions: permissions,
	}
}

// Validate validates the CreateRoleRequest
func (r *CreateRoleRequest) Validate() error {
	if r.Name == "" {
		return requests.NewValidationError("name", "Role name is required")
	}

	if len(r.Name) < 2 {
		return requests.NewValidationError("name", "Role name must be at least 2 characters long")
	}

	if len(r.Name) > 100 {
		return requests.NewValidationError("name", "Role name must be at most 100 characters long")
	}

	if r.Description != nil && len(*r.Description) > 500 {
		return requests.NewValidationError("description", "Description must be at most 500 characters long")
	}

	return nil
}

// GetName returns the role name
func (r *CreateRoleRequest) GetName() string {
	return r.Name
}

// GetDescription returns the description
func (r *CreateRoleRequest) GetDescription() *string {
	return r.Description
}

// GetPermissions returns the permissions
func (r *CreateRoleRequest) GetPermissions() []string {
	return r.Permissions
}
