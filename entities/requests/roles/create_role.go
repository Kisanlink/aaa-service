package roles

import (
	"github.com/Kisanlink/aaa-service/entities/requests"
)

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	requests.Request
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions" validate:"omitempty"`
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
		Request: requests.Request{
			Protocol:  protocol,
			Operation: operation,
			Version:   version,
			RequestID: requestID,
			Headers:   headers,
			Body:      body,
			Context:   context,
		},
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
