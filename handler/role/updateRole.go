package role

import (
	"fmt"
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdateRoleWithPermissionsRestApi updates a role
// @Summary Update a role
// @Description Updates an existing role with identified by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body model.CreateRoleRequest true "Role and permissions data"
// @Success 200 {object} helper.Response{data=model.Role} "Role updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 404 {object} helper.ErrorResponse "Role not found"
// @Failure 500 {object} helper.ErrorResponse "Failed to update role"
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRoleWithPermissionsRestApi(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Role ID is required"})
		return
	}

	var req model.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{err.Error()})
		return
	}

	err := helper.OnlyValidName(req.Name)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{fmt.Sprintf("Invalid: '%s' - %v", req.Name, err)})
		return
	}
	// Convert request to role and permissions
	updatedRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	// Update role with permissions
	if err := h.roleService.UpdateRoleWithPermissions(id, updatedRole); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Get the updated role with permissions
	updated, err := h.roleService.FindRoleByID(id)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role updated successfully", updated)
}
