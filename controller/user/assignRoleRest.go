package user

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type PermissionResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Resource    string `json:"resource"`
}

type RoleResponse struct {
	RoleName    string               `json:"role_name"`
	Permissions []PermissionResponse `json:"permissions"`
}

type AssignRolePermission struct {
	ID             string         `json:"id"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	IsValidated    bool           `json:"is_validated"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	RolePermission []RoleResponse `json:"role_permissions"`
}

type AssignRoleResponse struct {
	StatusCode    int                   `json:"status_code"`
	Success       bool                  `json:"success"`
	Message       string                `json:"message"`
	Data          *AssignRolePermission `json:"data"`
	DataTimeStamp string                `json:"data_time_stamp"`
}

func (s *Server) AssignRoleRestApi(c *gin.Context) {
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

	// Create user-role relationship
	userRole := model.UserRole{
		UserID:   req.UserID,
		RoleID:   role.ID,
		IsActive: true,
	}
	if err := s.UserRepo.CreateUserRoles(ctx, userRole); err != nil {
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
			"message":     "Failed to assign role to user",
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

	// Get role permissions in the correct format
	rolePermissions, err := s.UserRepo.FindUsageRights(ctx, updatedUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to fetch user usage rights",
			"data":        nil,
		})
		return
	}

	// Convert role permissions to the new structure
	var rolesResponse []RoleResponse
	for roleName, permissions := range rolePermissions {
		uniquePerms := make(map[string]PermissionResponse)
		for _, perm := range permissions {
			key := perm.Name + ":" + perm.Action + ":" + perm.Resource
			uniquePerms[key] = PermissionResponse{
				Name:        perm.Name,
				Description: perm.Description,
				Action:      perm.Action,
				Source:      perm.Source,
				Resource:    perm.Resource,
			}
		}

		// Convert unique permissions map to slice
		var permsSlice []PermissionResponse
		for _, perm := range uniquePerms {
			permsSlice = append(permsSlice, perm)
		}

		rolesResponse = append(rolesResponse, RoleResponse{
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

	// Build response
	response := &AssignRoleResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role assigned to user successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: &AssignRolePermission{
			ID:             updatedUser.ID,
			Username:       updatedUser.Username,
			Password:       "",
			IsValidated:    updatedUser.IsValidated,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			RolePermission: rolesResponse,
		},
	}
	c.JSON(http.StatusOK, response)
}
