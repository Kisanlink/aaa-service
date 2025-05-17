package rolepermission

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// GetAllRolePermissionsRestApi retrieves all role-permission assignments
// @Summary Get all role-permission assignments
// @Description Retrieves all role-permission relationships in the system with full details
// @Tags Role Permissions
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response{data=model.RolePermissionResponse} "Role-permission assignments fetched successfully"
// @Failure 500 {object} helper.Response "Failed to fetch role-permission connections"
// @Router /assign-permissions  [get]
func (s *RolePermHandler) GetAllRolePermissionsRestApi(c *gin.Context) {
	rolePermissions, err := s.rolePermService.GetAllRolePermissions()
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch role-permission connections"})
		return
	}

	var responseData []model.RolePermissionResponse
	for _, rp := range rolePermissions {
		if helper.IsZeroValued(rp.Role) || helper.IsZeroValued(rp.Permission) || rp.Permission.ID == "" {
			continue
		}

		responseData = append(responseData, model.RolePermissionResponse{
			ID:        rp.ID,
			CreatedAt: rp.CreatedAt,
			UpdatedAt: rp.UpdatedAt,
			Role: &model.Role{
				Base: model.Base{
					ID:        rp.Role.ID,
					CreatedAt: rp.Role.CreatedAt,
					UpdatedAt: rp.Role.UpdatedAt,
				},
				Name:        rp.Role.Name,
				Description: rp.Role.Description,
				Source:      rp.Role.Source,
			},
			Permissions: []*model.Permission{{
				Base: model.Base{
					ID:        rp.Permission.ID,
					CreatedAt: rp.Permission.CreatedAt,
					UpdatedAt: rp.Permission.UpdatedAt,
				},
				Name:           rp.Permission.Name,
				Description:    rp.Permission.Description,
				Action:         rp.Permission.Action,
				Resource:       rp.Permission.Resource,
				Source:         rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime,
				ValidEndTime:   rp.Permission.ValidEndTime,
			}},
			IsActive: rp.IsActive,
		})
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role with Permissions fetched successfully", responseData)
}
