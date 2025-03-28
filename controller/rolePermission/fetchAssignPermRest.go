package rolepermission

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)


type ConnRole struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type ConnPermission struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Action        string  `json:"action"`
	Resource      string    `json:"resource"`
	Source        string    `json:"source"`
	ValidStartTime string `json:"valid_start_time"`
	ValidEndTime   string `json:"valid_end_time"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type RolePermissionResponse struct {
	ID         string          `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Role       *ConnRole        `json:"role"`
	Permissions []*ConnPermission `json:"permission"`
	IsActive   bool            `json:"is_active"`
}

type GetAllRolePermissionsResponse struct {
	StatusCode int                     `json:"status_code"`
	Success    bool                    `json:"success"`
	Message    string                  `json:"message"`
	DataTimeStamp string             `json:"data_time_stamp"`
	Data       []RolePermissionResponse `json:"data"`
}

func (s *ConnectRolePermissionServer) GetAllRolePermissionsRestApi(c *gin.Context) {
	rolePermissions, err := s.RolePermissionRepo.GetAllRolePermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch role-permission connections",
			"success": false,
		})
		return
	}

	var responseData []RolePermissionResponse
	for _, rp := range rolePermissions {
		if IsZeroValued(rp.Role) || IsZeroValued(rp.Permission) || rp.Permission.ID == "" {
			continue
		}

		responseData = append(responseData, RolePermissionResponse{
			ID:        rp.ID,
			CreatedAt: rp.CreatedAt,
			UpdatedAt: rp.UpdatedAt,
			Role: &ConnRole{
				ID:          rp.Role.ID,
				Name:        rp.Role.Name,
				Description: rp.Role.Description,
				Source:      rp.Role.Source,
				CreatedAt:   rp.Role.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:   rp.Role.UpdatedAt.Format(time.RFC3339Nano),
			},
			Permissions: []*ConnPermission{{
				ID:            rp.Permission.ID,
				Name:          rp.Permission.Name,
				Description:   rp.Permission.Description,
				Action:        rp.Permission.Action,
				Resource:      rp.Permission.Resource,
				Source:        rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime.Format(time.RFC3339Nano),
				ValidEndTime:   rp.Permission.ValidEndTime.Format(time.RFC3339Nano),
				CreatedAt:     rp.Permission.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:     rp.Permission.UpdatedAt.Format(time.RFC3339Nano),
			}},
			IsActive: rp.IsActive,
		})
	}

	response := GetAllRolePermissionsResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Role with Permissions fetched successfully",
		DataTimeStamp: time.Now().Format(time.RFC3339Nano),
		Data:       responseData,
	}

	c.JSON(http.StatusOK, response)
}

func IsZeroValued[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}