package action

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type ActionHandler struct {
	actionService services.ActionServiceInterface
}

func NewActionHandler(
	actionService services.ActionServiceInterface,
) *ActionHandler {
	return &ActionHandler{
		actionService: actionService,
	}
}

// CreateActionRestApi creates a new action
// @Summary Create a new action
// @Description Creates a new action with the provided details
// @Tags Actions
// @Accept json
// @Produce json
// @Param request body model.CreateActionRequest true "Action creation data"
// @Success 201 {object} helper.Response{data=model.Action} "Action created successfully"
// @Failure 400 {object} helper.Response "Invalid request or missing required fields"
// @Failure 409 {object} helper.Response "Action already exists"
// @Failure 500 {object} helper.Response "Failed to create action"
// @Router /actions [post]
func (s *ActionHandler) CreateActionRestApi(c *gin.Context) {
	var req model.Action
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Name == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Action Name is required"})
		return
	}

	// Check if action already exists
	if err := s.actionService.CheckIfActionExists(req.Name); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{err.Error()})
		return
	}

	// Create new action
	newAction := model.Action{
		Name: req.Name,
	}

	if err := s.actionService.CreateAction(&newAction); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	actionResponse := &model.Action{
		Base: model.Base{
			ID:        newAction.ID,
			CreatedAt: newAction.CreatedAt,
			UpdatedAt: newAction.UpdatedAt,
		},
		Name: newAction.Name,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Action created successfully", actionResponse)
}
