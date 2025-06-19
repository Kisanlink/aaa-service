package user

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserServiceInterface
	RoleService services.RoleServiceInterface
}

func NewUserHandler(
	userService services.UserServiceInterface,
	RoleService services.RoleServiceInterface,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		RoleService: RoleService,
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
// @Failure 400 {object} helper.ErrorResponse "Invalid request body"
// @Failure 404 {object} helper.ErrorResponse "User or Role not found"
// @Failure 409 {object} helper.ErrorResponse "Role already assigned to user"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /assign-role [post]
func (s *UserHandler) AssignRoleRestApi(c *gin.Context) {
	var req model.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Invalid request body"})
		return
	}

	// Validate user exists
	user, err := s.userService.GetUserByID(req.UserID)
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

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	// Get updated user details
	updatedUser, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	userRoles, err := s.userService.FindUserRoles(user.ID)
	if err != nil {

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	helper.PrettyJSON(userRoles)
	roleNames := make([]string, 0, len(userRoles))
	for _, userRole := range userRoles {
		role, err := s.RoleService.FindRoleByID(userRole.RoleID)
		if err != nil {
			helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
		}
		roleNames = append(roleNames, role.Name)
	}

	err = client.DeleteRelationships(
		roleNames,
		user.Username,
		user.ID,
	)

	if err != nil {
		log.Printf("Error deleting relationships: %v", err)
	}

	err = client.CreateRelationships(
		roleNames,
		user.Username,
		user.ID,
	)

	if err != nil {
		log.Printf("Error creating relationships: %v", err)
	}

	rolesResponse, err := s.userService.GetUserRolesWithPermissions(req.UserID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}
	// // Build response data
	responseData := &model.AssignRolePermission{
		ID:          updatedUser.ID,
		Username:    updatedUser.Username,
		IsValidated: updatedUser.IsValidated,
		Roles:       rolesResponse.Roles,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role assigned to user successfully", responseData)
}
