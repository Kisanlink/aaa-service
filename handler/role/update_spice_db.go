package role

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// UpdateSpiceDb schema
// @Summary update spice db schema
// @Description update schema by Retrieves all roles
// @Tags SpiceDB
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=[]model.Role} "Roles retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve roles"
// @Router /update/schema [get]
func (h *RoleHandler) UpdateSpiceDb(c *gin.Context) {
	filter := make(map[string]interface{})
	roles, err := h.roleService.FindRoles(filter, 0, 0)
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
	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Schema updated successfully", roles)
}
