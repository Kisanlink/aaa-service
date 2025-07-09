package server

import (
	"context"

	"github.com/Kisanlink/aaa-service/handlers/addresses"
	"github.com/Kisanlink/aaa-service/handlers/admin"
	"github.com/Kisanlink/aaa-service/handlers/auth"
	"github.com/Kisanlink/aaa-service/handlers/permissions"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	router    *gin.Engine
	dbManager *db.DatabaseManager
	port      string
	handlers  *ServerHandlers
	logger    *zap.Logger
}

// ServerHandlers contains all HTTP handlers
type ServerHandlers struct {
	AuthHandler       *auth.AuthHandler
	UserHandler       *users.UserHandler
	RoleHandler       *roles.RoleHandler
	PermissionHandler *permissions.PermissionHandler
	AddressHandler    *addresses.AddressHandler
	AdminHandler      *admin.AdminHandler
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(dbManager *db.DatabaseManager, port string, logger *zap.Logger) *HTTPServer {
	// Initialize Gin router
	router := gin.New()

	// Initialize services
	addressService := services.NewAddressService(logger) // TODO: Fix logger interface issue
	roleService := services.NewRoleService(logger)       // TODO: Fix logger interface issue

	// Initialize utils
	validator := utils.NewValidator(logger) // TODO: Fix logger interface issue
	responder := utils.NewResponder(logger) // TODO: Fix logger interface issue

	// Initialize handlers
	pgManager := dbManager.GetPostgresManager()
	db, _ := pgManager.GetDB(context.TODO(), false) // TODO: Handle error properly
	userService := services.NewUserService(db, db, db, nil, logger, validator)
	userHandler := users.NewUserHandler(userService, validator, responder, logger)

	handlers := &ServerHandlers{
		AuthHandler:       auth.NewAuthHandler(userService, validator, responder, logger),
		UserHandler:       userHandler,
		RoleHandler:       roles.NewRoleHandler(roleService, validator, responder, logger),
		PermissionHandler: permissions.NewPermissionHandler(validator, responder, logger),
		AddressHandler:    addresses.NewAddressHandler(addressService, validator, responder, logger),
		AdminHandler:      admin.NewAdminHandler(validator, responder, logger),
	}

	server := &HTTPServer{
		router:    router,
		dbManager: dbManager,
		port:      port,
		handlers:  handlers,
		logger:    logger,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	logger.Info("HTTP server initialized successfully")
	return server
}

// setupMiddleware configures middleware for the router
func (s *HTTPServer) setupMiddleware() {
	routes.SetupMiddleware(s.router,
		cors.Default(),
		middleware.RequestID(),
		middleware.Logger(s.logger), // TODO: Fix logger interface issue
		middleware.ErrorHandler(),
		middleware.Recovery(),
	)
}

// setupRoutes configures all the routes using the routes package
func (s *HTTPServer) setupRoutes() {
	// Setup all routes using the centralized routes package
	routes.SetupRoutes(s.router, routes.AllHandlers{
		AuthHandler:       s.handlers.AuthHandler,
		UserHandler:       s.handlers.UserHandler,
		RoleHandler:       s.handlers.RoleHandler,
		PermissionHandler: s.handlers.PermissionHandler,
		AddressHandler:    s.handlers.AddressHandler,
		AdminHandler:      s.handlers.AdminHandler,
	})
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("port", s.port))
	return s.router.Run(":" + s.port)
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	// TODO: Implement graceful shutdown
	return nil
}

// GetRouter returns the gin router (useful for testing)
func (s *HTTPServer) GetRouter() *gin.Engine {
	return s.router
}

// All route handlers have been moved to their respective handler packages
// and route definitions have been moved to the routes package

// Role Management V2 handlers have been moved to role handler package

// Permission Management V2 handlers have been moved to permission handler package

// Address Management handlers have been moved to address handler package

// Admin Management V2 handlers have been moved to admin handler package
