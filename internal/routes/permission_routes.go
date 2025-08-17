package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/permissions"
	"github.com/Kisanlink/aaa-service/internal/middleware"
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
		perms.GET("", permissionHandler.ListPermissionsV2)
		// For create, require create permission specifically
		perms.POST("", authMiddleware.RequirePermission("permission", "create"), permissionHandler.CreatePermissionV2)
	}
}
