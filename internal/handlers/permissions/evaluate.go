package permissions

import (
	"net/http"

	reqPermissions "github.com/Kisanlink/aaa-service/internal/entities/requests/permissions"
	respPermissions "github.com/Kisanlink/aaa-service/internal/entities/responses/permissions"
	permissionService "github.com/Kisanlink/aaa-service/internal/services/permissions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EvaluatePermission handles POST /api/v2/permissions/evaluate
//
//	@Summary		Evaluate user permission
//	@Description	Check if a user has permission to perform an action on a resource
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		reqPermissions.EvaluatePermissionRequest	true	"Evaluation request"
//	@Success		200		{object}	respPermissions.EvaluationResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/permissions/evaluate [post]
func (h *PermissionHandler) EvaluatePermission(c *gin.Context) {
	h.logger.Info("Evaluating permission")

	var req reqPermissions.EvaluatePermissionRequest
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

	// Build evaluation context
	evalContext := &permissionService.EvaluationContext{
		OrganizationID: req.GetOrganizationID(),
		GroupID:        req.GetGroupID(),
		CustomAttrs:    req.Context,
	}

	// Evaluate permission through service
	result, err := h.permissionService.EvaluatePermission(
		c.Request.Context(),
		req.UserID,
		req.ResourceType,
		req.ResourceID,
		req.Action,
		evalContext,
	)
	if err != nil {
		h.logger.Error("Failed to evaluate permission", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to evaluate permission", err)
		return
	}

	// Convert to response
	response := respPermissions.NewEvaluationResponse(result, h.getRequestID(c))

	h.logger.Info("Permission evaluated successfully",
		zap.String("userID", req.UserID),
		zap.String("resourceType", req.ResourceType),
		zap.String("resourceID", req.ResourceID),
		zap.String("action", req.Action),
		zap.Bool("allowed", result.Allowed))

	c.JSON(http.StatusOK, response)
}

// EvaluateUserPermission handles POST /api/v2/users/:id/evaluate
//
//	@Summary		Evaluate user-specific permission
//	@Description	Check if a specific user has permission to perform an action on a resource
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string									true	"User ID"
//	@Param			request	body		reqPermissions.EvaluatePermissionRequest	true	"Evaluation request"
//	@Success		200		{object}	respPermissions.EvaluationResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v2/users/{id}/evaluate [post]
func (h *PermissionHandler) EvaluateUserPermission(c *gin.Context) {
	userID := c.Param("id")
	h.logger.Info("Evaluating user-specific permission", zap.String("userID", userID))

	if userID == "" {
		h.responder.SendValidationError(c, []string{"user ID is required"})
		return
	}

	var req reqPermissions.EvaluatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Override user ID from URL parameter
	req.UserID = userID

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Request validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Build evaluation context
	evalContext := &permissionService.EvaluationContext{
		OrganizationID: req.GetOrganizationID(),
		GroupID:        req.GetGroupID(),
		CustomAttrs:    req.Context,
	}

	// Evaluate permission through service
	result, err := h.permissionService.EvaluatePermission(
		c.Request.Context(),
		req.UserID,
		req.ResourceType,
		req.ResourceID,
		req.Action,
		evalContext,
	)
	if err != nil {
		h.logger.Error("Failed to evaluate permission", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to evaluate permission", err)
		return
	}

	// Convert to response
	response := respPermissions.NewEvaluationResponse(result, h.getRequestID(c))

	h.logger.Info("User permission evaluated successfully",
		zap.String("userID", req.UserID),
		zap.String("resourceType", req.ResourceType),
		zap.String("resourceID", req.ResourceID),
		zap.String("action", req.Action),
		zap.Bool("allowed", result.Allowed))

	c.JSON(http.StatusOK, response)
}
