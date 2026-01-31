package permissions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeletePermission handles DELETE /api/v1/permissions/:id
//
//	@Summary		Delete permission
//	@Description	Delete a permission by its unique identifier
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Permission ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		409	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v1/permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	permissionID := c.Param("id")

	var deletedBy string
	if userID, exists := c.Get("user_id"); exists {
		deletedBy, _ = userID.(string)
	}

	h.logger.Info("Deleting permission",
		zap.String("permissionID", permissionID),
		zap.String("deletedBy", deletedBy))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	// Check if permission exists
	permission, err := h.permissionService.GetPermissionByID(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.Error("Failed to get permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusNotFound, "Permission not found", err)
		return
	}

	// Check if permission is in use (assigned to any roles)
	// This check would ideally be in the service layer
	permissions, err := h.permissionService.GetPermissionsForRole(c.Request.Context(), permissionID)
	if err == nil && len(permissions) > 0 {
		h.responder.SendError(
			c,
			http.StatusConflict,
			"Permission is assigned to one or more roles. Remove assignments first.",
			nil,
		)
		return
	}

	// Delete permission through service
	if err := h.permissionService.DeletePermission(c.Request.Context(), permissionID); err != nil {
		h.logger.Error("Failed to delete permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to delete permission", err)
		return
	}

	h.logger.Info("Permission deleted successfully", zap.String("permissionID", permission.ID))
	c.Status(http.StatusNoContent)
}
