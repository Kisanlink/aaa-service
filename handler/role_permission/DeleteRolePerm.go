package rolepermission

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/handler/spicedb"
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
// @Router /assign-permissions/{id} [delete]
func (h *RolePermissionHandler) DeleteRolePermissionRestApi(c *gin.Context) {
	id := c.Param("id")
	err := h.rolePermissionService.DeleteByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}
	// Get all roles to build SpiceDB schema
	roles, err := h.roleService.FindRoles(map[string]interface{}{}, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	resource, err := h.resourceService.FindResources(map[string]interface{}{}, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	// Generate SpiceDB schema definitions
	schemaDefinitions := helper.GenerateSpiceDBSchema(roles, resource)

	// Update SpiceDB schema
	_, err = client.UpdateSchema(schemaDefinitions)
	if err != nil {
		log.Printf("Failed to update SpiceDB schema: %v", err)

	}

	spicedb.UpdateSpiceDBData(h.roleService, h.userService)
	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Role-permission deleted successfully", nil)
}
