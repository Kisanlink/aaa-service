package routes

import (
	resourceHandlers "github.com/Kisanlink/aaa-service/v2/internal/handlers/resources"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterResourceRoutes registers all resource-related routes with authentication
func RegisterResourceRoutes(
	router *gin.Engine,
	resourceHandler *resourceHandlers.ResourceHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Create a protected route group under /api/v1/resources
	v2 := router.Group("/api/v1")
	v2.Use(authMiddleware.HTTPAuthMiddleware())

	resources := v2.Group("/resources")
	{
		// Resource CRUD operations
		resources.POST("", authMiddleware.RequirePermission("resource", "create"), resourceHandler.CreateResource)
		resources.GET("", authMiddleware.RequirePermission("resource", "read"), resourceHandler.ListResources)
		resources.GET("/:id", authMiddleware.RequirePermission("resource", "read"), resourceHandler.GetResource)
		resources.PUT("/:id", authMiddleware.RequirePermission("resource", "update"), resourceHandler.UpdateResource)
		resources.DELETE("/:id", authMiddleware.RequirePermission("resource", "delete"), resourceHandler.DeleteResource)

		// Resource hierarchy operations
		resources.GET("/:id/children", authMiddleware.RequirePermission("resource", "read"), resourceHandler.GetChildren)
		resources.GET("/:id/hierarchy", authMiddleware.RequirePermission("resource", "read"), resourceHandler.GetHierarchy)
	}
}
