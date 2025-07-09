package routes

import (
	"github.com/Kisanlink/aaa-service/handlers/addresses"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/gin-gonic/gin"
)

// V1RouteHandlers contains all handlers needed for V1 routes
type V1RouteHandlers struct {
	UserHandler    *users.UserHandler
	RoleHandler    *roles.RoleHandler
	AddressHandler *addresses.AddressHandler
}

// SetupV1Routes configures all V1 API routes
func SetupV1Routes(router *gin.RouterGroup, handlers V1RouteHandlers) {
	// User Management V1
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", handlers.UserHandler.CreateUser)
		userRoutes.GET("/:id", handlers.UserHandler.GetUserByID)
		userRoutes.PUT("/:id", handlers.UserHandler.UpdateUser)
		userRoutes.DELETE("/:id", handlers.UserHandler.DeleteUser)
		userRoutes.GET("", handlers.UserHandler.ListUsers)
		userRoutes.POST("/:id/assign-role", handlers.UserHandler.AssignRole)
		userRoutes.DELETE("/:id/unassign-role", handlers.UserHandler.RemoveRole)
	}

	// Role Management V1
	roleRoutes := router.Group("/roles")
	{
		roleRoutes.POST("", handlers.RoleHandler.CreateRole)
		roleRoutes.GET("/:id", handlers.RoleHandler.GetRole)
		roleRoutes.PUT("/:id", handlers.RoleHandler.UpdateRole)
		roleRoutes.DELETE("/:id", handlers.RoleHandler.DeleteRole)
		roleRoutes.GET("", handlers.RoleHandler.ListRoles)
	}

	// Address Management V1
	addressRoutes := router.Group("/addresses")
	{
		addressRoutes.POST("", handlers.AddressHandler.CreateAddress)
		addressRoutes.GET("/:id", handlers.AddressHandler.GetAddress)
		addressRoutes.PUT("/:id", handlers.AddressHandler.UpdateAddress)
		addressRoutes.DELETE("/:id", handlers.AddressHandler.DeleteAddress)
		addressRoutes.GET("/search", handlers.AddressHandler.SearchAddresses)
	}
}
