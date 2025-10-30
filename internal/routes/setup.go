package routes

import (
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/handlers/admin"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/groups"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/organizations"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/permissions"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/roles"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	contactService "github.com/Kisanlink/aaa-service/v2/internal/services/contacts"
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
	OrganizationService  interfaces.OrganizationService
	GroupService         interfaces.GroupService
	Validator            interfaces.Validator
	Responder            interfaces.Responder
	Logger               *zap.Logger
}

// SetupAAA configures all AAA routes with proper authentication and authorization
func SetupAAA(router *gin.Engine, handlers RouteHandlers) {
	// Apply global security middleware
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestSizeLimit(10 * 1024 * 1024)) // 10MB limit
	router.Use(middleware.Timeout(30 * time.Second))
	if handlers.Logger != nil {
		// Note: Using gin's default logger for now, can be enhanced later
		router.Use(gin.Logger())
		router.Use(gin.Recovery())
		router.Use(middleware.SecureErrorHandler(middleware.NewErrorHandlerConfig(handlers.Logger)))
	}

	// Public routes (no authentication required)
	publicAPI := router.Group("/api/v1")
	publicAPI.Use(middleware.RateLimit()) // General rate limiting for public endpoints

	// Protected routes (authentication and authorization required)
	protectedAPI := router.Group("/api/v1")
	protectedAPI.Use(handlers.AuthMiddleware.HTTPAuthMiddleware())
	protectedAPI.Use(middleware.SensitiveOperationRateLimit()) // More restrictive rate limiting

	// Setup route groups
	SetupHealthRoutes(publicAPI, handlers.Logger)

	// Setup auth routes with AuthHandler
	if handlers.UserService != nil && handlers.Validator != nil && handlers.Responder != nil {
		SetupAuthRoutes(publicAPI, protectedAPI, handlers.AuthMiddleware, handlers.UserService, handlers.Validator, handlers.Responder, handlers.Logger)
	} else {
		handlers.Logger.Warn("Auth dependencies not available - auth routes will not be registered")
	}

	SetupUserRoutes(protectedAPI, handlers.AuthMiddleware, handlers.UserService, handlers.RoleService, handlers.Validator, handlers.Responder, handlers.Logger)
	SetupRoleRoutes(protectedAPI, handlers.AuthMiddleware, handlers.RoleHandler, handlers.Logger)
	SetupPermissionRoutes(protectedAPI, handlers.AuthMiddleware, handlers.PermissionHandler, handlers.Logger)
	SetupAuthorizationRoutes(protectedAPI, handlers.AuthorizationService, handlers.Logger)
	SetupAuditRoutes(protectedAPI, handlers.AuthMiddleware, handlers.AuditService, handlers.Logger)
	// Setup admin routes if AdminHandler is provided
	if handlers.AdminHandler != nil {
		if adminHandler, ok := handlers.AdminHandler.(*admin.AdminHandler); ok {
			SetupAdminRoutes(protectedAPI, adminHandler, handlers.AuthMiddleware)
		}
	} else {
		// Fallback to legacy admin routes setup if AdminHandler not available
		SetupLegacyAdminRoutes(protectedAPI, handlers.AuthMiddleware, handlers.AuthorizationService, handlers.AuditService, handlers.Logger)
	}

	SetupModuleRoutes(protectedAPI, handlers.Logger)
	if contactService, ok := handlers.ContactService.(*contactService.ContactService); ok {
		SetupContactRoutes(protectedAPI, handlers.AuthMiddleware, contactService, handlers.Validator, handlers.Responder, handlers.Logger)
	}

	// Setup organization routes if services are available
	//nolint:nestif // Complex conditional is necessary for service availability checks
	if handlers.OrganizationService != nil && handlers.GroupService != nil {
		// Import the organization handler package
		orgHandler := organizations.NewOrganizationHandler(
			handlers.OrganizationService,
			handlers.GroupService,
			handlers.Logger,
			handlers.Responder,
		)
		// Setup organization routes
		SetupOrganizationRoutes(protectedAPI, orgHandler, handlers.AuthMiddleware)

		// Also setup standalone group routes for direct group management
		groupHandler := groups.NewGroupHandler(
			handlers.GroupService,
			handlers.Logger,
			handlers.Responder,
		)
		// Register group routes under /api/v1/groups (as per the existing pattern)
		RegisterGroupRoutes(router, groupHandler, handlers.AuthMiddleware)
	} else {
		if handlers.Logger != nil {
			if handlers.OrganizationService == nil {
				handlers.Logger.Warn("Organization service not available - organization and group routes will not be registered")
			}
			if handlers.GroupService == nil {
				handlers.Logger.Warn("Group service not available - organization and group routes will not be registered")
			}
		}
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
		OrganizationService:  nil, // OrganizationService will be set separately if needed
		GroupService:         nil, // GroupService will be set separately if needed
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
		OrganizationService:  nil, // Will be set when services are available
		GroupService:         nil, // Will be set when services are available
		Validator:            validator,
		Responder:            responder,
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}

// SetupAAAWithOrganizations is an extended wrapper that includes organization and group services
func SetupAAAWithOrganizations(router *gin.Engine, authService *services.AuthService, authzService *services.AuthorizationService, auditService *services.AuditService, authMiddleware *middleware.AuthMiddleware, adminHandler *admin.AdminHandler, roleHandler *roles.RoleHandler, permissionHandler *permissions.PermissionHandler, userService interfaces.UserService, roleService interfaces.RoleService, contactService interface{}, organizationService interfaces.OrganizationService, groupService interfaces.GroupService, validator interfaces.Validator, responder interfaces.Responder, logger *zap.Logger) {
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
		OrganizationService:  organizationService,
		GroupService:         groupService,
		Validator:            validator,
		Responder:            responder,
		Logger:               logger,
	}
	SetupAAA(router, handlers)
}
