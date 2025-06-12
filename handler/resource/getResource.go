package resource

import (
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
)

// @Summary Get resources
// @Description Get resources with optional filtering by ID or name and pagination
// @Tags Resources
// @Param id query string false "Filter by resource ID"
// @Param name query string false "Filter by resource name"
// @Param page query int false "Page number (starts from 1)"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} helper.Response{data=[]model.Resource}
// @Router /resources [get]
func (s *ResourceHandler) GetResourcesRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	// Get filter parameters
	if id := c.Query("id"); id != "" {
		filter["id"] = id
	}
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	resources, err := s.resourceService.FindResources(filter, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Resources retrieved successfully", resources)
}
