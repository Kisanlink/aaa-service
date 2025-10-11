package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/organizations"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupOrganizationRoutes configures organization-related routes
// This function accepts a router group (typically the protected API group) and adds organization routes to it
func SetupOrganizationRoutes(apiGroup *gin.RouterGroup, orgHandler *organizations.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Organization routes with authentication
	// Note: The apiGroup is expected to be /api/v2 with auth middleware already applied
	org := apiGroup.Group("/organizations")
	{
		org.POST("", orgHandler.CreateOrganization)
		org.GET("", orgHandler.ListOrganizations)
		org.GET("/:id", orgHandler.GetOrganization)
		org.PUT("/:id", orgHandler.UpdateOrganization)
		org.DELETE("/:id", orgHandler.DeleteOrganization)
		org.GET("/:id/hierarchy", orgHandler.GetOrganizationHierarchy)
		org.POST("/:id/activate", orgHandler.ActivateOrganization)
		org.POST("/:id/deactivate", orgHandler.DeactivateOrganization)
		org.GET("/:id/stats", orgHandler.GetOrganizationStats)

		// Organization-scoped group management routes
		org.GET("/:id/groups", orgHandler.GetOrganizationGroups)
		org.POST("/:id/groups", orgHandler.CreateGroupInOrganization)
		org.GET("/:id/groups/:groupId", orgHandler.GetGroupInOrganization)

		// Group update/delete - restricted to super_admin only
		org.PUT("/:id/groups/:groupId", authMiddleware.RequireRole("super_admin"), orgHandler.UpdateGroupInOrganization)
		org.DELETE("/:id/groups/:groupId", authMiddleware.RequireRole("super_admin"), orgHandler.DeleteGroupInOrganization)

		// User-group management within organization context
		org.POST("/:id/groups/:groupId/users", orgHandler.AddUserToGroupInOrganization)
		org.DELETE("/:id/groups/:groupId/users/:userId", orgHandler.RemoveUserFromGroupInOrganization)
		org.GET("/:id/groups/:groupId/users", orgHandler.GetGroupUsersInOrganization)
		org.GET("/:id/users/:userId/groups", orgHandler.GetUserGroupsInOrganization)
		org.GET("/:id/users/:userId/effective-roles", orgHandler.GetUserEffectiveRolesInOrganization)

		// Role-group management within organization context
		org.POST("/:id/groups/:groupId/roles", orgHandler.AssignRoleToGroupInOrganization)
		org.DELETE("/:id/groups/:groupId/roles/:roleId", orgHandler.RemoveRoleFromGroupInOrganization)
		org.GET("/:id/groups/:groupId/roles", orgHandler.GetGroupRolesInOrganization)
	}
}
