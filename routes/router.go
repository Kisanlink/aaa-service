package routes

import (
	"github.com/Kisanlink/aaa-service/handlers/addresses"
	"github.com/Kisanlink/aaa-service/handlers/admin"
	"github.com/Kisanlink/aaa-service/handlers/auth"
	"github.com/Kisanlink/aaa-service/handlers/permissions"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	// Setup API routes - V2 only
	apiGroup := router.Group("/api")
	{
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

// SetupAAAWrapper configures AAA routes for the application
func SetupAAAWrapper(
	router *gin.Engine,
	authService *services.AuthService,
	authzService *services.AuthorizationService,
	auditService *services.AuditService,
	authMiddleware *middleware.AuthMiddleware,
	logger *zap.Logger,
) {
	handlers := RouteHandlers{
		AuthService:          authService,
		AuthorizationService: authzService,
		AuditService:         auditService,
		AuthMiddleware:       authMiddleware,
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}

// SetupMiddleware configures middleware for the router
func SetupMiddleware(router *gin.Engine, middlewares ...gin.HandlerFunc) {
	for _, middleware := range middlewares {
		router.Use(middleware)
	}
}
