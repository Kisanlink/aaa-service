package user

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// DeleteAssignRole delete a role to a user
// @Summary delete a role to a user
// @Description API to delete a specific role assigned to a user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body model.AssignRoleRequest true "Assign Role Request"
// @Success 200 {object} helper.Response{data=nil} "delete assigned Role successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request body"
// @Failure 404 {object} helper.ErrorResponse "User or Role not found"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /assign-role [delete]
func (s *UserHandler) DeleteAssignRoleRestApi(c *gin.Context) {
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

	if err := s.userService.DeleteUserRoles(req.UserID, role.ID); err != nil {

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

	userRoles, err := s.userService.FindUserRoles(user.ID)
	if err != nil {

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})
		return
	}

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

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role assigned to user successfully", nil)
}
