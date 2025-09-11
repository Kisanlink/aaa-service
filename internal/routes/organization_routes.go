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

			// Organization-scoped group management routes
			orgV1.GET("/:orgId/groups", orgHandler.GetOrganizationGroups)
			orgV1.POST("/:orgId/groups", orgHandler.CreateGroupInOrganization)
			orgV1.GET("/:orgId/groups/:groupId", orgHandler.GetGroupInOrganization)

			// User-group management within organization context
			orgV1.POST("/:orgId/groups/:groupId/users", orgHandler.AddUserToGroupInOrganization)
			orgV1.DELETE("/:orgId/groups/:groupId/users/:userId", orgHandler.RemoveUserFromGroupInOrganization)
			orgV1.GET("/:orgId/groups/:groupId/users", orgHandler.GetGroupUsersInOrganization)
			orgV1.GET("/:orgId/users/:userId/groups", orgHandler.GetUserGroupsInOrganization)
			orgV1.GET("/:orgId/users/:userId/effective-roles", orgHandler.GetUserEffectiveRolesInOrganization)

			// Role-group management within organization context
			orgV1.POST("/:orgId/groups/:groupId/roles", orgHandler.AssignRoleToGroupInOrganization)
			orgV1.DELETE("/:orgId/groups/:groupId/roles/:roleId", orgHandler.RemoveRoleFromGroupInOrganization)
			orgV1.GET("/:orgId/groups/:groupId/roles", orgHandler.GetGroupRolesInOrganization)
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

			// Organization-scoped group management routes
			orgV2.GET("/:orgId/groups", orgHandler.GetOrganizationGroups)
			orgV2.POST("/:orgId/groups", orgHandler.CreateGroupInOrganization)
			orgV2.GET("/:orgId/groups/:groupId", orgHandler.GetGroupInOrganization)

			// User-group management within organization context
			orgV2.POST("/:orgId/groups/:groupId/users", orgHandler.AddUserToGroupInOrganization)
			orgV2.DELETE("/:orgId/groups/:groupId/users/:userId", orgHandler.RemoveUserFromGroupInOrganization)
			orgV2.GET("/:orgId/groups/:groupId/users", orgHandler.GetGroupUsersInOrganization)
			orgV2.GET("/:orgId/users/:userId/groups", orgHandler.GetUserGroupsInOrganization)
			orgV2.GET("/:orgId/users/:userId/effective-roles", orgHandler.GetUserEffectiveRolesInOrganization)

			// Role-group management within organization context
			orgV2.POST("/:orgId/groups/:groupId/roles", orgHandler.AssignRoleToGroupInOrganization)
			orgV2.DELETE("/:orgId/groups/:groupId/roles/:roleId", orgHandler.RemoveRoleFromGroupInOrganization)
			orgV2.GET("/:orgId/groups/:groupId/roles", orgHandler.GetGroupRolesInOrganization)
		}
	}
}
