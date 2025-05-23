package action

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// GetAllActionsRestApi retrieves actions with optional filtering and pagination
// @Summary Get actions with pagination
// @Description Retrieves actions with optional filtering by ID or name and pagination support
// @Tags Actions
// @Accept json
// @Produce json
// @Param id query string false "Filter by action ID"
// @Param name query string false "Filter by action name"
// @Param page query int false "Page number (starts from 1)"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} helper.Response{data=[]model.Action} "List of actions retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve actions"
// @Router /actions [get]
func (s *ActionHandler) GetAllActionsRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	// Get filter parameters
	if id := c.Query("id"); id != "" {
		filter["id"] = id
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = "%" + name + "%"
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	actions, err := s.actionService.FindActions(filter, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Actions retrieved successfully", actions)
}
