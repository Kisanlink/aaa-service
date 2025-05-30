package permission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// DeletePermissionRestApi deletes a permission
// @Summary Delete a permission
// @Description Deletes an existing permission by its ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Success 200 {object} helper.Response "Permission deleted successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid permission ID"
// @Failure 404 {object} helper.ErrorResponse "Permission not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to delete permission"
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermissionRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Permission ID is required"})
		return
	}

	if err := h.permissionService.DeletePermission(id); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission deleted successfully", nil)
}
