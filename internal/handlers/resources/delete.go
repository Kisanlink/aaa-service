package resources

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeleteResource handles DELETE /api/v2/resources/:id
//
//	@Summary		Delete resource
//	@Description	Delete a resource by its unique identifier
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Resource ID"
//	@Param			cascade	query		boolean	false	"Delete children recursively"	default(false)
//	@Success		204		{object}	nil
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		409		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/resources/{id} [delete]
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	resourceID := c.Param("id")
	h.logger.Info("Deleting resource", zap.String("resourceID", resourceID))

	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	// Check if cascade delete is requested
	cascade := c.DefaultQuery("cascade", "false") == "true"

	// Check if resource exists
	exists, err := h.resourceService.Exists(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to check resource existence", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to check resource", err)
		return
	}

	if !exists {
		h.responder.SendError(c, http.StatusNotFound, "Resource not found", nil)
		return
	}

	// Check if resource has children (only if cascade is not enabled)
	if !cascade {
		hasChildren, err := h.resourceService.HasChildren(c.Request.Context(), resourceID)
		if err != nil {
			h.logger.Error("Failed to check children", zap.Error(err), zap.String("resourceID", resourceID))
			h.responder.SendError(c, http.StatusInternalServerError, "Failed to check children", err)
			return
		}

		if hasChildren {
			h.responder.SendError(
				c,
				http.StatusConflict,
				"Resource has children. Use cascade=true to delete recursively",
				nil,
			)
			return
		}
	}

	// Delete resource through service
	if cascade {
		err = h.resourceService.DeleteCascade(c.Request.Context(), resourceID)
	} else {
		err = h.resourceService.Delete(c.Request.Context(), resourceID)
	}

	if err != nil {
		h.logger.Error("Failed to delete resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to delete resource", err)
		return
	}

	h.logger.Info("Resource deleted successfully",
		zap.String("resourceID", resourceID),
		zap.Bool("cascade", cascade))

	c.Status(http.StatusNoContent)
}
