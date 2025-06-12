package rolepermission

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetAllRolesWithPermissionsRestApi retrieves roles with their permissions
// @Summary Get roles with permissions
// @Description Retrieves roles with their associated permissions, with optional filtering and pagination
// @Tags RolePermissions
// @Accept json
// @Produce json
// @Param role_id query string false "Filter by role ID"
// @Param role_name query string false "Filter by role name (case-insensitive partial match)"
// @Param permission_id query string false "Filter by permission ID"
// @Param page query int false "Page number (starts from 1)"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} helper.Response{data=[]model.GetRolePermissionResponse} "List of roles with permissions retrieved successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} helper.ErrorResponse "Failed to retrieve roles with permissions"
// @Router /role-permissions [get]
// GetAllRolesWithPermissionsRestApi retrieves roles with their permissions
func (s *RolePermissionHandler) GetAllRolesWithPermissionsRestApi(c *gin.Context) {
	filter := make(map[string]interface{})

	// Get filter parameters
	if roleID := c.Query("role_id"); roleID != "" {
		filter["role_id"] = roleID
	}
	if permissionID := c.Query("permission_id"); permissionID != "" {
		filter["permission_id"] = permissionID
	}
	if roleName := c.Query("role_name"); roleName != "" {
		role, err := s.roleService.GetRoleByName(roleName)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
				[]string{fmt.Sprintf("failed to fetch role details: %v", err)})
			return
		}
		filter["role_id"] = "%" + role.ID + "%"
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	// Call service to get role permissions with associations
	rolePermissions, err := s.rolePermissionService.FetchAll(filter, page, limit)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	// Transform to response format
	var response []model.GetRolePermissionResponse
	for _, rp := range rolePermissions {
		// Fetch role details if not already loaded
		var roleResponse *model.Role
		if rp.RoleID != "" {
			role, err := s.roleService.FindRoleByID(rp.RoleID)
			if err != nil {
				helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
					[]string{fmt.Sprintf("failed to fetch role details: %v", err)})
				return
			}

			// Create a minimal role response without permissions
			roleResponse = &model.Role{
				Base: model.Base{
					ID:        role.ID,
					CreatedAt: role.CreatedAt,
					UpdatedAt: role.UpdatedAt,
				},
				Name:        role.Name,
				Description: role.Description,
			}
		}
		// Fetch permission details if not already loaded
		var permission *model.Permission
		if rp.PermissionID != "" {
			permission, err = s.permissionService.GetPermissionByID(rp.PermissionID)
			if err != nil {
				helper.SendErrorResponse(c.Writer, http.StatusInternalServerError,
					[]string{fmt.Sprintf("failed to fetch permission details: %v", err)})
				return
			}
		}

		response = append(response, model.GetRolePermissionResponse{
			ID:         rp.ID,
			CreatedAt:  rp.CreatedAt,
			UpdatedAt:  rp.UpdatedAt,
			Role:       roleResponse,
			Permission: permission,
		})
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK,
		"Roles with permissions retrieved successfully", response)
}
