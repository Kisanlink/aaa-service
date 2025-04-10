package user

import (
	"log"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/client"
	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
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

type UserRoleResponse struct {
	Roles       []string             `json:"roles"`
	Permissions []PermissionResponse `json:"permissions"`
}

type AssignRolePermission struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	IsValidated bool              `json:"is_validated"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	UsageRight  *UserRoleResponse `json:"usage_right"`
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
	userRole := model.UserRole{
		UserID:   req.UserID,
		RoleID:   role.ID,
		IsActive: true,
	}

	if err := s.UserRepo.CreateUserRoles(ctx, userRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to assign role to user",
			"data":        nil,
		})
		return
	}
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

	roles, permissionsList, err := s.UserRepo.FindUsageRights(ctx, updatedUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": http.StatusInternalServerError,
			"success":     false,
			"message":     "Failed to fetch user usage rights",
			"data":        nil,
		})
		return
	}

	pbPermissions := make([]PermissionResponse, len(permissionsList))
	for i, perm := range permissionsList {
		pbPermissions[i] = PermissionResponse{
			Name:        perm.Name,
			Description: perm.Description,
			Action:      perm.Action,
			Source:      perm.Source,
			Resource:    perm.Resource,
		}
	}

	userRoleResponse := &UserRoleResponse{
		Roles:       roles,
		Permissions: pbPermissions,
	}

	createdAt := ""
	if !updatedUser.CreatedAt.IsZero() {
		createdAt = updatedUser.CreatedAt.Format(time.RFC3339)
	}

	updatedAt := ""
	if !updatedUser.UpdatedAt.IsZero() {
		updatedAt = updatedUser.UpdatedAt.Format(time.RFC3339)
	}

	connUser := &AssignRolePermission{
		ID:          updatedUser.ID,
		Username:    updatedUser.Username,
		Password:    "",
		IsValidated: updatedUser.IsValidated,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		UsageRight:  userRoleResponse,
	}
	response := &AssignRoleResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Role assigned to user successfully",
		Data:          connUser,
		DataTimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
