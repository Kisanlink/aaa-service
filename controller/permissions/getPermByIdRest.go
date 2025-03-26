package permissions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


type GetPermissionByIdResponse struct {
	StatusCode  int               `json:"status_code"`
	Success     bool              `json:"success"`
	Message     string            `json:"message"`
	Permission  PermissionResponse `json:"permission"`
}

func (s *PermissionServer) GetPermissionByIdRestApi(c *gin.Context) {
	// Get ID from path parameter
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID is required",
			"success": false,
		})
		return
	}

	permission, err := s.PermissionRepo.FindPermissionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Permission not found",
			"success": false,
		})
		return
	}

	response := GetPermissionByIdResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Permission retrieved successfully",
		Permission: PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
		},
	}

	c.JSON(http.StatusOK, response)
}