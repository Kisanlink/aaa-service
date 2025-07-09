package routes

import (
	"github.com/Kisanlink/aaa-service/handlers/addresses"
	"github.com/Kisanlink/aaa-service/handlers/admin"
	"github.com/Kisanlink/aaa-service/handlers/auth"
	"github.com/Kisanlink/aaa-service/handlers/permissions"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/gin-gonic/gin"
)

// AllHandlers contains all the handlers needed for routing
type AllHandlers struct {
	AuthHandler       *auth.AuthHandler
	UserHandler       *users.UserHandler
	RoleHandler       *roles.RoleHandler
	PermissionHandler *permissions.PermissionHandler
	AddressHandler    *addresses.AddressHandler
	AdminHandler      *admin.AdminHandler
}

// SetupRoutes configures all routes for the application
func SetupRoutes(router *gin.Engine, handlers AllHandlers) {
	// Setup health and utility routes
	SetupHealthRoutes(router)

	// Setup API versioned routes
	apiGroup := router.Group("/api")
	{
		// V1 Routes
		v1Group := apiGroup.Group("/v1")
		SetupV1Routes(v1Group, V1RouteHandlers{
			UserHandler:    handlers.UserHandler,
			RoleHandler:    handlers.RoleHandler,
			AddressHandler: handlers.AddressHandler,
		})

		// V2 Routes
		v2Group := apiGroup.Group("/v2")
		SetupV2Routes(v2Group, V2RouteHandlers{
			AuthHandler:       handlers.AuthHandler,
			UserHandler:       handlers.UserHandler,
			RoleHandler:       handlers.RoleHandler,
			PermissionHandler: handlers.PermissionHandler,
			AdminHandler:      handlers.AdminHandler,
		})
	}
}

// SetupMiddleware configures middleware for the router
func SetupMiddleware(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	for _, middleware := range middlewares {
		router.Use(middleware)
	}
}
