package rolepermission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)



type AssignPermissionRequest struct {
	Role        string   `json:"role" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type ConnRolePermission struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
	IsActive     bool   `json:"is_active"`
}

type AssignPermissionResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string               `json:"message"`
	Data       []ConnRolePermission `json:"data"`
}

func (s *ConnectRolePermissionServer) AssignPermissionRestApi(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate input
	if req.Role == "" || len(req.Permissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both role and permissions are required"})
		return
	}

	// Get the role
	role, err := s.RoleRepo.GetRoleByName(c.Request.Context(), req.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found: " + req.Role})
		return
	}

	// Get all permissions
	permissionIDs := make([]string, 0, len(req.Permissions))
	for _, permissionName := range req.Permissions {
		permission, err := s.PermissionRepo.FindPermissionByName(c.Request.Context(), permissionName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found: " + permissionName})
			return
		}
		permissionIDs = append(permissionIDs, permission.ID)
	}

	// Create role-permission connections
	var rolePermissions []*model.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermission := &model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permissionID,
			IsActive:     true,
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}

	// Save to database
	if err := s.RolePermissionRepo.CreateRolePermissions(c.Request.Context(), rolePermissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role-permission connections"})
		return
	}

	// Prepare response
	var connRolePermissions []ConnRolePermission
	for _, rp := range rolePermissions {
		connRolePermissions = append(connRolePermissions, ConnRolePermission{
			ID:           rp.ID,
			CreatedAt:    rp.CreatedAt.String(),
			UpdatedAt:    rp.UpdatedAt.String(),
			RoleID:       rp.RoleID,
			PermissionID: rp.PermissionID,
			IsActive:     rp.IsActive,
		})
	}

	response := AssignPermissionResponse{
		StatusCode: http.StatusCreated,
		Success: true,
		Message:    "Role with Permission created successfully",
		Data:       connRolePermissions,
	}

	c.JSON(http.StatusCreated, response)
}