package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupPermissionRoutes configures permission management routes
func SetupPermissionRoutes(protectedAPI *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware, permissionHandler *permissions.PermissionHandler, logger *zap.Logger) {
	// protectedAPI already has HTTPAuthMiddleware applied in SetupAAA
	perms := protectedAPI.Group("/permissions")
	// Authorization requires authenticated context
	perms.Use(authMiddleware.RequirePermission("permission", "read"))
	{
		// Permission CRUD operations
		perms.GET("", permissionHandler.ListPermissions)
		perms.GET("/:id", permissionHandler.GetPermission)
		perms.POST("", authMiddleware.RequirePermission("permission", "create"), permissionHandler.CreatePermission)
		perms.PUT("/:id", authMiddleware.RequirePermission("permission", "update"), permissionHandler.UpdatePermission)
		perms.DELETE("/:id", authMiddleware.RequirePermission("permission", "delete"), permissionHandler.DeletePermission)

		// Permission evaluation endpoints
		perms.POST("/evaluate", permissionHandler.EvaluatePermission)
	}

	// User-specific permission evaluation
	users := protectedAPI.Group("/users")
	{
		users.POST("/:id/evaluate", authMiddleware.RequirePermission("permission", "read"), permissionHandler.EvaluateUserPermission)
	}

	// Role-permission assignment routes
	roles := protectedAPI.Group("/roles")
	{
		// Model 1: Permission assignments (role_permissions table)
		roles.POST("/:id/permissions", authMiddleware.RequirePermission("role", "update"), permissionHandler.AssignPermissionsToRole)
		roles.DELETE("/:id/permissions/:permId", authMiddleware.RequirePermission("role", "update"), permissionHandler.RevokePermissionFromRole)
		roles.GET("/:id/permissions", authMiddleware.RequirePermission("role", "read"), permissionHandler.GetRolePermissions)

		// Model 2: Resource-action assignments (resource_permissions table)
		roles.POST("/:id/resources", authMiddleware.RequirePermission("role", "update"), permissionHandler.AssignResourcesToRole)
		roles.DELETE("/:id/resources/:resId", authMiddleware.RequirePermission("role", "update"), permissionHandler.RevokeResourceFromRole)
		roles.GET("/:id/resources", authMiddleware.RequirePermission("role", "read"), permissionHandler.GetRoleResources)
	}
}
