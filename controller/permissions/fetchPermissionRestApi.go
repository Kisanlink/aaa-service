package permissions

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)



type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source string `json:"source"`
	Action string `json:"action"`
	Resource string `json:"resource"`
	ValidStartTime string `json:"valid_start_time"`
	ValidEndTime string `json:"valid_end_time"`
}

type GetAllPermissionsResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message     string       `json:"message"`
	DataTimeStamp string             `json:"data_time_stamp"`
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
			Source: permission.Source,
			Action: permission.Action,
			Resource: permission.Resource,
			ValidStartTime: permission.ValidStartTime.Format(time.RFC3339Nano),
			ValidEndTime: permission.ValidEndTime.Format(time.RFC3339Nano),
		})
	}

	response := GetAllPermissionsResponse{
		StatusCode:  http.StatusOK,
		Success: true,
		Message:     "Permissions retrieved successfully",
		Permissions: responsePermissions,
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format

	}

	c.JSON(http.StatusOK, response)
}