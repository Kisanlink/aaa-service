package permissions

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetAllPermissionsRestApi retrieves all permissions
// @Summary Get all permissions
// @Description Retrieves a list of all permissions in the system with their details
// @Tags Permissions
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=[]model.Permission} "Permissions retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve permissions"
// @Router /permissions [get]
func (s *PermissionHandler) GetAllPermissionsRestApi(c *gin.Context) {
	permissions, err := s.permService.FindAllPermissions()
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to retrieve permissions"})
		return
	}

	var responsePermissions []model.Permission
	for _, permission := range permissions {
		responsePermissions = append(responsePermissions, model.Permission{
			Base: model.Base{
				ID:        permission.ID,
				CreatedAt: permission.CreatedAt,
				UpdatedAt: permission.UpdatedAt,
			},
			Name:           permission.Name,
			Description:    permission.Description,
			Source:         permission.Source,
			Action:         permission.Action,
			Resource:       permission.Resource,
			ValidStartTime: permission.ValidStartTime,
			ValidEndTime:   permission.ValidEndTime,
		})
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Permissions retrieved successfully", responsePermissions)
}
