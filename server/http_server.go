package server

import (
	"context"
	"net/http"

	"github.com/Kisanlink/aaa-service/handlers/addresses"
	"github.com/Kisanlink/aaa-service/handlers/admin"
	"github.com/Kisanlink/aaa-service/handlers/auth"
	"github.com/Kisanlink/aaa-service/handlers/health"
	"github.com/Kisanlink/aaa-service/handlers/permissions"
	"github.com/Kisanlink/aaa-service/handlers/roles"
	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/middleware"
	addressRepo "github.com/Kisanlink/aaa-service/repositories/addresses"
	roleRepo "github.com/Kisanlink/aaa-service/repositories/roles"
	userRepo "github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	router             *gin.Engine
	dbManager          *db.DatabaseManager
	maintenanceService interfaces.MaintenanceService
	port               string
	handlers           *ServerHandlers
	logger             *zap.Logger
	server             *http.Server
}

// ServerHandlers contains all HTTP handlers
type ServerHandlers struct {
	AuthHandler       *auth.AuthHandler
	UserHandler       *users.UserHandler
	RoleHandler       *roles.RoleHandler
	PermissionHandler *permissions.PermissionHandler
	AddressHandler    *addresses.AddressHandler
	AdminHandler      *admin.AdminHandler
	HealthHandler     *health.HealthHandler
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(dbManager *db.DatabaseManager, port string, logger *zap.Logger) *HTTPServer {
	// Initialize Gin router
	router := gin.New()

	// Initialize logger adapter for interfaces.Logger compatibility
	loggerAdapter := utils.NewLoggerAdapter(logger)

	// Get the primary database manager for repositories
	primaryDBManager := dbManager.GetManager(dbManager.GetPostgresManager().GetBackendType())
	if primaryDBManager == nil {
		// Fallback to in-memory for testing
		logger.Warn("No database manager available, using in-memory fallback")
		primaryDBManager = &InMemoryDBManager{logger: loggerAdapter}
	}

	// Initialize repositories
	userRepository := userRepo.NewUserRepository(primaryDBManager)
	addressRepository := addressRepo.NewAddressRepository(primaryDBManager)
	roleRepository := roleRepo.NewRoleRepository(primaryDBManager)
	userRoleRepository := roleRepo.NewUserRoleRepository(primaryDBManager)

	// Initialize cache service
	cacheService := services.NewCacheService("localhost:6379", "", 0, loggerAdapter)

	// Initialize maintenance service
	maintenanceService := services.NewMaintenanceService(cacheService, loggerAdapter)

	// Initialize utils
	validator := utils.NewValidator()
	responder := utils.NewResponder(loggerAdapter)

	// Initialize services
	addressService := services.NewAddressService(addressRepository, cacheService, loggerAdapter, validator)
	roleService := services.NewRoleService(roleRepository, userRoleRepository, cacheService, loggerAdapter, validator)
	userService := services.NewUserService(userRepository, roleRepository, userRoleRepository, cacheService, logger, validator)

	userHandler := users.NewUserHandler(userService, roleService, validator, responder, logger)

	handlers := &ServerHandlers{
		AuthHandler:       auth.NewAuthHandler(userService, validator, responder, logger),
		UserHandler:       userHandler,
		RoleHandler:       roles.NewRoleHandler(roleService, validator, responder, logger),
		PermissionHandler: permissions.NewPermissionHandler(validator, responder, logger),
		AddressHandler:    addresses.NewAddressHandler(addressService, validator, responder, logger),
		AdminHandler:      admin.NewAdminHandler(maintenanceService, validator, responder, logger),
		HealthHandler:     health.NewHealthHandler(dbManager, cacheService, responder, logger),
	}

	server := &HTTPServer{
		router:             router,
		dbManager:          dbManager,
		maintenanceService: maintenanceService,
		port:               port,
		handlers:           handlers,
		logger:             logger,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	logger.Info("HTTP server initialized successfully")
	return server
}

// InMemoryDBManager is a fallback implementation for testing
type InMemoryDBManager struct {
	logger interfaces.Logger
}

func (m *InMemoryDBManager) Connect(ctx context.Context) error                   { return nil }
func (m *InMemoryDBManager) Close() error                                        { return nil }
func (m *InMemoryDBManager) IsConnected() bool                                   { return true }
func (m *InMemoryDBManager) GetBackendType() db.BackendType                      { return db.BackendInMemory }
func (m *InMemoryDBManager) Create(ctx context.Context, model interface{}) error { return nil }
func (m *InMemoryDBManager) GetByID(ctx context.Context, id interface{}, model interface{}) error {
	return nil
}
func (m *InMemoryDBManager) Update(ctx context.Context, model interface{}) error { return nil }
func (m *InMemoryDBManager) Delete(ctx context.Context, id interface{}) error    { return nil }
func (m *InMemoryDBManager) List(ctx context.Context, filters []db.Filter, model interface{}) error {
	return nil
}
func (m *InMemoryDBManager) ApplyFilters(query interface{}, filters []db.Filter) (interface{}, error) {
	return query, nil
}
func (m *InMemoryDBManager) BuildFilter(field string, operator db.FilterOperator, value interface{}) db.Filter {
	return db.Filter{Field: field, Operator: operator, Value: value}
}

func (m *InMemoryDBManager) AutoMigrateModels(ctx context.Context, models ...interface{}) error {
	// In-memory database doesn't support schema migration
	return nil
}

// setupMiddleware configures middleware for the router
func (s *HTTPServer) setupMiddleware() {
	loggerAdapter := utils.NewLoggerAdapter(s.logger)
	responder := utils.NewResponder(loggerAdapter)

	routes.SetupMiddleware(s.router,
		cors.Default(),
		middleware.RequestID(),
		middleware.Logger(loggerAdapter),
		middleware.ErrorHandler,
		middleware.PanicRecoveryHandler(loggerAdapter),
		middleware.MaintenanceMode(s.maintenanceService, responder, loggerAdapter),
	)
}

// setupRoutes configures all the routes using the routes package
func (s *HTTPServer) setupRoutes() {
	// Setup health routes with the health handler
	routes.SetupHealthRoutes(s.router, s.handlers.HealthHandler)

	// Serve OpenAPI (Swagger) files and Scalar-powered docs UI
	s.router.StaticFile("/docs/swagger.json", "docs/swagger.json")
	s.router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")
	s.router.GET("/docs", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		specURL := scheme + "://" + c.Request.Host + "/docs/swagger.json"

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL:  specURL,
			DarkMode: true,
			CustomOptions: scalar.CustomOptions{
				PageTitle: "AAA Service API Reference",
			},
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to render API docs: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	})

	// Redirect root to docs for convenience
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	// Setup all other routes using the centralized routes package
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
	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
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
