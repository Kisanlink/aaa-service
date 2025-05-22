package resource

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// @Summary Get resources
// @Description Get resources with optional filtering by ID or name
// @Tags Resources
// @Param id query string false "Filter by resource ID"
// @Param name query string false "Filter by resource name"
// @Success 200 {object} helper.Response{data=[]model.Resource}
// @Router /resources [get]
func (s *ResourceHandler) GetResourcesRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	if id := c.Query("id"); id != "" {
		filter["id"] = id
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	resources, err := s.resourceService.FindResources(filter)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Resources retrieved successfully", resources)
}
