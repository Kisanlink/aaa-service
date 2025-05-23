package role

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// DeleteRoleRestApi deletes a role and its permissions
// @Summary Delete a role
// @Description Deletes a role and all its associated permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} helper.Response "Role deleted successfully"
// @Failure 400 {object} helper.Response "Invalid role ID"
// @Failure 500 {object} helper.Response "Failed to delete role"
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRoleRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Role ID is required"})
		return
	}

	if err := h.roleService.DeleteRole(id); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Get all roles to build SpiceDB schema
	roles, err := h.roleService.FindRoles(map[string]interface{}{}, 0, 0)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Generate SpiceDB schema definitions
	schemaDefinitions := helper.GenerateSpiceDBSchema(roles)

	// Update SpiceDB schema
	_, err = client.UpdateSchema(schemaDefinitions)
	if err != nil {
		log.Printf("Failed to update SpiceDB schema: %v", err)

	}
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role deleted successfully", nil)
}
