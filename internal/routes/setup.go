package routes

import (
	"github.com/Kisanlink/aaa-service/internal/handlers/admin"
	"github.com/Kisanlink/aaa-service/internal/handlers/permissions"
	"github.com/Kisanlink/aaa-service/internal/handlers/roles"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	"github.com/Kisanlink/aaa-service/internal/services"
	contactService "github.com/Kisanlink/aaa-service/internal/services/contacts"
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
	PermissionHandler    *permissions.PermissionHandler
	UserService          interfaces.UserService
	RoleService          interfaces.RoleService
	ContactService       interface{} // Using interface{} to avoid circular dependency
	Validator            interfaces.Validator
	Responder            interfaces.Responder
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

	// Setup V2 auth routes with AuthHandler for MPIN management (replaces old auth routes)
	if handlers.UserService != nil && handlers.Validator != nil && handlers.Responder != nil {
		SetupAuthV2Routes(publicAPI, protectedAPI, handlers.AuthMiddleware, handlers.UserService, handlers.Validator, handlers.Responder, handlers.Logger)
	} else {
		// Fallback to old auth routes if V2 dependencies are not available
		SetupAuthRoutes(publicAPI, protectedAPI, handlers.AuthService, handlers.Logger)
	}

	SetupUserRoutes(protectedAPI, handlers.AuthMiddleware, handlers.UserService, handlers.RoleService, handlers.Validator, handlers.Responder, handlers.Logger)
	SetupRoleRoutes(protectedAPI, handlers.AuthMiddleware, handlers.RoleHandler, handlers.Logger)
	SetupPermissionRoutes(protectedAPI, handlers.AuthMiddleware, handlers.PermissionHandler, handlers.Logger)
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
	if contactService, ok := handlers.ContactService.(*contactService.ContactService); ok {
		SetupContactRoutes(protectedAPI, handlers.AuthMiddleware, contactService, handlers.Validator, handlers.Responder, handlers.Logger)
	}
}

// SetupMiddleware configures global middleware for the router
func SetupMiddleware(router *gin.Engine, middleware ...gin.HandlerFunc) {
	for _, m := range middleware {
		router.Use(m)
	}
}

// SetupAAAWrapper is a wrapper function that maintains backward compatibility
func SetupAAAWrapper(router *gin.Engine, authService *services.AuthService, authzService *services.AuthorizationService, auditService *services.AuditService, authMiddleware *middleware.AuthMiddleware, logger *zap.Logger) {
	// Create default handlers for backward compatibility
	// Note: PermissionHandler requires DBManager which is not available in this wrapper
	// For backward compatibility, we'll create a nil handler
	var permissionHandler *permissions.PermissionHandler = nil

	handlers := RouteHandlers{
		AuthService:          authService,
		AuthorizationService: authzService,
		AuditService:         auditService,
		AuthMiddleware:       authMiddleware,
		AdminHandler:         nil, // AdminHandler will be set separately if needed
		RoleHandler:          nil, // RoleHandler will be set separately if needed
		PermissionHandler:    permissionHandler,
		UserService:          nil, // UserService will be set separately if needed
		RoleService:          nil, // RoleService will be set separately if needed
		ContactService:       nil, // ContactService will be set separately if needed
		Validator:            nil, // Validator will be set separately if needed
		Responder:            nil, // Responder will be set separately if needed
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}

// SetupAAAWithAdmin is an extended wrapper that includes AdminHandler
func SetupAAAWithAdmin(router *gin.Engine, authService *services.AuthService, authzService *services.AuthorizationService, auditService *services.AuditService, authMiddleware *middleware.AuthMiddleware, adminHandler *admin.AdminHandler, roleHandler *roles.RoleHandler, permissionHandler *permissions.PermissionHandler, userService interfaces.UserService, roleService interfaces.RoleService, contactService interface{}, validator interfaces.Validator, responder interfaces.Responder, logger *zap.Logger) {
	handlers := RouteHandlers{
		AuthService:          authService,
		AuthorizationService: authzService,
		AuditService:         auditService,
		AuthMiddleware:       authMiddleware,
		AdminHandler:         adminHandler,
		RoleHandler:          roleHandler,
		PermissionHandler:    permissionHandler,
		UserService:          userService,
		RoleService:          roleService,
		ContactService:       contactService,
		Validator:            validator,
		Responder:            responder,
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}
