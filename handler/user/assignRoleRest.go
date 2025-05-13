package user

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userService     services.UserServiceInterface
	RoleService     services.RoleServiceInterface
	PermService     services.PermissionServiceInterface
	RolePermService services.RolePermissionServiceInterface
}

func NewUserHandler(
	userService services.UserServiceInterface,
	RoleService services.RoleServiceInterface,
	PermService services.PermissionServiceInterface,
	RolePermService services.RolePermissionServiceInterface) *UserHandler {
	return &UserHandler{
		userService:     userService,
		RoleService:     RoleService,
		PermService:     PermService,
		RolePermService: RolePermService,
	}
}

// AssignRoleRestApi assigns a role to a user
// @Summary Assign a role to a user
// @Description Assigns a specified role to a user and returns the updated user details with roles and permissions
// @Tags Users
// @Accept json
// @Produce json
// @Param request body model.AssignRoleRequest true "Assign Role Request"
// @Success 200 {object} helper.Response{data=model.AssignRolePermission} "Role assigned successfully"
// @Failure 400 {object} helper.Response "Invalid request body"
// @Failure 404 {object} helper.Response "User or Role not found"
// @Failure 409 {object} helper.Response "Role already assigned to user"
// @Failure 500 {object} helper.Response "Internal server error"
// @Router /assign-role [post]
func (s *UserHandler) AssignRoleRestApi(c *gin.Context) {
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	// Validate user exists
	_, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"User not found"})
		return
	}

	// Validate role exists
	role, err := s.RoleService.GetRoleByName(req.Role)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{"Role not found"})
		return
	}

	// Create user-role relationship
	userRole := model.UserRole{
		UserID:   req.UserID,
		RoleID:   role.ID,
		IsActive: true,
	}
	if err := s.userService.CreateUserRoles(userRole); err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				helper.SendErrorResponse(c.Writer, http.StatusConflict, []string{st.Message()})
			default:
				helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{st.Message()})
			}
			return
		}
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to assign role to user"})
		return
	}

	// Get updated user details
	updatedUser, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user details"})
		return
	}

	// Get roles and permissions for relationship updates
	roles, permissions, actions, err := s.userService.FindUserRolesAndPermissions(updatedUser.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user roles and permissions"})
		return
	}

	// Update relationships in external service
	_, err = client.DeleteUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		log.Printf("Failed to delete relationships: %v", err)
	}
	_, err = client.CreateUserRoleRelationship(
		updatedUser.Username,
		helper.LowerCaseSlice(roles),
		helper.LowerCaseSlice(permissions),
		helper.LowerCaseSlice(actions),
	)
	if err != nil {
		log.Printf("Failed to create relationships: %v", err)
	}

	// Get role permissions in the correct format
	rolePermissions, err := s.userService.FindUsageRights(updatedUser.ID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{"Failed to fetch user usage rights"})
		return
	}

	// Convert role permissions to the new structure
	var rolesResponse []model.RoleResp
	for roleName, permissions := range rolePermissions {
		uniquePerms := make(map[string]model.Permission)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			uniquePerms[key] = model.Permission{
				Base: model.Base{
					ID:        perm.ID,
					CreatedAt: perm.CreatedAt,
					UpdatedAt: perm.UpdatedAt,
				},
				Name:        perm.Name,
				Description: perm.Description,
				Action:      perm.Action,
				Source:      perm.Source,
				Resource:    perm.Resource,
			}
		}

		// Convert unique permissions map to slice
		var permsSlice []model.Permission
		for _, perm := range uniquePerms {
			permsSlice = append(permsSlice, perm)
		}

		rolesResponse = append(rolesResponse, model.RoleResp{
			RoleName:    roleName,
			Permissions: permsSlice,
		})
	}

	// Format timestamps
	createdAt := ""
	if !updatedUser.CreatedAt.IsZero() {
		createdAt = updatedUser.CreatedAt.Format(time.RFC3339Nano)
	}
	updatedAt := ""
	if !updatedUser.UpdatedAt.IsZero() {
		updatedAt = updatedUser.UpdatedAt.Format(time.RFC3339Nano)
	}

	// Build response data
	responseData := &model.AssignRolePermission{
		ID:             updatedUser.ID,
		Username:       updatedUser.Username,
		IsValidated:    updatedUser.IsValidated,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		RolePermission: rolesResponse,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role assigned to user successfully", responseData)
}
