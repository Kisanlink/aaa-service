package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/users"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupUserRoutes configures user management routes with proper service integration
func SetupUserRoutes(
	protectedAPI *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
) {
	// Initialize handler with the provided services
	userHandler := users.NewUserHandler(userService, roleService, validator, responder, logger)

	users := protectedAPI.Group("/users")
	{
		// User CRUD operations
		users.POST("", authMiddleware.RequirePermission("user", "create"), userHandler.CreateUser)
		users.GET("", authMiddleware.RequirePermission("user", "read"), userHandler.ListUsers)
		users.GET("/:id", authMiddleware.RequirePermission("user", "view"), userHandler.GetUserByID)
		users.PUT("/:id", authMiddleware.RequirePermission("user", "update"), userHandler.UpdateUser)
		users.DELETE("/:id", authMiddleware.RequirePermission("user", "delete"), userHandler.DeleteUser)

		// User search and validation
		users.GET("/search", authMiddleware.RequirePermission("user", "read"), userHandler.SearchUsers)
		users.POST("/:id/validate", authMiddleware.RequirePermission("user", "update"), userHandler.ValidateUser)

		// User role management
		users.POST("/:id/roles/:roleId", authMiddleware.RequirePermission("user", "update"), userHandler.AssignRole)
		users.DELETE("/:id/roles/:roleId", authMiddleware.RequirePermission("user", "update"), userHandler.RemoveRole)
	}
}
