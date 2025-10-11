package resources

import (
	"net/http"

	reqResources "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/resources"
	respResources "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/resources"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateResource handles PUT /api/v2/resources/:id
//
//	@Summary		Update resource
//	@Description	Update an existing resource
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string								true	"Resource ID"
//	@Param			resource	body		reqResources.UpdateResourceRequest	true	"Resource update data"
//	@Success		200			{object}	respResources.ResourceResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v2/resources/{id} [put]
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	resourceID := c.Param("id")
	h.logger.Info("Updating resource", zap.String("resourceID", resourceID))

	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	var req reqResources.UpdateResourceRequest
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

	// Get existing resource
	resource, err := h.resourceService.GetByID(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusNotFound, "Resource not found", err)
		return
	}

	// Apply updates
	if req.Name != nil {
		resource.Name = *req.Name
	}
	if req.Type != nil {
		resource.Type = *req.Type
	}
	if req.Description != nil {
		resource.Description = *req.Description
	}
	if req.IsActive != nil {
		resource.IsActive = *req.IsActive
	}
	if req.ParentID != nil {
		resource.ParentID = req.ParentID
	}
	if req.OwnerID != nil {
		resource.OwnerID = req.OwnerID
	}

	// Update resource through service
	if err := h.resourceService.Update(c.Request.Context(), resource); err != nil {
		h.logger.Error("Failed to update resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to update resource", err)
		return
	}

	// Get updated resource to ensure we have the latest data
	updatedResource, err := h.resourceService.GetByID(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get updated resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get updated resource", err)
		return
	}

	// Convert to response
	response := respResources.NewResourceResponse(updatedResource)

	h.logger.Info("Resource updated successfully", zap.String("resourceID", resourceID))
	h.responder.SendSuccess(c, http.StatusOK, response)
}
