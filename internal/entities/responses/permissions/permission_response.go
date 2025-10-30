package permissions

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

// PermissionResponse represents a single permission in API responses
// @Description Complete permission details including associated resource and action information
type PermissionResponse struct {
	ID           string     `json:"id" example:"PERM00000001"`
	Name         string     `json:"name" example:"crop_management_read"`
	Description  string     `json:"description" example:"Permission to view and read crop information in the farm inventory"`
	ResourceID   *string    `json:"resource_id,omitempty" example:"RES1760615540005820900"`
	ResourceName *string    `json:"resource_name,omitempty" example:"crop_management"`
	ActionID     *string    `json:"action_id,omitempty" example:"ACT1760615540005820901"`
	ActionName   *string    `json:"action_name,omitempty" example:"read"`
	IsActive     bool       `json:"is_active" example:"true"`
	CreatedAt    time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt    time.Time  `json:"updated_at" example:"2024-01-20T14:45:00Z"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// NewPermissionResponse creates a new PermissionResponse from a Permission model
func NewPermissionResponse(permission *models.Permission) *PermissionResponse {
	if permission == nil {
		return nil
	}

	response := &PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		ResourceID:  permission.ResourceID,
		ActionID:    permission.ActionID,
		IsActive:    permission.IsActive,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
		DeletedAt:   permission.DeletedAt,
	}

	// Add resource name if available
	if permission.Resource != nil {
		response.ResourceName = &permission.Resource.Name
	}

	// Add action name if available
	if permission.Action != nil {
		response.ActionName = &permission.Action.Name
	}

	return response
}

// PermissionListResponse represents a paginated list of permissions
// @Description Response structure for a list of permissions with pagination
type PermissionListResponse struct {
	Success    bool                `json:"success" example:"true"`
	Message    string              `json:"message" example:"Permissions retrieved successfully"`
	Data       *PermissionListData `json:"data"`
	Pagination *PaginationInfo     `json:"pagination"`
	Timestamp  time.Time           `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	RequestID  string              `json:"request_id" example:"req_abc123"`
}

// PermissionListData contains the actual permission data
type PermissionListData struct {
	Permissions []*PermissionResponse `json:"permissions"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

// NewPermissionListResponse creates a new PermissionListResponse
func NewPermissionListResponse(
	permissions []*models.Permission,
	page, limit, total int,
	requestID string,
) *PermissionListResponse {
	permissionResponses := make([]*PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		permissionResponses = append(permissionResponses, NewPermissionResponse(permission))
	}

	// Prevent division by zero when limit is 0
	var totalPages int
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	} else {
		totalPages = 1
	}
	if totalPages < 1 {
		totalPages = 1
	}

	return &PermissionListResponse{
		Success: true,
		Message: "Permissions retrieved successfully",
		Data: &PermissionListData{
			Permissions: permissionResponses,
		},
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
}
