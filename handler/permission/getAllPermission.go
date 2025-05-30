package permission

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetAllPermissionsRestApi gets all permissions with optional filtering
// @Summary Get all permissions
// @Description Retrieves all permissions with optional filtering
// @Tags Permissions
// @Accept json
// @Produce json
// @Param roleId query string false "Filter by role ID"
// @Param resource query string false "Filter by resource"
// @Param action query string false "Filter by action"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} helper.Response{data=[]model.Permission} "Permissions retrieved successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid filter parameters"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve permissions"
// @Router /permissions [get]
func (h *PermissionHandler) GetAllPermissionsRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	// Get query parameters
	if roleID := c.Query("roleId"); roleID != "" {
		filter["roleId"] = roleID
	}
	if resource := c.Query("resource"); resource != "" {
		filter["resource"] = resource
	}
	if action := c.Query("action"); action != "" {
		filter["action"] = action
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	permissions, err := h.permissionService.FindPermissions(filter, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permissions retrieved successfully", permissions)
}
