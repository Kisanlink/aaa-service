package resource

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdateResourceRestApi updates an existing resource
// @Summary Update a resource
// @Description Updates an existing resource with the provided details
// @Tags Resources
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Param request body model.CreateResourceRequest true "Resource update data"
// @Success 200 {object} helper.Response{data=model.Resource} "Resource updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request body"
// @Failure 404 {object} helper.ErrorResponse "Resource not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to update resource"
// @Router /resources/{id} [put]
func (s *ResourceHandler) UpdateResourceRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Resource ID is required"})
		return
	}

	var req model.Resource
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{err.Error()})
		return
	}

	err := helper.OnlyValidName(req.Name)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{fmt.Sprintf("Invalid: '%s' - %v", req.Name, err)})
		return
	}
	// Update only allowed fields
	updateData := model.Resource{
		Name: helper.SanitizeDBName(req.Name), // Only update name field
	}

	if err := s.resourceService.UpdateResource(id, updateData); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Fetch updated resource
	updatedResource, err := s.resourceService.FindResourceByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Resource updated successfully", updatedResource)
}
