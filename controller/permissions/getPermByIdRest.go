package permissions

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)


type GetPermissionByIdResponse struct {
	StatusCode  int               `json:"status_code"`
	Success     bool              `json:"success"`
	Message     string            `json:"message"`
	DataTimeStamp string             `json:"data_time_stamp"`
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
		DataTimeStamp: time.Now().Format(time.RFC3339), // Current time in RFC3339 string format
		Permission: PermissionResponse{
			ID:          permission.ID,
			Name:        permission.Name,
			Description: permission.Description,
			Source: permission.Source,
			Action: permission.Action,
			Resource: permission.Resource,
			ValidStartTime: permission.ValidStartTime.Format(time.RFC3339Nano),
			ValidEndTime: permission.ValidEndTime.Format(time.RFC3339Nano),
		},
	}

	c.JSON(http.StatusOK, response)
}