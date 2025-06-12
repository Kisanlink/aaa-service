package permission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetPermissionByIDRestApi gets a permission by ID
// @Summary Get permission by ID
// @Description Retrieves a permission by its ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} helper.Response{data=model.Permission} "Permission retrieved successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid permission ID"
// @Failure 404 {object} helper.ErrorResponse "Permission not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve permission"
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermissionByIDRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Permission ID is required"})
		return
	}

	permission, err := h.permissionService.GetPermissionByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission retrieved successfully", permission)
}
