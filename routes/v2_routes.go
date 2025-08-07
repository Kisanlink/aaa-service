package routes

import (
	"net/http"

	"github.com/Kisanlink/aaa-service/handlers/admin"
	"github.com/Kisanlink/aaa-service/handlers/auth"
	"github.com/Kisanlink/aaa-service/handlers/permissions"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/gin-gonic/gin"
)

// V2RouteHandlers contains all handlers needed for V2 routes
type V2RouteHandlers struct {
	AuthHandler       *auth.AuthHandler
	UserHandler       *users.UserHandler
	RoleHandler       *roles.RoleHandler
	PermissionHandler *permissions.PermissionHandler
	AdminHandler      *admin.AdminHandler
}

// SetupV2Routes configures all V2 API routes
func SetupV2Routes(router *gin.RouterGroup, handlers V2RouteHandlers) {
	// Authentication V2
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", handlers.AuthHandler.LoginV2)
		authRoutes.POST("/register", handlers.AuthHandler.RegisterV2)
		authRoutes.POST("/refresh", handlers.AuthHandler.RefreshTokenV2)
		authRoutes.POST("/logout", handlers.AuthHandler.LogoutV2)
		authRoutes.POST("/forgot-password", handlers.AuthHandler.ForgotPasswordV2)
		authRoutes.POST("/reset-password", handlers.AuthHandler.ResetPasswordV2)

		// MFA routes (TODO: Implement when MFA service is available)
		authRoutes.POST("/mfa/setup", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "MFA setup - To be implemented"})
		})
		authRoutes.POST("/mfa/verify", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "MFA verify - To be implemented"})
		})
		authRoutes.POST("/mfa/disable", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "MFA disable - To be implemented"})
		})

		// Social authentication (TODO: Implement when social auth service is available)
		authRoutes.GET("/social/:provider", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Social auth - To be implemented"})
		})
		authRoutes.POST("/social/:provider/callback", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Social auth callback - To be implemented"})
		})
	}

	// User Management V2
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", handlers.UserHandler.CreateUser)
		userRoutes.GET("/:id", handlers.UserHandler.GetUserByID)
		userRoutes.PUT("/:id", handlers.UserHandler.UpdateUser)
		userRoutes.DELETE("/:id", handlers.UserHandler.DeleteUser)
		userRoutes.GET("", handlers.UserHandler.ListUsers)
		userRoutes.GET("/:id/profile", handlers.UserHandler.GetUserByID) // Reuse existing method
		userRoutes.PUT("/:id/profile", handlers.UserHandler.UpdateUser)  // Reuse existing method
		userRoutes.POST("/:id/roles", handlers.UserHandler.AssignRole)
		userRoutes.DELETE("/:id/roles/:roleId", handlers.UserHandler.RemoveRole)
		userRoutes.GET("/:id/roles", handlers.UserHandler.GetUserByID)       // Will return user with roles
		userRoutes.GET("/:id/permissions", handlers.UserHandler.GetUserByID) // Will return user with permissions
	}

	// Role Management V2
	roleRoutes := router.Group("/roles")
	{
		roleRoutes.POST("", handlers.RoleHandler.CreateRoleV2)
		roleRoutes.GET("/:id", handlers.RoleHandler.GetRoleV2)
		roleRoutes.PUT("/:id", handlers.RoleHandler.UpdateRoleV2)
		roleRoutes.DELETE("/:id", handlers.RoleHandler.DeleteRoleV2)
		roleRoutes.GET("", handlers.RoleHandler.ListRolesV2)
		roleRoutes.POST("/:id/permissions", handlers.RoleHandler.AssignPermissionV2)
		roleRoutes.DELETE("/:id/permissions/:permissionId", handlers.RoleHandler.RemovePermissionV2)
		roleRoutes.GET("/:id/permissions", handlers.RoleHandler.GetRolePermissionsV2)
		roleRoutes.GET("/hierarchy", handlers.RoleHandler.GetRoleHierarchyV2)
		roleRoutes.POST("/:id/children", handlers.RoleHandler.AddChildRoleV2)
	}

	// Permission Management V2
	permissionRoutes := router.Group("/permissions")
	{
		permissionRoutes.POST("", handlers.PermissionHandler.CreatePermissionV2)
		permissionRoutes.GET("/:id", handlers.PermissionHandler.GetPermissionV2)
		permissionRoutes.PUT("/:id", handlers.PermissionHandler.UpdatePermissionV2)
		permissionRoutes.DELETE("/:id", handlers.PermissionHandler.DeletePermissionV2)
		permissionRoutes.GET("", handlers.PermissionHandler.ListPermissionsV2)
		permissionRoutes.POST("/evaluate", handlers.PermissionHandler.EvaluatePermissionV2)
		permissionRoutes.POST("/temporary", handlers.PermissionHandler.GrantTemporaryPermissionV2)
	}

	// Admin Management V2
	adminRoutes := router.Group("/admin")
	{
		adminRoutes.GET("/health/detailed", handlers.AdminHandler.DetailedHealthCheckV2)
		adminRoutes.GET("/metrics", handlers.AdminHandler.MetricsV2)
		adminRoutes.GET("/audit", handlers.AdminHandler.AuditLogsV2)
		adminRoutes.POST("/maintenance", handlers.AdminHandler.MaintenanceModeV2)
		adminRoutes.GET("/system", handlers.AdminHandler.GetSystemInfo)
	}
}
