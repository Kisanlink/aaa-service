package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/organizations"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupOrganizationRoutes configures organization-related routes
func SetupOrganizationRoutes(router *gin.Engine, orgHandler *organizations.Handler, authMiddleware *middleware.AuthMiddleware) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Organization routes with authentication
		orgV1 := v1.Group("/organizations")
		orgV1.Use(authMiddleware.HTTPAuthMiddleware())
		{
			orgV1.POST("/", orgHandler.CreateOrganization)
			orgV1.GET("/", orgHandler.ListOrganizations)
			orgV1.GET("/:id", orgHandler.GetOrganization)
			orgV1.PUT("/:id", orgHandler.UpdateOrganization)
			orgV1.DELETE("/:id", orgHandler.DeleteOrganization)
			orgV1.GET("/:id/hierarchy", orgHandler.GetOrganizationHierarchy)
			orgV1.POST("/:id/activate", orgHandler.ActivateOrganization)
			orgV1.POST("/:id/deactivate", orgHandler.DeactivateOrganization)
			orgV1.GET("/:id/stats", orgHandler.GetOrganizationStats)
		}
	}

	// API v2 routes (if needed for future enhancements)
	v2 := router.Group("/api/v2")
	{
		// Organization routes with authentication
		orgV2 := v2.Group("/organizations")
		orgV2.Use(authMiddleware.HTTPAuthMiddleware())
		{
			orgV2.POST("/", orgHandler.CreateOrganization)
			orgV2.GET("/", orgHandler.ListOrganizations)
			orgV2.GET("/:id", orgHandler.GetOrganization)
			orgV2.PUT("/:id", orgHandler.UpdateOrganization)
			orgV2.DELETE("/:id", orgHandler.DeleteOrganization)
			orgV2.GET("/:id/hierarchy", orgHandler.GetOrganizationHierarchy)
			orgV2.POST("/:id/activate", orgHandler.ActivateOrganization)
			orgV2.POST("/:id/deactivate", orgHandler.DeactivateOrganization)
			orgV2.GET("/:id/stats", orgHandler.GetOrganizationStats)
		}
	}
}
