package rolepermission

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


type GetRolePermissionByRoleNameResponse struct {
	StatusCode int                  `json:"status_code"`
	Success    bool                 `json:"success"`
	Message    string               `json:"message"`
	Data       *RolePermissionResponse `json:"data"`
}

func (s *ConnectRolePermissionServer) GetRolePermissionByRoleNameRestApi(c *gin.Context) {
	roleName := c.Query("role")
	if roleName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Role name is required as query parameter 'role'",
			"success": false,
		})
		return
	}

	role, err := s.RoleRepo.GetRoleByName(c.Request.Context(), roleName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Role not found: " + roleName,
			"success": false,
		})
		return
	}

	rolePermissions, err := s.RolePermissionRepo.GetRolePermissionsByRoleID(c.Request.Context(), role.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch role-permission connections",
			"success": false,
		})
		return
	}

	if len(rolePermissions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No permissions found for role: " + roleName,
			"success": false,
		})
		return
	}
	response := &RolePermissionResponse{
		ID:        rolePermissions[0].ID,
		CreatedAt: rolePermissions[0].CreatedAt,
		UpdatedAt: rolePermissions[0].UpdatedAt,
		Role: &ConnRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		},
		Permissions: make([]*ConnPermission, 0),
		IsActive:   rolePermissions[0].IsActive,
	}

	// Add permissions
	for _, rp := range rolePermissions {
		if !IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			response.Permissions = append(response.Permissions, &ConnPermission{
				ID:            rp.Permission.ID,
				Name:          rp.Permission.Name,
				Description:   rp.Permission.Description,
				Action:        rp.Permission.Action,
				Resource:      rp.Permission.Resource,
				Source:        rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime,
				ValidEndTime:   rp.Permission.ValidEndTime,
				CreatedAt:     rp.Permission.CreatedAt,
				UpdatedAt:     rp.Permission.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, GetRolePermissionByRoleNameResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Role with Permissions fetched successfully for role: " + roleName,
		Data:       response,
	})
}
