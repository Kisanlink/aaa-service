package role

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetAllRolesRestApi retrieves roles with optional filtering and pagination
// @Summary Get roles with pagination
// @Description Retrieves roles with optional filtering by ID or name and pagination support
// @Tags Roles
// @Accept json
// @Produce json
// @Param id query string false "Filter by role ID"
// @Param name query string false "Filter by role name"
// @Param page query int false "Page number (starts from 1)"
// @Param limit query int false "Number of items per page"
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

	// Get pagination parameters from query, default to 0 (which means no pagination)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	roles, err := h.roleService.FindRoles(filter, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Roles retrieved successfully", roles)
}
