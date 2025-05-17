package roles

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetRoleByIdRestApi retrieves a role by ID
// @Summary Get role by ID
// @Description Retrieves a single role's details by its unique identifier
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} helper.Response{data=model.Role} "Role retrieved successfully"
// @Failure 400 {object} helper.Response "ID is required"
// @Failure 404 {object} helper.Response "Role not found"
// @Router /roles/{id} [get]
func (s *RoleHandler) GetRoleByIdRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"ID is required"})
		return
	}

	role, err := s.roleService.FindRoleByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Role not found"})
		return
	}

	roleResponse := model.Role{
		Base: model.Base{
			CreatedAt: role.CreatedAt,
			UpdatedAt: role.UpdatedAt,
			ID:        role.ID,
		},
		Name:        role.Name,
		Description: role.Description,
		Source:      role.Source,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role retrieved successfully", roleResponse)
}
