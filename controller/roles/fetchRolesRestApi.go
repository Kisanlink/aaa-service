package roles

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


type GetAllRolesResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string         `json:"message"`
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
		})
	}

	response := GetAllRolesResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "Roles retrieved successfully",
		Roles:      responseRoles,
	}

	c.JSON(http.StatusOK, response)
}