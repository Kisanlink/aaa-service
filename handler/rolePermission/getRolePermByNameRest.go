package rolepermission

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetRolePermissionByRoleNameRestApi retrieves permissions for a specific role
// @Summary Get permissions by role name
// @Description Retrieves all permissions assigned to a specific role identified by name
// @Tags Role Permissions
// @Accept json
// @Produce json
// @Param role query string true "Role name" example("admin")
// @Success 200 {object} helper.Response{data=model.RolePermissionWrapper} "Role permissions fetched successfully"
// @Failure 400 {object} helper.Response "Role name parameter is required"
// @Failure 404 {object} helper.Response "Role not found or no permissions assigned"
// @Failure 500 {object} helper.Response "Failed to fetch role-permission connections"
// @Router /assign-permissions/by  [get]
func (s *RolePermHandler) GetRolePermissionByRoleNameRestApi(c *gin.Context) {
	roleName := c.Query("role")
	if roleName == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Role name is required as query parameter 'role'"})
		return
	}

	role, err := s.roleService.GetRoleByName(roleName)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Role not found: " + roleName})
		return
	}

	rolePermissions, err := s.rolePermService.GetRolePermissionsByRoleID(role.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch role-permission connections"})
		return
	}

	if len(rolePermissions) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"No permissions found for role: " + roleName})
		return
	}

	response := &model.RolePermissionResponse{
		ID:        rolePermissions[0].ID,
		CreatedAt: rolePermissions[0].CreatedAt,
		UpdatedAt: rolePermissions[0].UpdatedAt,
		Role: &model.Role{
			Base: model.Base{
				ID:        role.ID,
				CreatedAt: role.CreatedAt,
				UpdatedAt: role.UpdatedAt,
			},
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
		},
		Permissions: make([]*model.Permission, 0),
		IsActive:    rolePermissions[0].IsActive,
	}

	// Add permissions
	for _, rp := range rolePermissions {
		if !helper.IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			response.Permissions = append(response.Permissions, &model.Permission{
				Base: model.Base{
					ID:        rp.Permission.ID,
					CreatedAt: rp.Permission.CreatedAt,
					UpdatedAt: rp.Permission.UpdatedAt,
				},
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				Resource:       rp.Permission.Resource,
				Source:         rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime,
				ValidEndTime:   rp.Permission.ValidEndTime,
			})
		}
	}

	wrappedResponse := model.RolePermissionWrapper{
		Data:      response,
		Timestamp: time.Now().Format(time.RFC3339Nano),
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role with Permissions fetched successfully for role: "+roleName, wrappedResponse)
}
