package action

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// DeleteActionRestApi deletes an action
// @Summary Delete an action
// @Description Deletes an existing action by ID
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Action ID"
// @Success 200 {object} helper.Response "Action deleted successfully"
// @Failure 400 {object} helper.Response "Invalid action ID"
// @Failure 404 {object} helper.Response "Action not found"
// @Failure 500 {object} helper.Response "Failed to delete action"
// @Router /actions/{id} [delete]
func (s *ActionHandler) DeleteActionRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Action ID is required"})
		return
	}

	if err := s.actionService.DeleteAction(id); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Action deleted successfully", nil)
}
