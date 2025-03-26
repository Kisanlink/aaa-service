package permissions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)



type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetAllPermissionsResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message     string       `json:"message"`
	Permissions []Permission `json:"permissions"`
}

func (s *PermissionServer) GetAllPermissionsRestApi(c *gin.Context) {
	permissions, err := s.PermissionRepo.FindAllPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve permissions",
		})
		return
	}

	var responsePermissions []Permission
	for _, permission := range permissions {
		responsePermissions = append(responsePermissions, Permission{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
		})
	}

	response := GetAllPermissionsResponse{
		StatusCode:  http.StatusOK,
		Success: true,
		Message:     "Permissions retrieved successfully",
		Permissions: responsePermissions,
	}

	c.JSON(http.StatusOK, response)
}