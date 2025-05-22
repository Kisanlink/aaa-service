package role

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// UpdateRoleWithPermissionsRestApi updates a role and its permissions
// @Summary Update a role with permissions
// @Description Updates an existing role and its permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body model.CreateRoleRequest true "Role and permissions data"
// @Success 200 {object} helper.Response{data=model.Role} "Role updated successfully"
// @Failure 400 {object} helper.Response "Invalid request"
// @Failure 404 {object} helper.Response "Role not found"
// @Failure 500 {object} helper.Response "Failed to update role"
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

	// Convert request to role and permissions
	updatedRole := model.Role{
		Name:        req.Name,
		Description: req.Description,
		Source:      req.Source,
	}

	var permissions []model.Permission
	for _, perm := range req.Permissions {
		permissions = append(permissions, model.Permission{
			Resource: perm.Resource,
			Actions:  perm.Actions,
		})
	}

	// Update role with permissions
	if err := h.roleService.UpdateRoleWithPermissions(id, updatedRole, permissions); err != nil {
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
