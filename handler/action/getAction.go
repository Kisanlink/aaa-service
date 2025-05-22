package action

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetAllActionsRestApi retrieves all actions with optional filtering
// @Summary Get  actions
// @Description Retrieves all actions with optional filtering by ID or name
// @Tags Actions
// @Accept json
// @Produce json
// @Param id query string false "Filter by action ID"
// @Param name query string false "Filter by action name"
// @Success 200 {object} helper.Response{data=[]model.Action} "List of actions retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve actions"
// @Router /actions [get]
func (s *ActionHandler) GetAllActionsRestApi(c *gin.Context) {
	filter := make(map[string]interface{})
	if id := c.Query("id"); id != "" {
		filter["id"] = id
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = "%" + name + "%"
	}

	actions, err := s.actionService.FindActions(filter)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Actions retrieved successfully", actions)
}
