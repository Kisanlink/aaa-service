package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/auth"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAuthRoutes configures authentication routes using AuthHandler
func SetupAuthRoutes(
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

	// Public auth routes with comprehensive security
	authGroup := publicAPI.Group("/auth")
	authGroup.Use(middleware.AuthenticationRateLimit())               // Rate limiting for auth endpoints
	authGroup.Use(sanitizationMiddleware.SanitizeInput())             // Input sanitization
	authGroup.Use(middleware.ValidateContentType("application/json")) // Content type validation
	authGroup.Use(middleware.ValidateJSONStructure(5, 20))            // JSON structure validation
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	// Protected auth routes (require authentication) with enhanced security
	protectedAuthGroup := protectedAPI.Group("/auth")
	protectedAuthGroup.Use(authMiddleware.HTTPAuthMiddleware())
	protectedAuthGroup.Use(middleware.SensitiveOperationRateLimit())           // Sensitive operation rate limiting
	protectedAuthGroup.Use(sanitizationMiddleware.SanitizeInput())             // Input sanitization
	protectedAuthGroup.Use(middleware.ValidateContentType("application/json")) // Content type validation
	protectedAuthGroup.Use(middleware.ValidateJSONStructure(3, 10))            // Stricter JSON validation for sensitive ops
	{
		protectedAuthGroup.POST("/logout", authHandler.Logout)

		// MPIN operations with additional rate limiting
		mpinGroup := protectedAuthGroup.Group("/")
		mpinGroup.Use(middleware.MPinRateLimit()) // Additional MPIN-specific rate limiting
		{
			mpinGroup.POST("/set-mpin", authHandler.SetMPin)
			mpinGroup.POST("/update-mpin", authHandler.UpdateMPin)
		}
	}
}
