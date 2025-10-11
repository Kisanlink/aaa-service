package roles

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
)

// RoleResponse represents a role response
type RoleResponse struct {
	responses.Response
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsActive    bool     `json:"is_active"`
	Permissions []string `json:"permissions,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// NewRoleResponse creates a new RoleResponse from a Role model
func NewRoleResponse(role *models.Role) *RoleResponse {
	// Convert permissions to string array
	var permissionNames []string
	for _, permission := range role.Permissions {
		permissionNames = append(permissionNames, permission.Name)
	}

	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsActive:    role.IsActive,
		Permissions: permissionNames,
		CreatedAt:   role.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   role.UpdatedAt.Format(time.RFC3339),
	}
}

// NewRoleResponseFromModel creates a new RoleResponse from a Role model
func NewRoleResponseFromModel(role *models.Role) *RoleResponse {
	return NewRoleResponse(role)
}

// GetID returns the role ID
func (r *RoleResponse) GetID() string {
	return r.ID
}

// GetName returns the role name
func (r *RoleResponse) GetName() string {
	return r.Name
}

// GetDescription returns the description
func (r *RoleResponse) GetDescription() string {
	return r.Description
}

// GetPermissions returns the permissions
func (r *RoleResponse) GetPermissions() []string {
	return r.Permissions
}

// GetCreatedAt returns the created at timestamp
func (r *RoleResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *RoleResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
