package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterGroupRoutes registers all group-related routes
func RegisterGroupRoutes(router *gin.Engine, groupHandler *groups.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Group management routes
	groupRoutes := router.Group("/api/v1/groups")
	{
		// Public routes (if any)
		groupRoutes.GET("/:id", groupHandler.GetGroup)
		groupRoutes.GET("", groupHandler.ListGroups)

		// Protected routes requiring authentication
		authenticated := groupRoutes.Group("")
		authenticated.Use(authMiddleware.HTTPAuthMiddleware())
		{
			authenticated.POST("", groupHandler.CreateGroup)
			authenticated.PUT("/:id", groupHandler.UpdateGroup)
			authenticated.DELETE("/:id", groupHandler.DeleteGroup)

			// Group membership routes
			authenticated.POST("/:id/members", groupHandler.AddMemberToGroup)
			authenticated.DELETE("/:id/members/:principal_id", groupHandler.RemoveMemberFromGroup)
			authenticated.GET("/:id/members", groupHandler.GetGroupMembers)
		}
	}
}
