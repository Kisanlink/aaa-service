package rolepermission

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type RolePermHandler struct {
	roleService     services.RoleServiceInterface
	permService     services.PermissionServiceInterface
	userService     services.UserServiceInterface
	rolePermService services.RolePermissionServiceInterface
}

func NewRolePermHandler(
	roleService services.RoleServiceInterface,
	permService services.PermissionServiceInterface,
	rolePermService services.RolePermissionServiceInterface,
	userService services.UserServiceInterface,
) *RolePermHandler {
	return &RolePermHandler{
		roleService:     roleService,
		permService:     permService,
		userService:     userService,
		rolePermService: rolePermService,
	}
}

// AssignPermissionRestApi assigns permissions to a role
// @Summary Assign permissions to role
// @Description Assigns one or more permissions to a role and updates all affected user relationships
// @Tags Role Permissions
// @Accept json
// @Produce json
// @Param request body model.AssignPermissionRequest true "Permission assignment request"
// @Success 201 {object} helper.Response{data=model.ConnRolePermissionResponse} "Permissions assigned successfully"
// @Failure 400 {object} helper.Response "Invalid request or missing required fields"
// @Failure 404 {object} helper.Response "Role or permission not found"
// @Failure 409 {object} helper.Response "Permission already assigned to role"
// @Failure 500 {object} helper.Response "Failed to assign permissions or update relationships"
// @Router /assign-permissions [post]
func (s *RolePermHandler) AssignPermissionRestApi(c *gin.Context) {
	var req model.AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	if req.Role == "" || len(req.Permissions) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Both role_name and permission_names are required"})
		return
	}

	role, err := s.roleService.GetRoleByName(req.Role)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Role with name " + req.Role + " not found"})
		return
	}

	permissionIDs := make([]string, 0)
	for _, permissionName := range req.Permissions {
		permission, err := s.permService.FindPermissionByName(permissionName)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Permission with name " + permissionName + " not found"})
			return
		}
		permissionIDs = append(permissionIDs, permission.ID)

		existing, _ := s.rolePermService.GetRolePermissionByNames(req.Role, permissionName)
		if existing != nil {
			helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{"Permission '" + permissionName + "' is already assigned to role '" + req.Role + "'"})
			return
		}
	}

	var rolePermissions []*model.RolePermission
	for _, permissionID := range permissionIDs {
		rolePermission := &model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permissionID,
			IsActive:     true,
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}

	if err := s.rolePermService.CreateRolePermissions(rolePermissions); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to create role-permission connections"})
		return
	}

	roles, permissions, actions, usernames, err := s.userService.FindRoleUsersAndPermissionsByRoleId(role.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"failed to fetch user roles and permissions"})
		return
	}

	for _, username := range usernames {
		_, err := client.DeleteUserRoleRelationship(
			username,
			roles,
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to delete relationships for user %s: %v", username, err)
			continue
		}

		_, err = client.CreateUserRoleRelationship(
			username,
			helper.LowerCaseSlice(roles),
			helper.LowerCaseSlice(permissions),
			helper.LowerCaseSlice(actions),
		)
		if err != nil {
			log.Printf("Failed to create relationships for user %s: %v", username, err)
			continue
		}
	}

	fetchedRolePermissions, err := s.rolePermService.GetRolePermissionsByRoleID(role.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch role-permission connections"})
		return
	}

	if len(fetchedRolePermissions) == 0 {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"No permissions found for this role"})
		return
	}

	var rolePermissionPtrs []*model.RolePermission
	for i := range fetchedRolePermissions {
		rolePermissionPtrs = append(rolePermissionPtrs, &fetchedRolePermissions[i])
	}

	response := &model.ConnRolePermissionResponse{
		ID:        rolePermissionPtrs[0].ID,
		CreatedAt: rolePermissionPtrs[0].CreatedAt.Format(time.RFC3339),
		UpdatedAt: rolePermissionPtrs[0].UpdatedAt.Format(time.RFC3339),
		Role: &model.Role{
			Base: model.Base{
				ID:        role.ID,
				CreatedAt: role.CreatedAt,
				UpdatedAt: role.UpdatedAt,
			},
			Name:        role.Name,
			Description: role.Description,
			Source:      role.Source,
		},
		Permissions: []*model.Permission{},
		IsActive:    rolePermissionPtrs[0].IsActive,
	}

	for _, rp := range rolePermissionPtrs {
		if !helper.IsZeroValued(rp.Permission) && rp.Permission.ID != "" {
			response.Permissions = append(response.Permissions, &model.Permission{
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
			})
		}
	}

	helper.SendSuccessResponse(c.Writer, http.StatusCreated, "Role with Permission created successfully", response)
}
