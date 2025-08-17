package routes

import (
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/Kisanlink/aaa-service/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(publicAPI, protectedAPI *gin.RouterGroup, authService *services.AuthService, logger *zap.Logger) {
	// Public auth routes
	auth := publicAPI.Group("/auth")
	{
		auth.POST("/login", createLoginHandler(authService, logger))
		auth.POST("/login-username", createUsernameLoginHandler(authService, logger))
		auth.POST("/register", createRegisterHandler(authService, logger))
		auth.POST("/refresh", createRefreshHandler(authService, logger))
	}

	// Protected auth routes
	protectedAuth := protectedAPI.Group("/auth")
	{
		protectedAuth.POST("/logout", createLogoutHandler(authService, logger))
	}
}

// LoginV2 handles POST /v2/auth/login
// @Summary User login (V2)
// @Description Authenticate user with username and password and MFA
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body services.LoginRequest true "Login credentials"
// @Success 200 {object} responses.LoginSuccessResponse
// @Failure 400 {object} responses.ErrorResponseSwagger
// @Failure 401 {object} responses.ErrorResponseSwagger
// @Failure 500 {object} responses.ErrorResponseSwagger
// @Router /api/v2/auth/login [post]
func createLoginHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.Login(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Login failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// createUsernameLoginHandler creates a handler for username-based login
func createUsernameLoginHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.UsernameLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.LoginWithUsername(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Username login failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// RegisterV2 handles POST /v2/auth/register
// @Summary User registration (V2)
// @Description Register a new user account with enhanced validation
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body services.RegisterRequest true "Registration data"
// @Success 201 {object} responses.RegisterSuccessResponse
// @Failure 400 {object} responses.ErrorResponseSwagger
// @Failure 409 {object} responses.ErrorResponseSwagger
// @Failure 500 {object} responses.ErrorResponseSwagger
// @Router /api/v2/auth/register [post]
func createRegisterHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.Register(c.Request.Context(), &req)
		if err != nil {
			logger.Error("Registration failed", zap.Error(err))

			// Handle different error types properly
			if conflictErr, ok := err.(*errors.ConflictError); ok {
				c.JSON(http.StatusConflict, gin.H{
					"success":   false,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"code":      "CONFLICT",
					"message":   "Registration failed",
					"details":   conflictErr.Error(),
				})
				return
			}

			if validationErr, ok := err.(*errors.ValidationError); ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":   false,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
					"code":      "VALIDATION_ERROR",
					"message":   "Validation failed",
					"errors":    []string{validationErr.Error()},
				})
				return
			}

			// Default error response
			c.JSON(http.StatusBadRequest, gin.H{"error": "registration failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, response)
	}
}

// RefreshTokenV2 handles POST /v2/auth/refresh
// @Summary Refresh access token (V2)
// @Description Refresh access token using refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} responses.RefreshTokenSuccessResponse
// @Failure 400 {object} responses.ErrorResponseSwagger
// @Failure 401 {object} responses.ErrorResponseSwagger
// @Failure 500 {object} responses.ErrorResponseSwagger
// @Router /api/v2/auth/refresh [post]
func createRefreshHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
			MPin         string `json:"mpin" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
			return
		}

		response, err := authService.RefreshToken(c.Request.Context(), req.RefreshToken, req.MPin)
		if err != nil {
			logger.Error("Token refresh failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token refresh failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// LogoutV2 handles POST /v2/auth/logout
// @Summary User logout (V2)
// @Description Logout user and invalidate tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.LogoutSuccessResponse
// @Failure 401 {object} responses.ErrorResponseSwagger
// @Failure 500 {object} responses.ErrorResponseSwagger
// @Router /api/v2/auth/logout [post]
func createLogoutHandler(authService *services.AuthService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "message": "user context not found"})
			return
		}

		err := authService.Logout(c.Request.Context(), userID.(string))
		if err != nil {
			logger.Error("Logout failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
	}
}
