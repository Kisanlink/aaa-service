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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ConnPermission struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Action        string  `json:"action"`
	Resource      string    `json:"resource"`
	Source        string    `json:"source"`
	ValidStartTime time.Time `json:"valid_start_time"`
	ValidEndTime   time.Time `json:"valid_end_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
				CreatedAt:   rp.Role.CreatedAt,
				UpdatedAt:   rp.Role.UpdatedAt,
			},
			Permissions: []*ConnPermission{{
				ID:            rp.Permission.ID,
				Name:          rp.Permission.Name,
				Description:   rp.Permission.Description,
				Action:        rp.Permission.Action,
				Resource:      rp.Permission.Resource,
				Source:        rp.Permission.Source,
				ValidStartTime: rp.Permission.ValidStartTime,
				ValidEndTime:   rp.Permission.ValidEndTime,
				CreatedAt:     rp.Permission.CreatedAt,
				UpdatedAt:     rp.Permission.UpdatedAt,
			}},
			IsActive: rp.IsActive,
		})
	}

	response := GetAllRolePermissionsResponse{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Role with Permissions fetched successfully",
		Data:       responseData,
	}

	c.JSON(http.StatusOK, response)
}

func IsZeroValued[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}