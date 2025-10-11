package permissions

import (
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
)

// AssignmentResponse represents the result of a permission assignment operation
// @Description Response structure for permission assignment
type AssignmentResponse struct {
	Success   bool            `json:"success" example:"true"`
	Message   string          `json:"message" example:"Permissions assigned successfully"`
	Data      *AssignmentData `json:"data"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID string          `json:"request_id" example:"req_abc123"`
}

// AssignmentData contains the assignment result details
type AssignmentData struct {
	RoleID              string                    `json:"role_id" example:"ROLE_abc123"`
	AssignedPermissions []*AssignedPermissionInfo `json:"assigned_permissions,omitempty"`
	AssignedResources   []*AssignedResourceInfo   `json:"assigned_resources,omitempty"`
	FailedAssignments   []*FailedAssignmentInfo   `json:"failed_assignments,omitempty"`
}

// AssignedPermissionInfo contains information about an assigned permission
type AssignedPermissionInfo struct {
	PermissionID   string    `json:"permission_id" example:"PERM_abc123"`
	PermissionName string    `json:"permission_name" example:"manage_users"`
	AssignedAt     time.Time `json:"assigned_at" example:"2024-01-01T00:00:00Z"`
}

// AssignedResourceInfo contains information about an assigned resource
type AssignedResourceInfo struct {
	ResourceID   string    `json:"resource_id" example:"RES_abc123"`
	ResourceName string    `json:"resource_name" example:"User Management"`
	Actions      []string  `json:"actions" example:"read,write,delete"`
	AssignedAt   time.Time `json:"assigned_at" example:"2024-01-01T00:00:00Z"`
}

// FailedAssignmentInfo contains information about a failed assignment
type FailedAssignmentInfo struct {
	ID     string `json:"id" example:"PERM_abc123"`
	Reason string `json:"reason" example:"Permission not found"`
}

// NewAssignmentResponse creates a new AssignmentResponse for permission assignment
func NewAssignmentResponse(
	roleID string,
	permissions []*models.RolePermission,
	failed []*FailedAssignmentInfo,
	requestID string,
) *AssignmentResponse {
	assignedPermissions := make([]*AssignedPermissionInfo, 0, len(permissions))
	for _, rp := range permissions {
		info := &AssignedPermissionInfo{
			PermissionID: rp.PermissionID,
			AssignedAt:   rp.CreatedAt,
		}

		// Add permission name if available
		if rp.Permission != nil {
			info.PermissionName = rp.Permission.Name
		}

		assignedPermissions = append(assignedPermissions, info)
	}

	message := "Permissions assigned successfully"
	if len(failed) > 0 {
		message = "Permissions assigned with some failures"
	}

	return &AssignmentResponse{
		Success: true,
		Message: message,
		Data: &AssignmentData{
			RoleID:              roleID,
			AssignedPermissions: assignedPermissions,
			FailedAssignments:   failed,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}

// NewResourceAssignmentResponse creates a new AssignmentResponse for resource assignment
func NewResourceAssignmentResponse(
	roleID string,
	resources []*models.ResourcePermission,
	failed []*FailedAssignmentInfo,
	requestID string,
) *AssignmentResponse {
	assignedResources := make([]*AssignedResourceInfo, 0, len(resources))

	// Group by resource
	resourceMap := make(map[string]*AssignedResourceInfo)
	for _, rp := range resources {
		if existing, ok := resourceMap[rp.ResourceID]; ok {
			existing.Actions = append(existing.Actions, rp.Action)
		} else {
			info := &AssignedResourceInfo{
				ResourceID: rp.ResourceID,
				Actions:    []string{rp.Action},
				AssignedAt: rp.CreatedAt,
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

	return &AssignmentResponse{
		Success: true,
		Message: message,
		Data: &AssignmentData{
			RoleID:            roleID,
			AssignedResources: assignedResources,
			FailedAssignments: failed,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}
