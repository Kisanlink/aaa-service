package role_assignments

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// RoleAssignmentResponse represents the result of a role assignment operation
// @Description Response structure for role assignment operations
type RoleAssignmentResponse struct {
	Success   bool                `json:"success" example:"true"`
	Message   string              `json:"message" example:"Assignment successful"`
	Data      *RoleAssignmentData `json:"data"`
	Timestamp time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string              `json:"request_id" example:"req_abc123"`
}

// RoleAssignmentData contains the assignment result details
type RoleAssignmentData struct {
	RoleID              string                    `json:"role_id" example:"ROLE_abc123"`
	RoleName            string                    `json:"role_name" example:"admin"`
	AssignedPermissions []*AssignedPermissionInfo `json:"assigned_permissions,omitempty"`
	AssignedResources   []*AssignedResourceInfo   `json:"assigned_resources,omitempty"`
	FailedAssignments   []*FailedAssignmentInfo   `json:"failed_assignments,omitempty"`
}

// AssignedPermissionInfo contains information about an assigned permission
type AssignedPermissionInfo struct {
	PermissionID   string    `json:"permission_id" example:"PERM_abc123"`
	PermissionName string    `json:"permission_name" example:"manage_users"`
	ResourceName   string    `json:"resource_name,omitempty" example:"User Management"`
	ActionName     string    `json:"action_name,omitempty" example:"manage"`
	AssignedAt     time.Time `json:"assigned_at" example:"2024-01-01T00:00:00Z"`
}

// AssignedResourceInfo contains information about an assigned resource
type AssignedResourceInfo struct {
	ResourceID   string    `json:"resource_id" example:"RES_abc123"`
	ResourceName string    `json:"resource_name" example:"User Management"`
	ResourceType string    `json:"resource_type" example:"aaa/user"`
	Actions      []string  `json:"actions" example:"read,write,delete"`
	AssignedAt   time.Time `json:"assigned_at" example:"2024-01-01T00:00:00Z"`
}

// FailedAssignmentInfo contains information about a failed assignment
type FailedAssignmentInfo struct {
	ID     string `json:"id" example:"PERM_abc123"`
	Type   string `json:"type" example:"permission"`
	Reason string `json:"reason" example:"Permission not found"`
}

// NewRolePermissionAssignmentResponse creates a new response for permission assignment
func NewRolePermissionAssignmentResponse(
	role *models.Role,
	assignments []*models.RolePermission,
	failed []*FailedAssignmentInfo,
	requestID string,
) *RoleAssignmentResponse {
	assignedPermissions := make([]*AssignedPermissionInfo, 0, len(assignments))
	for _, rp := range assignments {
		info := &AssignedPermissionInfo{
			PermissionID: rp.PermissionID,
			AssignedAt:   rp.CreatedAt,
		}

		// Add permission details if available
		if rp.Permission != nil {
			info.PermissionName = rp.Permission.Name
			if rp.Permission.Resource != nil {
				info.ResourceName = rp.Permission.Resource.Name
			}
			if rp.Permission.Action != nil {
				info.ActionName = rp.Permission.Action.Name
			}
		}

		assignedPermissions = append(assignedPermissions, info)
	}

	message := "Permissions assigned successfully"
	if len(failed) > 0 {
		message = "Permissions assigned with some failures"
	}

	return &RoleAssignmentResponse{
		Success: true,
		Message: message,
		Data: &RoleAssignmentData{
			RoleID:              role.ID,
			RoleName:            role.Name,
			AssignedPermissions: assignedPermissions,
			FailedAssignments:   failed,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}

// NewRoleResourceAssignmentResponse creates a new response for resource assignment
func NewRoleResourceAssignmentResponse(
	role *models.Role,
	assignments []*models.ResourcePermission,
	failed []*FailedAssignmentInfo,
	requestID string,
) *RoleAssignmentResponse {
	assignedResources := make([]*AssignedResourceInfo, 0)

	// Group by resource
	resourceMap := make(map[string]*AssignedResourceInfo)
	for _, rp := range assignments {
		if existing, ok := resourceMap[rp.ResourceID]; ok {
			existing.Actions = append(existing.Actions, rp.Action)
		} else {
			info := &AssignedResourceInfo{
				ResourceID:   rp.ResourceID,
				ResourceType: rp.ResourceType,
				Actions:      []string{rp.Action},
				AssignedAt:   rp.CreatedAt,
			}

			// Add resource name if available
			if rp.Resource != nil {
				info.ResourceName = rp.Resource.Name
			}

			resourceMap[rp.ResourceID] = info
		}
	}

	for _, info := range resourceMap {
		assignedResources = append(assignedResources, info)
	}

	message := "Resources assigned successfully"
	if len(failed) > 0 {
		message = "Resources assigned with some failures"
	}

	return &RoleAssignmentResponse{
		Success: true,
		Message: message,
		Data: &RoleAssignmentData{
			RoleID:            role.ID,
			RoleName:          role.Name,
			AssignedResources: assignedResources,
			FailedAssignments: failed,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}

// RolePermissionsListResponse represents a list of permissions assigned to a role
// @Description Response structure for listing role permissions
type RolePermissionsListResponse struct {
	Success   bool                 `json:"success" example:"true"`
	Message   string               `json:"message" example:"Role permissions retrieved successfully"`
	Data      *RolePermissionsData `json:"data"`
	Timestamp time.Time            `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string               `json:"request_id" example:"req_abc123"`
}

// RolePermissionsData contains the role permissions list
type RolePermissionsData struct {
	RoleID      string                    `json:"role_id" example:"ROLE_abc123"`
	RoleName    string                    `json:"role_name" example:"admin"`
	Permissions []*AssignedPermissionInfo `json:"permissions"`
}

// NewRolePermissionsListResponse creates a new response for role permissions list
func NewRolePermissionsListResponse(
	role *models.Role,
	permissions []*models.RolePermission,
	requestID string,
) *RolePermissionsListResponse {
	assignedPermissions := make([]*AssignedPermissionInfo, 0, len(permissions))
	for _, rp := range permissions {
		info := &AssignedPermissionInfo{
			PermissionID: rp.PermissionID,
			AssignedAt:   rp.CreatedAt,
		}

		if rp.Permission != nil {
			info.PermissionName = rp.Permission.Name
			if rp.Permission.Resource != nil {
				info.ResourceName = rp.Permission.Resource.Name
			}
			if rp.Permission.Action != nil {
				info.ActionName = rp.Permission.Action.Name
			}
		}

		assignedPermissions = append(assignedPermissions, info)
	}

	return &RolePermissionsListResponse{
		Success: true,
		Message: "Role permissions retrieved successfully",
		Data: &RolePermissionsData{
			RoleID:      role.ID,
			RoleName:    role.Name,
			Permissions: assignedPermissions,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}
