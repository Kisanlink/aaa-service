package rolepermission

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetRolePermissionByIDRestApi retrieves a single role-permission by ID
// @Summary Get role-permission by ID
// @Description Retrieves a single role-permission relationship by its ID
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param id path string true "Role-Permission ID"
// @Success 200 {object} helper.Response{data=model.GetRolePermissionResponse} "Role-permission retrieved successfully"
// @Failure 404 {object} helper.ErrorResponse "Role-permission not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve role-permission"
// @Router /assign-permissions/{id} [get]
func (s *RolePermissionHandler) GetRolePermissionByIDRestApi(c *gin.Context) {
	id := c.Param("id")

	rolePermission, err := s.rolePermissionService.FetchByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	if rolePermission == nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound,
			[]string{"role-permission not found"})
		return
	}

	// Transform to response format
	var roleResponse *model.Role
	if rolePermission.RoleID != "" {
		role, err := s.roleService.FindRoleByID(rolePermission.RoleID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
				[]string{fmt.Sprintf("failed to fetch role details: %v", err)})
			return
		}

		roleResponse = &model.Role{
			Base: model.Base{
				ID:        role.ID,
				CreatedAt: role.CreatedAt,
				UpdatedAt: role.UpdatedAt,
			},
			Name:        role.Name,
			Description: role.Description,
		}
	}

	var permission *model.Permission
	if rolePermission.PermissionID != "" {
		permission, err = s.permissionService.GetPermissionByID(rolePermission.PermissionID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
				[]string{fmt.Sprintf("failed to fetch permission details: %v", err)})
			return
		}
	}

	response := model.GetRolePermissionResponse{
		ID:         rolePermission.ID,
		CreatedAt:  rolePermission.CreatedAt,
		UpdatedAt:  rolePermission.UpdatedAt,
		Role:       roleResponse,
		Permission: permission,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Role-permission retrieved successfully", response)
}
