package roles

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetAllRolesRestApi retrieves all roles
// @Summary Get all roles
// @Description Retrieves a list of all roles in the system
// @Tags Roles
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=[]model.Role} "Roles retrieved successfully"
// @Failure 500 {object} helper.Response "Failed to retrieve roles"
// @Router /roles [get]
func (s *RoleHandler) GetAllRolesRestApi(c *gin.Context) {
	roles, err := s.roleService.FindAllRoles()
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to retrieve roles"})
		return
	}

	var responseRoles []model.Role
	for _, role := range roles {
		responseRoles = append(responseRoles, model.Role{
			Base: model.Base{
				CreatedAt: role.CreatedAt,
				UpdatedAt: role.UpdatedAt,
				ID:        role.ID,
			},
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
		})
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Roles retrieved successfully", responseRoles)
}
