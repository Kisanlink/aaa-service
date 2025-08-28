package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/auth"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAuthV2Routes configures V2 authentication routes using AuthHandler
func SetupAuthV2Routes(
	publicAPI, protectedAPI *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	userService interfaces.UserService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) {
	// Create AuthHandler instance
	authHandler := auth.NewAuthHandler(userService, validator, responder, logger)

	// Public V2 auth routes with authentication rate limiting
	authV2 := publicAPI.Group("/auth")
	authV2.Use(middleware.AuthenticationRateLimit()) // Add rate limiting for auth endpoints
	{
		authV2.POST("/login", authHandler.LoginV2)
		authV2.POST("/register", authHandler.RegisterV2)
		authV2.POST("/refresh", authHandler.RefreshTokenV2)
		authV2.POST("/forgot-password", authHandler.ForgotPasswordV2)
		authV2.POST("/reset-password", authHandler.ResetPasswordV2)
	}

	// Protected V2 auth routes (require authentication) with sensitive operation rate limiting
	protectedAuthV2 := protectedAPI.Group("/auth")
	protectedAuthV2.Use(authMiddleware.HTTPAuthMiddleware())
	protectedAuthV2.Use(middleware.SensitiveOperationRateLimit()) // Add rate limiting for sensitive operations
	{
		protectedAuthV2.POST("/logout", authHandler.LogoutV2)
		protectedAuthV2.POST("/set-mpin", authHandler.SetMPinV2)
		protectedAuthV2.POST("/update-mpin", authHandler.UpdateMPinV2)
	}
}
