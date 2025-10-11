package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/auth"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
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

	// Create input sanitization middleware
	sanitizationMiddleware := middleware.NewInputSanitizationMiddleware(logger)

	// Public V2 auth routes with comprehensive security
	authV2 := publicAPI.Group("/auth")
	authV2.Use(middleware.AuthenticationRateLimit())               // Rate limiting for auth endpoints
	authV2.Use(sanitizationMiddleware.SanitizeInput())             // Input sanitization
	authV2.Use(middleware.ValidateContentType("application/json")) // Content type validation
	authV2.Use(middleware.ValidateJSONStructure(5, 20))            // JSON structure validation
	{
		authV2.POST("/login", authHandler.LoginV2)
		authV2.POST("/register", authHandler.RegisterV2)
		authV2.POST("/refresh", authHandler.RefreshTokenV2)
		authV2.POST("/forgot-password", authHandler.ForgotPasswordV2)
		authV2.POST("/reset-password", authHandler.ResetPasswordV2)
	}

	// Protected V2 auth routes (require authentication) with enhanced security
	protectedAuthV2 := protectedAPI.Group("/auth")
	protectedAuthV2.Use(authMiddleware.HTTPAuthMiddleware())
	protectedAuthV2.Use(middleware.SensitiveOperationRateLimit())           // Sensitive operation rate limiting
	protectedAuthV2.Use(sanitizationMiddleware.SanitizeInput())             // Input sanitization
	protectedAuthV2.Use(middleware.ValidateContentType("application/json")) // Content type validation
	protectedAuthV2.Use(middleware.ValidateJSONStructure(3, 10))            // Stricter JSON validation for sensitive ops
	{
		protectedAuthV2.POST("/logout", authHandler.LogoutV2)

		// MPIN operations with additional rate limiting
		mpinGroup := protectedAuthV2.Group("/")
		mpinGroup.Use(middleware.MPinRateLimit()) // Additional MPIN-specific rate limiting
		{
			mpinGroup.POST("/set-mpin", authHandler.SetMPinV2)
			mpinGroup.POST("/update-mpin", authHandler.UpdateMPinV2)
		}
	}
}
