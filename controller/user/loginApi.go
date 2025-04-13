package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Source      string `json:"source"`
	Resource    string `json:"resource"`
}

type UserResponse struct {
	ID              string                  `json:"id"`
	Username        string                  `json:"username"`
	Password        string                  `json:"password"`
	IsValidated     bool                    `json:"is_validated"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
	RolePermissions map[string][]Permission `json:"role_permissions"`
}

type LoginResponse struct {
	StatusCode    int          `json:"status_code"`
	Success       bool         `json:"success"`
	Message       string       `json:"message"`
	Data          UserResponse `json:"data"`
	DataTimeStamp string       `json:"data_time_stamp"`
}

func ConvertAndDeduplicateRolePermissions(input map[string][]model.Permission) map[string][]Permission {
	converted := make(map[string][]Permission)

	for role, permissions := range input {
		uniquePerms := make(map[string]Permission)

		// Deduplicate permissions
		for _, perm := range permissions {
			key := perm.Name + "|" + perm.Action + "|" + perm.Resource
			if _, exists := uniquePerms[key]; !exists {
				uniquePerms[key] = Permission{
					Name:        perm.Name,
					Description: perm.Description,
					Action:      perm.Action,
					Source:      perm.Source,
					Resource:    perm.Resource,
				}
			}
		}

		// Convert map to slice
		var permSlice []Permission
		for _, perm := range uniquePerms {
			permSlice = append(permSlice, perm)
		}

		converted[role] = permSlice
	}

	return converted
}

func (s *Server) LoginRestApi(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate inputs
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if !helper.IsValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username '" + req.Username + "' contains invalid characters. Only a-z, A-Z, 0-9, /, _, |, -, =, + are allowed, and spaces are prohibited.",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Find user
	existingUser, err := s.UserRepo.FindUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get role permissions
	rolePermissions, err := s.UserRepo.FindUsageRights(c.Request.Context(), existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user permissions"})
		return
	}

	// Set auth headers
	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set auth headers"})
		return
	}

	// Prepare response
	response := LoginResponse{
		StatusCode:    http.StatusOK,
		Success:       true,
		Message:       "Login successful",
		DataTimeStamp: time.Now().Format(time.RFC3339),
		Data: UserResponse{
			ID:              existingUser.ID,
			Username:        existingUser.Username,
			Password:        "", // Empty for security
			IsValidated:     existingUser.IsValidated,
			CreatedAt:       existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:       existingUser.UpdatedAt.Format(time.RFC3339Nano),
			RolePermissions: ConvertAndDeduplicateRolePermissions(rolePermissions),
		},
	}

	c.JSON(http.StatusOK, response)
}
