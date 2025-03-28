package roles

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)


type GetAllRolesResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string         `json:"message"`
	DataTimeStamp string             `json:"data_time_stamp"`
	Roles      []RoleResponse `json:"roles"`
}

func (s *RoleServer) GetAllRolesRestApi(c *gin.Context) {
	roles, err := s.RoleRepo.FindAllRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve roles",
		})
		return
	}

	var responseRoles []RoleResponse
	for _, role := range roles {
		responseRoles = append(responseRoles, RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source: role.Source,
		})
	}

	response := GetAllRolesResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "Roles retrieved successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		Roles:      responseRoles,
	}

	c.JSON(http.StatusOK, response)
}