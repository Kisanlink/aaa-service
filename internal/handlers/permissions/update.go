package permissions

import (
	"net/http"

	reqPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/permissions"
	respPermissions "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/permissions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdatePermission handles PUT /api/v1/permissions/:id
//
//	@Summary		Update permission
//	@Description	Update an existing permission
//	@Tags			permissions
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string									true	"Permission ID"
//	@Param			permission	body		reqPermissions.UpdatePermissionRequest	true	"Permission update data"
//	@Success		200			{object}	respPermissions.PermissionResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	permissionID := c.Param("id")
	h.logger.Info("Updating permission", zap.String("permissionID", permissionID))

	if permissionID == "" {
		h.responder.SendValidationError(c, []string{"permission ID is required"})
		return
	}

	var req reqPermissions.UpdatePermissionRequest
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

	// Check if there are any updates
	if !req.HasUpdates() {
		h.responder.SendValidationError(c, []string{"no fields to update"})
		return
	}

	// Get existing permission
	permission, err := h.permissionService.GetPermissionByID(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.Error("Failed to get permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusNotFound, "Permission not found", err)
		return
	}

	// Apply updates
	if req.Name != nil {
		permission.Name = *req.Name
	}
	if req.Description != nil {
		permission.Description = *req.Description
	}
	if req.ResourceID != nil {
		permission.ResourceID = req.ResourceID
	}
	if req.ActionID != nil {
		permission.ActionID = req.ActionID
	}
	if req.IsActive != nil {
		permission.IsActive = *req.IsActive
	}

	// Update permission through service
	if err := h.permissionService.UpdatePermission(c.Request.Context(), permission); err != nil {
		h.logger.Error("Failed to update permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to update permission", err)
		return
	}

	// Get updated permission to ensure we have the latest data
	updatedPermission, err := h.permissionService.GetPermissionByID(c.Request.Context(), permissionID)
	if err != nil {
		h.logger.Error("Failed to get updated permission", zap.Error(err), zap.String("permissionID", permissionID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get updated permission", err)
		return
	}

	// Convert to response
	response := respPermissions.NewPermissionResponse(updatedPermission)

	h.logger.Info("Permission updated successfully", zap.String("permissionID", permissionID))
	h.responder.SendSuccess(c, http.StatusOK, response)
}
