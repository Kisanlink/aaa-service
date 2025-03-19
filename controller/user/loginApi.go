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

type LoginResponse struct {
	StatusCode   int       `json:"statusCode"`
	Message      string    `json:"message"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         User       `json:"user"`
}

type User struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	IsValidated bool       `json:"isValidated"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
	UserRoles   []UserRoleResponse  `json:"userRoles"`
}

type UserRoleResponse struct {
	Roles       []string       `json:"roles"`
	Permissions []string `json:"permissions"`
	Actions     []string     `json:"actions"`
}

func (s *Server) LoginRestApi(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	existingUser, err := s.UserRepo.FindUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	accessToken, err := helper.GenerateAccessToken(existingUser.ID, existingUser.Username, existingUser.IsValidated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := helper.GenerateRefreshToken(existingUser.ID, existingUser.Username, existingUser.IsValidated)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(c.Request.Context(), existingUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles and permissions"})
		return
	}

	userRole := &UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}
	response := LoginResponse{
		StatusCode:   200,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: User{
			ID:          existingUser.ID,
			Username:    existingUser.Username,
			IsValidated: existingUser.IsValidated,
			CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   existingUser.UpdatedAt.Format(time.RFC3339Nano),
			UserRoles:   []UserRoleResponse{*userRole},
		},
	}

	c.JSON(http.StatusOK, response)
}
