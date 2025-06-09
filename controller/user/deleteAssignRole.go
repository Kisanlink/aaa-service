package user

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) DeleteAssignRoleRestApi(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": http.StatusBadRequest,
			"success":     false,
			"message":     "Invalid request body",
			"data":        nil,
		})
		return
	}
	ctx := c.Request.Context()

	// Validate user exists
	_, err := s.UserRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"success":     false,
			"message":     "User not found",
			"data":        nil,
		})
		return
	}

	// Validate role exists
	role, err := s.RoleRepo.GetRoleByName(ctx, req.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": http.StatusNotFound,
			"success":     false,
			"message":     "Role not found",
			"data":        nil,
		})
		return
	}

	if err := s.UserRepo.DeleteUserRole(ctx, req.UserID, role.ID); err != nil {
		// Handle specific error cases
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, gin.H{
					"status_code": http.StatusConflict,
					"success":     false,
					"message":     st.Message(),
					"data":        nil,
				})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"status_code": http.StatusInternalServerError,
					"success":     false,
					"message":     st.Message(),
					"data":        nil,
				})
				return
			}
		}

		// Fallback for non-gRPC errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to delete assigned role to user",
			"data":        nil,
		})
		return
	}

	// Get updated user details
	updatedUser, err := s.UserRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to fetch user details",
			"data":        nil,
		})
		return
	}

	// Get roles and permissions for relationship updates
	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(ctx, updatedUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to fetch user roles and permissions",
			"data":        nil,
		})
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

	// Build response
	response := &AssignRoleResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role Deleted  successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data:          nil,
	}
	c.JSON(http.StatusOK, response)
}
