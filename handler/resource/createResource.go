package resource

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	resourceService services.ResourceServiceInterface
}

func NewResourceHandler(
	resourceService services.ResourceServiceInterface,
) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
	}
}

// CreateResourceRestApi creates a new resource
// @Summary Create resource
// @Description Creates a new resource with the provided details (Format: db_name_table_name - e.g. db_aaa_users or db_farmer_crop_cycle)
// @Tags Resources
// @Accept json
// @Produce json
// @Param request body model.CreateResourceRequest true "Resource creation data"
// @Success 201 {object} helper.Response{data=model.Resource} "Resource created successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request body or resource name format (must be db/name/table_name)"
// @Failure 409 {object} helper.ErrorResponse "Resource already exists"
// @Failure 500 {object} helper.ErrorResponse "Failed to create resource"
// @Router /resources [post]
func (s *ResourceHandler) CreateResourceRestApi(c *gin.Context) {
	var req model.Resource
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{err.Error()})
		return
	}

	if req.Name == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Resource name is required"})
		return
	}

	err := helper.OnlyValidName(req.Name)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{err.Error()})
		return
	}
	// Check if resource already exists
	if err := s.resourceService.CheckIfResourceExists(req.Name); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{err.Error()})
		return
	}

	req.Name = helper.SanitizeDBName(req.Name)
	// Create new resource
	if err := s.resourceService.CreateResource(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Resource created successfully", req)
}
