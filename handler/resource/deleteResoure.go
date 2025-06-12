package resource

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// DeleteResourceRestApi deletes a resource
// @Summary Delete a resource
// @Description Deletes an existing resource by ID
// @Tags Resources
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} helper.Response "Resource deleted successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid resource ID"
// @Failure 500 {object} helper.ErrorResponse "Failed to delete resource"
// @Router /resources/{id} [delete]
func (s *ResourceHandler) DeleteResourceRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Resource ID is required"})
		return
	}

	if err := s.resourceService.DeleteResource(id); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Resource deleted successfully", nil)
}
