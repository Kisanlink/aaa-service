package groups

import (
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/responses"
)

// GroupRoleResponse represents the response for group role assignment operations
type GroupRoleResponse struct {
	responses.Response
	Message        string     `json:"message"`
	GroupID        string     `json:"group_id"`
	OrganizationID string     `json:"organization_id"`
	Role           RoleDetail `json:"role"`
	AssignedBy     string     `json:"assigned_by"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	EndsAt         *time.Time `json:"ends_at,omitempty"`
	IsActive       bool       `json:"is_active"`
}

// RoleDetail represents detailed role information in responses
type RoleDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// GroupRoleDetail represents detailed group role relationship information
type GroupRoleDetail struct {
	ID             string     `json:"id"`
	GroupID        string     `json:"group_id"`
	RoleID         string     `json:"role_id"`
	OrganizationID string     `json:"organization_id"`
	Role           RoleDetail `json:"role"`
	AssignedBy     string     `json:"assigned_by"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	EndsAt         *time.Time `json:"ends_at,omitempty"`
	IsActive       bool       `json:"is_active"`
}

// RemoveGroupRoleResponse represents the response for group role removal operations
type RemoveGroupRoleResponse struct {
	Message        string `json:"message"`
	GroupID        string `json:"group_id"`
	RoleID         string `json:"role_id"`
	OrganizationID string `json:"organization_id"`
}

// NewGroupRoleResponse creates a new GroupRoleResponse
func NewGroupRoleResponse(groupRole *models.GroupRole, message string) *GroupRoleResponse {
	response := &GroupRoleResponse{
		Message:        message,
		GroupID:        groupRole.GroupID,
		OrganizationID: groupRole.OrganizationID,
		AssignedBy:     groupRole.AssignedBy,
		StartsAt:       groupRole.StartsAt,
		EndsAt:         groupRole.EndsAt,
		IsActive:       groupRole.IsActive,
	}

	// Set role details if available
	if groupRole.Role != nil {
		response.Role = RoleDetail{
			ID:          groupRole.Role.ID,
			Name:        groupRole.Role.Name,
			Description: groupRole.Role.Description,
			IsActive:    groupRole.Role.IsActive,
		}
	}

	return response
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

// NewGroupRoleDetail creates a GroupRoleDetail from a GroupRole model
func NewGroupRoleDetail(groupRole *models.GroupRole) GroupRoleDetail {
	detail := GroupRoleDetail{
		ID:             groupRole.ID,
		GroupID:        groupRole.GroupID,
		RoleID:         groupRole.RoleID,
		OrganizationID: groupRole.OrganizationID,
		AssignedBy:     groupRole.AssignedBy,
		StartsAt:       groupRole.StartsAt,
		EndsAt:         groupRole.EndsAt,
		IsActive:       groupRole.IsActive,
	}

	// Set role details if available
	if groupRole.Role != nil {
		detail.Role = NewRoleDetail(groupRole.Role)
	}

	return detail
}

// NewRemoveGroupRoleResponse creates a new RemoveGroupRoleResponse
func NewRemoveGroupRoleResponse(groupID, roleID, organizationID, message string) *RemoveGroupRoleResponse {
	return &RemoveGroupRoleResponse{
		Message:        message,
		GroupID:        groupID,
		RoleID:         roleID,
		OrganizationID: organizationID,
	}
}

// GetType returns the response type
func (r *GroupRoleResponse) GetType() string {
	return "GroupRoleResponse"
}

// IsSuccess returns whether the response indicates success
func (r *GroupRoleResponse) IsSuccess() bool {
	return true
}

// GetProtocol returns the transport protocol
func (r *GroupRoleResponse) GetProtocol() string {
	return "http"
}

// GetOperation returns the operation
func (r *GroupRoleResponse) GetOperation() string {
	return "post"
}

// GetVersion returns the API version
func (r *GroupRoleResponse) GetVersion() string {
	return "v1"
}

// GetResponseID returns the response ID
func (r *GroupRoleResponse) GetResponseID() string {
	return ""
}

// GetHeaders returns response headers
func (r *GroupRoleResponse) GetHeaders() map[string][]string {
	return nil
}

// GetBody returns the response body
func (r *GroupRoleResponse) GetBody() interface{} {
	return r
}

// GetContext returns response context
func (r *GroupRoleResponse) GetContext() map[string]interface{} {
	return nil
}

// ToProto converts to protocol buffer format
func (r *GroupRoleResponse) ToProto() interface{} {
	return nil
}

// String returns a string representation
func (r *GroupRoleResponse) String() string {
	return fmt.Sprintf("GroupRoleResponse{GroupID: %s, Role: %s, Message: %s}",
		r.GroupID, r.Role.Name, r.Message)
}

// GetType returns the response type for RemoveGroupRoleResponse
func (r *RemoveGroupRoleResponse) GetType() string {
	return "RemoveGroupRoleResponse"
}

// String returns a string representation for RemoveGroupRoleResponse
func (r *RemoveGroupRoleResponse) String() string {
	return fmt.Sprintf("RemoveGroupRoleResponse{GroupID: %s, RoleID: %s, Message: %s}",
		r.GroupID, r.RoleID, r.Message)
}
