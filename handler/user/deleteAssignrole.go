package user

import (
	"log"
	"net/http"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
)

// DeleteAssignRoleRestApi removes a role from a user
// @Summary Remove a role from a user
// @Description Removes a specified role from a user and returns the updated user details with remaining roles and permissions
// @Tags Users
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param role path string true "Role Name"
// @Success 200 {object} helper.Response{data=model.AssignRolePermission} "Role removed successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid user ID or role"
// @Failure 404 {object} helper.ErrorResponse "User or Role not found"
// @Failure 500 {object} helper.ErrorResponse "Internal server error"
// @Router /remove/{role}/by/{userID} [delete]
func (s *UserHandler) DeleteAssignRoleRestApi(c *gin.Context) {
	// Get userID and role from path parameters
	userID := c.Param("userID")
	role := c.Param("role")

	if userID == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"User ID is required"})
		return
	}

	if role == "" {
		helper.SendErrorResponse(c.Writer, http.StatusBadRequest, []string{"Role is required"})
		return
	}

	// Validate user exists
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		helper.SendErrorResponse(c.Writer, http.StatusNotFound, []string{err.Error()})
		return
	}

	// Delete user-role relationship
	if err := s.userService.DeleteUserRoles(userID); err != nil {

		helper.SendErrorResponse(c.Writer, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	// Update relationships in the authorization service
	err = client.DeleteRelationships(
		[]string{role}, // Delete the specific role being removed
		user.Username,
		user.ID,
	)
	if err != nil {
		log.Printf("Error deleting relationships: %v", err)
	}

	// Build response data
	responseData := &model.AssignRolePermission{
		ID:          user.ID,
		Username:    user.Username,
		IsValidated: user.IsValidated,
	}

	helper.SendSuccessResponse(c.Writer, http.StatusOK, "Role removed from user successfully", responseData)
}
