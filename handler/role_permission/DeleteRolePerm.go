package rolepermission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// DeleteRolePermissionRestApi deletes a role-permission relationship
// @Summary Delete role-permission
// @Description Deletes a role-permission relationship by its ID
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param id path string true "Role-Permission ID"
// @Success 200 {object} helper.Response "Role-permission deleted successfully"
// @Failure 404 {object} helper.ErrorResponse "Role-permission not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to delete role-permission"
// @Router /role-permissions/{id} [delete]
func (s *RolePermissionHandler) DeleteRolePermissionRestApi(c *gin.Context) {
	id := c.Param("id")
	err := s.rolePermissionService.DeleteByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Role-permission deleted successfully", nil)
}
