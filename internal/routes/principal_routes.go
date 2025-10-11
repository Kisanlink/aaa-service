package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/principals"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterPrincipalRoutes registers all principal and service-related routes
func RegisterPrincipalRoutes(router *gin.Engine, principalHandler *principals.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Principal management routes
	principalRoutes := router.Group("/api/v1/principals")
	{
		// Public routes (if any)
		principalRoutes.GET("/:id", principalHandler.GetPrincipal)
		principalRoutes.GET("", principalHandler.ListPrincipals)

		// Protected routes requiring authentication
		authenticated := principalRoutes.Group("")
		authenticated.Use(authMiddleware.HTTPAuthMiddleware())
		{
			authenticated.POST("", principalHandler.CreatePrincipal)
			authenticated.PUT("/:id", principalHandler.UpdatePrincipal)
			authenticated.DELETE("/:id", principalHandler.DeletePrincipal)
		}
	}

	// Service management routes
	serviceRoutes := router.Group("/api/v1/services")
	{
		// Public routes (if any)
		serviceRoutes.GET("/:id", principalHandler.GetService)
		serviceRoutes.GET("", principalHandler.ListServices)

		// Protected routes requiring authentication
		authenticated := serviceRoutes.Group("")
		authenticated.Use(authMiddleware.HTTPAuthMiddleware())
		{
			authenticated.POST("", principalHandler.CreateService)
			authenticated.DELETE("/:id", principalHandler.DeleteService)
			authenticated.POST("/generate-api-key", principalHandler.GenerateAPIKey)
		}
	}
}
