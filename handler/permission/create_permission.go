package permission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService services.PermissionServiceInterface
}

func NewPermissionHandler(
	permissionService services.PermissionServiceInterface,
) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// CreatePermissionRestApi creates a new permission
// @Summary Create a new permission
// @Description Creates a new permission with the provided details
// @Tags Permissions
// @Accept json
// @Produce json
// @Param request body model.CreatePermissionRequest true "Permission creation data"
// @Success 201 {object} helper.Response{data=model.Permission} "Permission created successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request or missing required fields"
// @Failure 409 {object} helper.ErrorResponse "Permission already exists for this role+resource"
// @Failure 500 {object} helper.ErrorResponse "Failed to create permission"
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermissionRestApi(c *gin.Context) {
	var req model.Permission
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Resource == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Resource is required"})
		return
	}

	if len(req.Actions) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"At least one action is required"})
		return
	}

	if err := h.permissionService.CreatePermission(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Permission created successfully", req)
}
