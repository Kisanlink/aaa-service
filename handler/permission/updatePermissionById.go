package permission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdatePermissionRestApi updates a permission
// @Summary Update a permission
// @Description Updates an existing permission with the provided details
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID"
// @Param request body model.CreatePermissionRequest true "Permission update data"
// @Success 200 {object} helper.Response{data=model.Permission} "Permission updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request or missing required fields"
// @Failure 404 {object} helper.ErrorResponse "Permission not found"
// @Failure 409 {object} helper.ErrorResponse "Permission already exists for this role+resource"
// @Failure 500 {object} helper.ErrorResponse "Failed to update permission"
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermissionRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Permission ID is required"})
		return
	}

	var req model.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Resource == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Resource is required"})
		return
	}

	if len(req.Actions) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"At least one action is required"})
		return
	}

	if err := h.permissionService.UpdatePermission(id, req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Get updated permission to return
	updatedPermission, err := h.permissionService.GetPermissionByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission updated successfully", updatedPermission)
}
