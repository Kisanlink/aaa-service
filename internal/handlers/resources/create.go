package resources

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/internal/entities/models"
	reqResources "github.com/Kisanlink/aaa-service/internal/entities/requests/resources"
	respResources "github.com/Kisanlink/aaa-service/internal/entities/responses/resources"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateResource handles POST /api/v2/resources
//
//	@Summary		Create a new resource
//	@Description	Create a new resource with name, type, and description
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			resource	body		reqResources.CreateResourceRequest	true	"Resource creation data"
//	@Success		201			{object}	respResources.ResourceResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v2/resources [post]
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	h.logger.Info("Creating resource")

	var req reqResources.CreateResourceRequest
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

	// Create resource through service
	var resource *models.Resource
	var err error

	if req.ParentID != nil && req.OwnerID != nil {
		// Create base resource first, then set parent and owner
		resource, err = h.resourceService.Create(c.Request.Context(), req.Name, req.Type, req.Description)
		if err == nil && req.ParentID != nil {
			err = h.resourceService.SetParent(c.Request.Context(), resource.ID, *req.ParentID)
		}
		if err == nil && req.OwnerID != nil {
			err = h.resourceService.SetOwner(c.Request.Context(), resource.ID, *req.OwnerID)
		}
		// Reload the resource to get updated values
		if err == nil {
			resource, err = h.resourceService.GetByID(c.Request.Context(), resource.ID)
		}
	} else if req.ParentID != nil {
		resource, err = h.resourceService.CreateWithParent(c.Request.Context(), req.Name, req.Type, req.Description, *req.ParentID)
	} else if req.OwnerID != nil {
		resource, err = h.resourceService.CreateWithOwner(c.Request.Context(), req.Name, req.Type, req.Description, *req.OwnerID)
	} else {
		resource, err = h.resourceService.Create(c.Request.Context(), req.Name, req.Type, req.Description)
	}

	if err != nil {
		h.logger.Error("Failed to create resource", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to create resource", err)
		return
	}

	// Convert to response
	response := respResources.NewResourceResponse(resource)

	h.logger.Info("Resource created successfully", zap.String("resourceID", response.ID))
	h.responder.SendSuccess(c, http.StatusCreated, response)
}
