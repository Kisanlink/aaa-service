package role

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetAllRolesRestApi retrieves all roles with optional filtering
// @Summary Get  roles
// @Description Retrieves all roles with optional filtering by ID or name
// @Tags Roles
// @Accept json
// @Produce json
// @Param id query string false "Filter by role ID"
// @Param name query string false "Filter by role name"
// @Success 200 {object} helper.Response{data=[]model.Role} "Roles retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve roles"
// @Router /roles [get]
func (h *RoleHandler) GetAllRolesRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	// Get query parameters
	if id := c.Query("id"); id != "" {
		filter["id"] = id
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	roles, err := h.roleService.FindRoles(filter)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Roles retrieved successfully", roles)
}
