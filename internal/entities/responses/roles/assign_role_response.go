package roles

import (
	"fmt"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/entities/responses"
)

// AssignRoleResponse represents the response for role assignment operations
type AssignRoleResponse struct {
	responses.Response
	Message string     `json:"message"`
	UserID  string     `json:"user_id"`
	Role    RoleDetail `json:"role"`
}

// RoleDetail represents detailed role information in responses
type RoleDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// UserRoleDetail represents detailed user role relationship information
type UserRoleDetail struct {
	ID       string     `json:"id"`
	UserID   string     `json:"user_id"`
	RoleID   string     `json:"role_id"`
	Role     RoleDetail `json:"role"`
	IsActive bool       `json:"is_active"`
}

// NewAssignRoleResponse creates a new AssignRoleResponse
func NewAssignRoleResponse(userID string, role *models.Role, message string) *AssignRoleResponse {
	return &AssignRoleResponse{
		Message: message,
		UserID:  userID,
		Role: RoleDetail{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			IsActive:    role.IsActive,
		},
	}
}

// NewRoleDetail creates a RoleDetail from a Role model
func NewRoleDetail(role *models.Role) RoleDetail {
	return RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsActive:    role.IsActive,
	}
}

// NewUserRoleDetail creates a UserRoleDetail from a UserRole model
func NewUserRoleDetail(userRole *models.UserRole) UserRoleDetail {
	return UserRoleDetail{
		ID:       userRole.ID,
		UserID:   userRole.UserID,
		RoleID:   userRole.RoleID,
		Role:     NewRoleDetail(&userRole.Role),
		IsActive: userRole.IsActive,
	}
}

// GetType returns the response type
func (r *AssignRoleResponse) GetType() string {
	return "AssignRoleResponse"
}

// IsSuccess returns whether the response indicates success
func (r *AssignRoleResponse) IsSuccess() bool {
	return true
}

// GetProtocol returns the transport protocol
func (r *AssignRoleResponse) GetProtocol() string {
	return "http"
}

// GetOperation returns the operation
func (r *AssignRoleResponse) GetOperation() string {
	return "post"
}

// GetVersion returns the API version
func (r *AssignRoleResponse) GetVersion() string {
	return "v2"
}

// GetResponseID returns the response ID
func (r *AssignRoleResponse) GetResponseID() string {
	return ""
}

// GetHeaders returns response headers
func (r *AssignRoleResponse) GetHeaders() map[string][]string {
	return nil
}

// GetBody returns the response body
func (r *AssignRoleResponse) GetBody() interface{} {
	return r
}

// GetContext returns response context
func (r *AssignRoleResponse) GetContext() map[string]interface{} {
	return nil
}

// ToProto converts to protocol buffer format
func (r *AssignRoleResponse) ToProto() interface{} {
	return nil
}

// String returns a string representation
func (r *AssignRoleResponse) String() string {
	return fmt.Sprintf("AssignRoleResponse{UserID: %s, Role: %s, Message: %s}",
		r.UserID, r.Role.Name, r.Message)
}
