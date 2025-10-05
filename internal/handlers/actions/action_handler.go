package actions

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/internal/entities/requests/actions"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ActionHandler handles HTTP requests for action operations
type ActionHandler struct {
	actionService *services.ActionService
	validator     interfaces.Validator
	responder     interfaces.Responder
	logger        *zap.Logger
}

// NewActionHandler creates a new ActionHandler instance
func NewActionHandler(
	actionService *services.ActionService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) *ActionHandler {
	return &ActionHandler{
		actionService: actionService,
		validator:     validator,
		responder:     responder,
		logger:        logger,
	}
}

// CreateAction handles POST /actions
//
//	@Summary		Create a new action
//	@Description	Create a new action with the provided information
//	@Tags			actions
//	@Accept			json
//	@Produce		json
//	@Param			action	body		actions.CreateActionRequest	true	"Action creation data"
//	@Success		201		{object}	actions.ActionResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/actions [post]
func (h *ActionHandler) CreateAction(c *gin.Context) {
	h.logger.Info("Creating action")

	var req actions.CreateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Create action through service
	actionResponse, err := h.actionService.CreateAction(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create action", zap.Error(err))
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, conflictErr.Error(), conflictErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Action created successfully", zap.String("actionID", actionResponse.ID))
	h.responder.SendSuccess(c, http.StatusCreated, actionResponse)
}

// GetAction handles GET /actions/:id
//
//	@Summary		Get an action by ID
//	@Description	Retrieve an action by its ID
//	@Tags			actions
//	@Produce		json
//	@Param			id	path		string	true	"Action ID"
//	@Success		200	{object}	actions.ActionResponse
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/actions/{id} [get]
func (h *ActionHandler) GetAction(c *gin.Context) {
	actionID := c.Param("id")
	if actionID == "" {
		h.responder.SendValidationError(c, []string{"action ID is required"})
		return
	}

	h.logger.Info("Getting action", zap.String("actionID", actionID))

	// Get action through service
	actionResponse, err := h.actionService.GetAction(c.Request.Context(), actionID)
	if err != nil {
		h.logger.Error("Failed to get action", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, actionResponse)
}

// UpdateAction handles PUT /actions/:id
//
//	@Summary		Update an action
//	@Description	Update an existing action with the provided information
//	@Tags			actions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Action ID"
//	@Param			action	body		actions.UpdateActionRequest	true	"Action update data"
//	@Success		200		{object}	actions.ActionResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		404		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/actions/{id} [put]
func (h *ActionHandler) UpdateAction(c *gin.Context) {
	actionID := c.Param("id")
	if actionID == "" {
		h.responder.SendValidationError(c, []string{"action ID is required"})
		return
	}

	h.logger.Info("Updating action", zap.String("actionID", actionID))

	var req actions.UpdateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Additional validation using validator service
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Error("Struct validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Update action through service
	actionResponse, err := h.actionService.UpdateAction(c.Request.Context(), actionID, &req)
	if err != nil {
		h.logger.Error("Failed to update action", zap.Error(err))
		if validationErr, ok := err.(*errors.ValidationError); ok {
			h.responder.SendValidationError(c, []string{validationErr.Error()})
			return
		}
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Action updated successfully", zap.String("actionID", actionID))
	h.responder.SendSuccess(c, http.StatusOK, actionResponse)
}

// DeleteAction handles DELETE /actions/:id
//
//	@Summary		Delete an action
//	@Description	Soft delete an action by ID
//	@Tags			actions
//	@Produce		json
//	@Param			id	path	string	true	"Action ID"
//	@Success		204	"Action deleted successfully"
//	@Failure		400	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/v2/actions/{id} [delete]
func (h *ActionHandler) DeleteAction(c *gin.Context) {
	actionID := c.Param("id")
	if actionID == "" {
		h.responder.SendValidationError(c, []string{"action ID is required"})
		return
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responder.SendError(c, http.StatusUnauthorized, "user ID not found in context", nil)
		return
	}

	h.logger.Info("Deleting action", zap.String("actionID", actionID), zap.String("deletedBy", userID.(string)))

	// Delete action through service
	err := h.actionService.DeleteAction(c.Request.Context(), actionID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to delete action", zap.Error(err))
		if notFoundErr, ok := err.(*errors.NotFoundError); ok {
			h.responder.SendError(c, http.StatusNotFound, notFoundErr.Error(), notFoundErr)
			return
		}
		h.responder.SendInternalError(c, err)
		return
	}

	h.logger.Info("Action deleted successfully", zap.String("actionID", actionID))
	c.Status(http.StatusNoContent)
}

// ListActions handles GET /actions
//
//	@Summary		List actions
//	@Description	Retrieve a paginated list of actions
//	@Tags			actions
//	@Produce		json
//	@Param			limit	query		int	false	"Number of actions to return (default: 10, max: 100)"
//	@Param			offset	query		int	false	"Number of actions to skip (default: 0)"
//	@Success		200		{object}	actions.ActionListResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		500		{object}	responses.ErrorResponse
//	@Router			/api/v2/actions [get]
func (h *ActionHandler) ListActions(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Info("Listing actions", zap.Int("limit", limit), zap.Int("offset", offset))

	// List actions through service
	actionsResponse, err := h.actionService.ListActions(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list actions", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, actionsResponse)
}

// GetActionsByService handles GET /actions/service/:serviceName
//
//	@Summary		Get actions by service
//	@Description	Retrieve actions for a specific service
//	@Tags			actions
//	@Produce		json
//	@Param			serviceName	path		string	true	"Service name"
//	@Param			limit		query		int		false	"Number of actions to return (default: 10, max: 100)"
//	@Param			offset		query		int		false	"Number of actions to skip (default: 0)"
//	@Success		200			{object}	actions.ActionListResponse
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		500			{object}	responses.ErrorResponse
//	@Router			/api/v2/actions/service/{serviceName} [get]
func (h *ActionHandler) GetActionsByService(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		h.responder.SendValidationError(c, []string{"service name is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	h.logger.Info("Getting actions by service", zap.String("serviceName", serviceName), zap.Int("limit", limit), zap.Int("offset", offset))

	// Get actions by service through service
	actionsResponse, err := h.actionService.GetActionsByService(c.Request.Context(), serviceName, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get actions by service", zap.Error(err))
		h.responder.SendInternalError(c, err)
		return
	}

	h.responder.SendSuccess(c, http.StatusOK, actionsResponse)
}
