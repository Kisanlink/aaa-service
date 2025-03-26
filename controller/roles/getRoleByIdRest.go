package roles

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// type RoleResponse struct {
// 	ID          string `json:"id"`
// 	Name        string `json:"name"`
// 	Description string `json:"description"`
// }

type GetRoleByIdResponse struct {
	StatusCode int          `json:"status_code"`
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Role       RoleResponse `json:"role"`
}

func (s *RoleServer) GetRoleByIdRestApi(c *gin.Context) {
	// Get ID from path parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID is required",
			"success": false,
		})
		return
	}

	role, err := s.RoleRepo.FindRoleByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Role not found",
			"success": false,
		})
		return
	}

	response := GetRoleByIdResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Role retrieved successfully",
		Role: RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
			Source: role.Source,
		},
	}

	c.JSON(http.StatusOK, response)
}