package permissions

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetPermissionByIdRestApi retrieves a permission by ID
// @Summary Get permission by ID
// @Description Retrieves a single permission's details by its unique identifier
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} helper.Response{data=model.Permission} "Permission retrieved successfully"
// @Failure 400 {object} helper.Response "ID is required"
// @Failure 404 {object} helper.Response "Permission not found"
// @Router /permissions/{id} [get]
func (s *PermissionHandler) GetPermissionByIdRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"ID is required"})
		return
	}

	permission, err := s.permService.FindPermissionByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Permission not found"})
		return
	}

	Permission := model.Permission{
		Base: model.Base{
			CreatedAt: permission.CreatedAt,
			UpdatedAt: permission.UpdatedAt,
			ID:        permission.ID,
		},
		Name:           permission.Name,
		Description:    permission.Description,
		Source:         permission.Source,
		Action:         permission.Action,
		Resource:       permission.Resource,
		ValidStartTime: permission.ValidStartTime,
		ValidEndTime:   permission.ValidEndTime,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permission retrieved successfully", Permission)
}
