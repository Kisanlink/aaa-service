package action

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdateActionRestApi updates an existing action
// @Summary Update an action
// @Description Updates an existing action with the provided details
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Action ID"
// @Param request body model.CreateActionRequest true "Action update data"
// @Success 200 {object} helper.Response{data=model.Action} "Action updated successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 404 {object} helper.Response "Action not found"
// @Failure 500 {object} helper.Response "Failed to update action"
// @Router /actions/{id} [put]
func (s *ActionHandler) UpdateActionRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Action ID is required"})
		return
	}

	var req model.Action
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	// Update only allowed fields
	updateData := model.Action{
		Name: req.Name, // Only update name field
	}

	if err := s.actionService.UpdateAction(id, updateData); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Fetch updated action
	updatedAction, err := s.actionService.FindActionByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Action updated successfully", updatedAction)
}
