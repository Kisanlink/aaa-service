package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/admin"
	"github.com/Kisanlink/aaa-service/internal/handlers/roles"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteHandlers contains all handlers needed for AAA routes
type RouteHandlers struct {
	AuthService          *services.AuthService
	AuthorizationService *services.AuthorizationService
	AuditService         *services.AuditService
	AuthMiddleware       *middleware.AuthMiddleware
	AdminHandler         interface{} // Using interface{} to avoid circular dependency
	RoleHandler          *roles.RoleHandler
	Logger               *zap.Logger
}

// SetupAAA configures all AAA routes with proper authentication and authorization
func SetupAAA(router *gin.Engine, handlers RouteHandlers) {
	// Public routes (no authentication required)
	publicAPI := router.Group("/api/v2")

	// Protected routes (authentication and authorization required)
	protectedAPI := router.Group("/api/v2")
	protectedAPI.Use(handlers.AuthMiddleware.HTTPAuthMiddleware())

	// Setup route groups
	SetupHealthRoutes(publicAPI, handlers.Logger)
	SetupAuthRoutes(publicAPI, protectedAPI, handlers.AuthService, handlers.Logger)
	SetupUserRoutes(protectedAPI, handlers.AuthMiddleware, handlers.Logger)
	SetupRoleRoutes(protectedAPI, handlers.AuthMiddleware, handlers.RoleHandler, handlers.Logger)
	SetupPermissionRoutes(protectedAPI, handlers.AuthMiddleware, handlers.Logger)
	SetupAuthorizationRoutes(protectedAPI, handlers.AuthorizationService, handlers.Logger)
	SetupAuditRoutes(protectedAPI, handlers.AuthMiddleware, handlers.AuditService, handlers.Logger)
	SetupAdminRoutes(protectedAPI, handlers.AuthMiddleware, handlers.AuthorizationService, handlers.AuditService, handlers.Logger)

	// Setup v2 admin routes if AdminHandler is provided
	if handlers.AdminHandler != nil {
		if adminHandler, ok := handlers.AdminHandler.(*admin.AdminHandler); ok {
			SetupAdminV2Routes(protectedAPI, adminHandler, handlers.AuthMiddleware)
		}
	}

	SetupModuleRoutes(protectedAPI, handlers.Logger)
}

// SetupMiddleware configures global middleware for the router
func SetupMiddleware(router *gin.Engine, middleware ...gin.HandlerFunc) {
	for _, m := range middleware {
		router.Use(m)
	}
}

// SetupAAAWrapper is a wrapper function that maintains backward compatibility
func SetupAAAWrapper(router *gin.Engine, authService *services.AuthService, authzService *services.AuthorizationService, auditService *services.AuditService, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) {
	handlers := RouteHandlers{
		AuthService:          authService,
		AuthorizationService: authzService,
		AuditService:         auditService,
		AuthMiddleware:       authMiddleware,
		AdminHandler:         nil, // AdminHandler will be set separately if needed
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}

// SetupAAAWithAdmin is an extended wrapper that includes AdminHandler
func SetupAAAWithAdmin(router *gin.Engine, authService *services.AuthService, authzService *services.AuthorizationService, auditService *services.AuditService, authMiddleware *middleware.AuthMiddleware, adminHandler *admin.AdminHandler, roleHandler *roles.RoleHandler, logger *zap.Logger) {
	handlers := RouteHandlers{
		AuthService:          authService,
		AuthorizationService: authzService,
		AuditService:         auditService,
		AuthMiddleware:       authMiddleware,
		AdminHandler:         adminHandler,
		RoleHandler:          roleHandler,
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}
