package roles

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
)

// UserRoleResponse represents a user role response
type UserRoleResponse struct {
	responses.Response
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	RoleID    string        `json:"role_id"`
	Role      *RoleResponse `json:"role,omitempty"`
	IsActive  bool          `json:"is_active"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// NewUserRoleResponse creates a new UserRoleResponse from a UserRole model
func NewUserRoleResponse(userRole *models.UserRole) *UserRoleResponse {
	response := &UserRoleResponse{
		ID:        userRole.ID,
		UserID:    userRole.UserID,
		RoleID:    userRole.RoleID,
		IsActive:  userRole.IsActive,
		CreatedAt: userRole.CreatedAt.Format(time.RFC3339),
		UpdatedAt: userRole.UpdatedAt.Format(time.RFC3339),
	}

	// Include role if available
	if userRole.Role.ID != "" {
		response.Role = NewRoleResponse(&userRole.Role)
	}

	return response
}

// NewUserRoleResponseFromModel creates a new UserRoleResponse from a UserRole model
func NewUserRoleResponseFromModel(userRole *models.UserRole) *UserRoleResponse {
	return NewUserRoleResponse(userRole)
}

// GetID returns the user role ID
func (r *UserRoleResponse) GetID() string {
	return r.ID
}

// GetUserID returns the user ID
func (r *UserRoleResponse) GetUserID() string {
	return r.UserID
}

// GetRoleID returns the role ID
func (r *UserRoleResponse) GetRoleID() string {
	return r.RoleID
}

// GetRole returns the role
func (r *UserRoleResponse) GetRole() *RoleResponse {
	return r.Role
}

// GetCreatedAt returns the created at timestamp
func (r *UserRoleResponse) GetCreatedAt() string {
	return r.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (r *UserRoleResponse) GetUpdatedAt() string {
	return r.UpdatedAt
}
