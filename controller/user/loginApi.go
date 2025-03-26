package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/helper"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRole struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	RolePermissionID string `json:"role_permission_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type UserResponse struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	MobileNumber uint64     `json:"mobile_number"`
	CountryCode  string     `json:"country_code"`
	IsValidated  bool       `json:"is_validated"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
	UserRoles    []UserRole `json:"user_roles"`
}

type LoginResponse struct {
	StatusCode int               `json:"status_code"`
	Success bool               `json:"success"`
	Message    string      `json:"message"`
	User       UserResponse `json:"user"`
}

func (s *Server) LoginRestApi(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	existingUser, err := s.UserRepo.FindUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	userRoles, err := s.UserRepo.FindUserRoles(c.Request.Context(), existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles"})
		return
	}

	// Convert model.UserRole to API UserRole
	apiUserRoles := make([]UserRole, len(userRoles))
	for i, role := range userRoles {
		apiUserRoles[i] = UserRole{
			ID:               role.ID,
			UserID:           role.UserID,
			RolePermissionID: role.RolePermissionID,
			CreatedAt:        role.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:        role.UpdatedAt.Format(time.RFC3339Nano),
		}
	}

	// Set auth headers with tokens
	if err := helper.SetAuthHeadersWithTokensRest(
		c,
		existingUser.ID,
		existingUser.Username,
		existingUser.IsValidated,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set auth headers"})
		return
	}

	response := LoginResponse{
		StatusCode: http.StatusOK,
		Success: true,
		Message:    "Login successful",
		User: UserResponse{
			ID:           existingUser.ID,
			Username:     existingUser.Username,
			MobileNumber: existingUser.MobileNumber,
			CountryCode:  *existingUser.CountryCode,
			IsValidated:  existingUser.IsValidated,
			CreatedAt:    existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    existingUser.UpdatedAt.Format(time.RFC3339Nano),
			UserRoles:    apiUserRoles,
		},
	}

	c.JSON(http.StatusOK, response)
}