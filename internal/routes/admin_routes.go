package routes

import (
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/admin"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupAdminRoutes configures admin routes
func SetupAdminRoutes(protectedAPI *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware, authzService *services.AuthorizationService, auditService *services.AuditService, logger *zap.Logger) {
	admin := protectedAPI.Group("/admin")
	admin.Use(authMiddleware.RequireRole("admin"))
	{
		admin.POST("/grant-permission", createGrantPermissionHandler(authzService, logger))
		admin.POST("/revoke-permission", createRevokePermissionHandler(authzService, logger))
		admin.POST("/assign-role", createAssignRoleHandler(authzService, logger))
		admin.POST("/remove-role", createRemoveRoleHandler(authzService, logger))
		admin.POST("/archive-logs", createArchiveLogsHandler(auditService, logger))
	}
}

// SetupAdminV2Routes configures v2 admin routes with AdminHandler
func SetupAdminV2Routes(protectedAPI *gin.RouterGroup, adminHandler *admin.AdminHandler, authMiddleware *middleware.AuthMiddleware) {
	adminGroup := protectedAPI.Group("/admin")
	adminGroup.Use(authMiddleware.RequireRole("super_admin", "admin"))
	{
		// Audit endpoint
		adminGroup.GET("/audit", adminHandler.AuditLogsV2)

		// Health and metrics endpoints
		adminGroup.GET("/health/detailed", adminHandler.DetailedHealthCheckV2)
		adminGroup.GET("/metrics", adminHandler.MetricsV2)

		// System info endpoint
		adminGroup.GET("/system", adminHandler.GetSystemInfo)

		// Maintenance endpoints
		adminGroup.GET("/maintenance", adminHandler.GetMaintenanceStatus)
		adminGroup.POST("/maintenance", adminHandler.MaintenanceModeV2)
		adminGroup.PATCH("/maintenance/message", adminHandler.UpdateMaintenanceMessage)
	}
}

func createGrantPermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Grant permission endpoint accessed")
		c.JSON(200, gin.H{"message": "Grant permission endpoint - implementation needed"})
	}
}

func createRevokePermissionHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Revoke permission endpoint accessed")
		c.JSON(200, gin.H{"message": "Revoke permission endpoint - implementation needed"})
	}
}

func createAssignRoleHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Assign role endpoint accessed")
		c.JSON(200, gin.H{"message": "Assign role endpoint - implementation needed"})
	}
}

func createRemoveRoleHandler(authzService *services.AuthorizationService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Remove role endpoint accessed")
		c.JSON(200, gin.H{"message": "Remove role endpoint - implementation needed"})
	}
}

func createArchiveLogsHandler(auditService *services.AuditService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Archive logs endpoint accessed")
		c.JSON(200, gin.H{"message": "Archive logs endpoint - implementation needed"})
	}
}
