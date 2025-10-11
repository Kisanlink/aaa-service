package resources

import (
	"net/http"

	reqResources "github.com/Kisanlink/aaa-service/internal/entities/requests/resources"
	respResources "github.com/Kisanlink/aaa-service/internal/entities/responses/resources"
	"github.com/Kisanlink/aaa-service/internal/services/resources"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetResource handles GET /api/v2/resources/:id
//
//	@Summary		Get resource by ID
//	@Description	Retrieve a resource by its unique identifier
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Resource ID"
//	@Success		200	{object}	respResources.ResourceResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/resources/{id} [get]
func (h *ResourceHandler) GetResource(c *gin.Context) {
	resourceID := c.Param("id")
	h.logger.Info("Getting resource by ID", zap.String("resourceID", resourceID))

	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	// Get resource through service
	resource, err := h.resourceService.GetByID(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusNotFound, "Resource not found", err)
		return
	}

	// Convert to response
	response := respResources.NewResourceResponse(resource)

	h.logger.Info("Resource retrieved successfully", zap.String("resourceID", resourceID))
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// ListResources handles GET /api/v2/resources
//
//	@Summary		List resources
//	@Description	Get a paginated list of resources with optional filters
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			type		query		string	false	"Resource type filter"
//	@Param			parent_id	query		string	false	"Parent ID filter"
//	@Param			owner_id	query		string	false	"Owner ID filter"
//	@Param			is_active	query		boolean	false	"Active status filter"
//	@Param			search		query		string	false	"Search term"
//	@Param			limit		query		int		false	"Number of resources to return"	default(10)
//	@Param			offset		query		int		false	"Number of resources to skip"		default(0)
//	@Success		200			{object}	respResources.ResourceListResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v2/resources [get]
func (h *ResourceHandler) ListResources(c *gin.Context) {
	h.logger.Info("Listing resources")

	// Parse query parameters
	var req reqResources.QueryResourceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query parameters", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Error("Query validation failed", zap.Error(err))
		h.responder.SendValidationError(c, []string{err.Error()})
		return
	}

	// Get resources from service
	resources, err := h.resourceService.List(c.Request.Context(), req.GetLimit(), req.GetOffset())
	if err != nil {
		h.logger.Error("Failed to list resources", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to retrieve resources", err)
		return
	}

	// Get total count
	total, err := h.resourceService.Count(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to count resources", zap.Error(err))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to count resources", err)
		return
	}

	// Calculate page number
	page := (req.GetOffset() / req.GetLimit()) + 1

	// Create response
	response := respResources.NewResourceListResponse(
		resources,
		page,
		req.GetLimit(),
		int(total),
		h.getRequestID(c),
	)

	h.logger.Info("Resources listed successfully",
		zap.Int("count", len(resources)),
		zap.Int64("total", total))

	c.JSON(http.StatusOK, response)
}

// GetChildren handles GET /api/v2/resources/:id/children
//
//	@Summary		Get child resources
//	@Description	Get all direct children of a resource
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Resource ID"
//	@Success		200	{object}	respResources.ResourceWithChildrenResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/resources/{id}/children [get]
func (h *ResourceHandler) GetChildren(c *gin.Context) {
	resourceID := c.Param("id")
	h.logger.Info("Getting child resources", zap.String("parentID", resourceID))

	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	// Get parent resource
	resource, err := h.resourceService.GetByID(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get resource", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusNotFound, "Resource not found", err)
		return
	}

	// Get children
	children, err := h.resourceService.GetChildren(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get children", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get children", err)
		return
	}

	// Create response
	response := respResources.NewResourceWithChildrenResponse(resource, children)

	h.logger.Info("Children retrieved successfully",
		zap.String("resourceID", resourceID),
		zap.Int("count", len(children)))

	h.responder.SendSuccess(c, http.StatusOK, response)
}

// GetHierarchy handles GET /api/v2/resources/:id/hierarchy
//
//	@Summary		Get resource hierarchy
//	@Description	Get the full hierarchical tree starting from a resource
//	@Tags			resources
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Resource ID"
//	@Success		200	{object}	respResources.ResourceHierarchyResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/v2/resources/{id}/hierarchy [get]
func (h *ResourceHandler) GetHierarchy(c *gin.Context) {
	resourceID := c.Param("id")
	h.logger.Info("Getting resource hierarchy", zap.String("resourceID", resourceID))

	if resourceID == "" {
		h.responder.SendValidationError(c, []string{"resource ID is required"})
		return
	}

	// Get hierarchy from service
	tree, err := h.resourceService.GetHierarchy(c.Request.Context(), resourceID)
	if err != nil {
		h.logger.Error("Failed to get hierarchy", zap.Error(err), zap.String("resourceID", resourceID))
		h.responder.SendError(c, http.StatusInternalServerError, "Failed to get hierarchy", err)
		return
	}

	// Convert to response using the proper conversion
	treeResp := &respResources.ResourceTree{
		Resource: tree.Resource,
		Children: convertServiceTreeToResponseTree(tree.Children),
	}
	response := respResources.NewResourceHierarchyResponse(treeResp)

	h.logger.Info("Hierarchy retrieved successfully", zap.String("resourceID", resourceID))
	h.responder.SendSuccess(c, http.StatusOK, response)
}

// Helper function to convert service tree to response tree
func convertServiceTreeToResponseTree(serviceTrees []*resources.ResourceTree) []*respResources.ResourceTree {
	result := make([]*respResources.ResourceTree, 0, len(serviceTrees))
	for _, st := range serviceTrees {
		result = append(result, &respResources.ResourceTree{
			Resource: st.Resource,
			Children: convertServiceTreeToResponseTree(st.Children),
		})
	}
	return result
}
