package rolepermission

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

type AssignPermissionRequest struct {
	Role        string   `json:"role" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

type ConnRolePermissionResponse struct {
	ID          string            `json:"id"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Role        *ConnRole         `json:"role"`
	Permissions []*ConnPermission `json:"permissions"`
	IsActive    bool              `json:"is_active"`
}

type AssignPermissionResponse struct {
	StatusCode    int                         `json:"status_code"`
	Success       bool                        `json:"success"`
	Message       string                      `json:"message"`
	Data          *ConnRolePermissionResponse `json:"data"`
	DataTimeStamp string                      `json:"data_time_stamp"`
}

func (s *ConnectRolePermissionServer) AssignPermissionRestApi(c *gin.Context) {
	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Role == "" || len(req.Permissions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both role_name and permission_names are required"})
		return
	}

	ctx := c.Request.Context()
	role, err := s.RoleRepo.GetRoleByName(ctx, req.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role with name " + req.Role + " not found"})
		return
	}

	permissionIDs := make([]string, 0)
	for _, permissionName := range req.Permissions {
		permission, err := s.PermissionRepo.FindPermissionByName(ctx, permissionName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission with name " + permissionName + " not found"})
			return
		}
		permissionIDs = append(permissionIDs, permission.ID)

		existing, err := s.RolePermissionRepo.GetRolePermissionByNames(ctx, req.Role, permissionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing role-permission connection"})
			return
		}
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Permission '" + permissionName + "' is already assigned to role '" + req.Role + "'",
			})
			return
		}
	}

	var rolePermissions []*model.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermission := &model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permissionID,
			IsActive:     true,
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}

	if err := s.RolePermissionRepo.CreateRolePermissions(ctx, rolePermissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role-permission connections"})
		return
	}

	roles, permissions, actions, usernames, err := s.userRepo.FindRoleUsersAndPermissionsByRoleId(ctx, role.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user roles and permissions"})
		return
	}
	// log.Println(roles, permissions, actions, usernames)

	for _, username := range usernames {
		deleteResponse, err := client.DeleteUserRoleRelationship(
			username,
			roles,
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to delete relationships for user %s: %v", username, err)
			continue
		}
		log.Printf("User roles and permissions deleted successfully for %s: %s", username, deleteResponse)

		createResponse, err := client.CreateUserRoleRelationship(
			username,
			helper.LowerCaseSlice(roles),
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to create relationships for user %s: %v", username, err)
			continue
		}
		log.Printf("Relationships created successfully for %s: %v", username, createResponse)
	}

	fetchedRolePermissions, err := s.RolePermissionRepo.GetRolePermissionsByRoleID(ctx, role.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role-permission connections"})
		return
	}

	if len(fetchedRolePermissions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No permissions found for this role"})
		return
	}

	var rolePermissionPtrs []*model.RolePermission
	for i := range fetchedRolePermissions {
		rolePermissionPtrs = append(rolePermissionPtrs, &fetchedRolePermissions[i])
	}

	response := &ConnRolePermissionResponse{
		ID:        rolePermissionPtrs[0].ID,
		CreatedAt: rolePermissionPtrs[0].CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: rolePermissionPtrs[0].UpdatedAt.Format(time.RFC3339Nano),
		Role: &ConnRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
			CreatedAt:   role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   role.UpdatedAt.Format(time.RFC3339Nano),
		},
		Permissions: []*ConnPermission{},
		IsActive:    rolePermissionPtrs[0].IsActive,
	}

	for _, rp := range rolePermissionPtrs {
		if !IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			response.Permissions = append(response.Permissions, &ConnPermission{
				ID:             rp.Permission.ID,
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				Resource:       rp.Permission.Resource,
				Source:         rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime.Format(time.RFC3339Nano),
				ValidEndTime:   rp.Permission.ValidEndTime.Format(time.RFC3339Nano),
				CreatedAt:      rp.Permission.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:      rp.Permission.UpdatedAt.Format(time.RFC3339Nano),
			})
		}
	}

	c.JSON(http.StatusCreated, &AssignPermissionResponse{
		StatusCode:    http.StatusCreated,
		Success:       true,
		Message:       "Role with Permission created successfully",
		Data:          response,
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
	})
}
