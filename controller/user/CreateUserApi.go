package user

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Mobile      string `json:"mobile" binding:"required"`
}

type CreateUserResponse struct {
	StatusCode   int       `json:"statusCode"`
	Message      string    `json:"message"`
	User         User      `json:"user"`
}

func (s *Server) CreateUserRestApi(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	if err := s.UserRepo.CheckIfUserExists(c.Request.Context(), req.Username); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := &model.User{
		Username:    req.Username,
		Password:    string(hashedPassword),
		MobileNumber: &req.Mobile,
		IsValidated: false,
	}

	createdUser, err := s.UserRepo.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	roles, permissions, actions, err := s.UserRepo.FindUserRolesAndPermissions(c.Request.Context(), createdUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user roles and permissions"})
		return
	}

	userRole := &UserRoleResponse{
		Roles:       roles,
		Permissions: permissions,
		Actions:     actions,
	}

	response := CreateUserResponse{
		StatusCode:   http.StatusCreated,
		Message:      "User created successfully",
		User: User{
			ID:          createdUser.ID,
			Username:    createdUser.Username,
			IsValidated: createdUser.IsValidated,
			CreatedAt:   createdUser.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:   createdUser.UpdatedAt.Format(time.RFC3339Nano),
			UserRoles:   []UserRoleResponse{*userRole},
		},
	}

	c.JSON(http.StatusCreated, response)
}