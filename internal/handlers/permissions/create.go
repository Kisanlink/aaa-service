package permissions

import (
	"net/http"

	reqPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/permissions"
	respPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/permissions"
	"github.com/Kisanlink/aaa-service/v2/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreatePermission handles POST /api/v1/permissions
//
//	@Summary		Create a new permission
//	@Description	Create a new permission with name, resource, and action
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			permission	body		reqPermissions.CreatePermissionRequest	true	"Permission creation data"
//	@Success		201			{object}	respPermissions.PermissionResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		409			{object}	map[string]interface{}	"Permission with the same name already exists"
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	h.logger.Info("Creating permission")

	var req reqPermissions.CreatePermissionRequest
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

	// Convert to model
	permission := req.ToModel()

	// Create permission through service
	if err := h.permissionService.CreatePermission(c.Request.Context(), permission); err != nil {
		h.logger.Error("Failed to create permission", zap.Error(err))

		// Handle conflict error (duplicate permission)
		if conflictErr, ok := err.(*errors.ConflictError); ok {
			h.responder.SendError(c, http.StatusConflict, "Permission already exists", conflictErr)
			return
		}

		h.responder.SendError(c, http.StatusInternalServerError, "Failed to create permission", err)
		return
	}

	// Get the created permission to ensure we have all related data
	createdPermission, err := h.permissionService.GetPermissionByID(c.Request.Context(), permission.ID)
	if err != nil {
		h.logger.Error("Failed to get created permission", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get created permission", err)
		return
	}

	// Convert to response
	response := respPermissions.NewPermissionResponse(createdPermission)

	h.logger.Info("Permission created successfully", zap.String("permissionID", permission.ID))
	h.responder.SendSuccess(c, http.StatusCreated, response)
}
