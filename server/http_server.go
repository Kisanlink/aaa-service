package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/aaa-service/handlers/users"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/middleware"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/gin-gonic/gin"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	server       *http.Server
	router       *gin.Engine
	logger       interfaces.Logger
	dbManager    interfaces.DatabaseManager
	cacheService interfaces.CacheService
	services     *Services
	handlers     *Handlers
	utils        *Utils
}

// Services contains all service instances
type Services struct {
	UserService    interfaces.UserService
	AddressService interfaces.AddressService
	RoleService    interfaces.RoleService
}

// Handlers contains all handler instances
type Handlers struct {
	UserHandler *users.UserHandler
}

// Utils contains all utility instances
type Utils struct {
	Validator *utils.Validator
	Responder *utils.Responder
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(
	logger interfaces.Logger,
	dbManager interfaces.DatabaseManager,
	cacheService interfaces.CacheService,
	userRepo interfaces.UserRepository,
	addressRepo interfaces.AddressRepository,
	roleRepo interfaces.RoleRepository,
	userRoleRepo interfaces.UserRoleRepository,
) (*HTTPServer, error) {
	// Create utils
	validator := utils.NewValidator(logger)
	responder := utils.NewResponder(logger)

	// Create services
	userService := services.NewUserService(userRepo, roleRepo, userRoleRepo, cacheService, logger, validator)
	addressService := services.NewAddressService(addressRepo, cacheService, logger, validator)
	roleService := services.NewRoleService(roleRepo, userRoleRepo, cacheService, logger, validator)

	// Create handlers
	userHandler := users.NewUserHandler(userService, validator, responder)

	// Create router
	router := gin.New()

	// Create server
	server := &HTTPServer{
		router:       router,
		logger:       logger,
		dbManager:    dbManager,
		cacheService: cacheService,
		services: &Services{
			UserService:    userService,
			AddressService: addressService,
			RoleService:    roleService,
		},
		handlers: &Handlers{
			UserHandler: userHandler,
		},
		utils: &Utils{
			Validator: validator,
			Responder: responder,
		},
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	// Create HTTP server
	server.server = &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

// setupMiddleware configures middleware for the server
func (s *HTTPServer) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Request ID middleware
	s.router.Use(middleware.RequestID())

	// Logging middleware
	s.router.Use(middleware.Logger(s.logger))

	// CORS middleware
	s.router.Use(middleware.CORS())

	// Rate limiting middleware
	s.router.Use(middleware.RateLimit())

	// Authentication middleware (for protected routes)
	// s.router.Use(middleware.Auth(s.authService))

	// Request timeout middleware
	s.router.Use(middleware.Timeout(30 * time.Second))
}

// setupRoutes configures all routes for the server
func (s *HTTPServer) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/ready", s.readyCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("/", s.handlers.UserHandler.CreateUser)
			users.GET("/:id", s.handlers.UserHandler.GetUserByID)
			users.PUT("/:id", s.handlers.UserHandler.UpdateUser)
			users.DELETE("/:id", s.handlers.UserHandler.DeleteUser)
			users.GET("/", s.handlers.UserHandler.ListUsers)
			users.GET("/search", s.handlers.UserHandler.SearchUsers)
			users.POST("/:id/validate", s.handlers.UserHandler.ValidateUser)
			users.POST("/:id/roles/:roleId", s.handlers.UserHandler.AssignRole)
			users.DELETE("/:id/roles/:roleId", s.handlers.UserHandler.RemoveRole)
		}

		// Address routes
		addresses := v1.Group("/addresses")
		{
			addresses.POST("/", s.createAddress)
			addresses.GET("/:id", s.getAddress)
			addresses.PUT("/:id", s.updateAddress)
			addresses.DELETE("/:id", s.deleteAddress)
			addresses.GET("/search", s.searchAddresses)
		}

		// Role routes
		roles := v1.Group("/roles")
		{
			roles.POST("/", s.createRole)
			roles.GET("/:id", s.getRole)
			roles.PUT("/:id", s.updateRole)
			roles.DELETE("/:id", s.deleteRole)
			roles.GET("/", s.listRoles)
		}
	}

	// Swagger documentation
	s.router.GET("/swagger/*any", s.swaggerHandler)
}

// healthCheck handles health check requests
func (s *HTTPServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "aaa-service",
		"version":   "1.0.0",
	})
}

// readyCheck handles readiness check requests
func (s *HTTPServer) readyCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check database connectivity
	if err := s.dbManager.HealthCheck(ctx); err != nil {
		s.logger.Error("Database health check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "aaa-service",
			"error":     "database connection failed",
		})
		return
	}

	// Check cache connectivity
	if !s.cacheService.Exists("health_check") {
		if err := s.cacheService.Set("health_check", "ok", 60); err != nil {
			s.logger.Error("Cache health check failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"service":   "aaa-service",
				"error":     "cache connection failed",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "aaa-service",
		"version":   "1.0.0",
	})
}

// Address handlers
func (s *HTTPServer) createAddress(c *gin.Context) {
	// Implementation for creating address
	s.utils.Responder.SendSuccess(c, http.StatusCreated, gin.H{"message": "Address created successfully"})
}

func (s *HTTPServer) getAddress(c *gin.Context) {
	// Implementation for getting address
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Address retrieved successfully"})
}

func (s *HTTPServer) updateAddress(c *gin.Context) {
	// Implementation for updating address
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Address updated successfully"})
}

func (s *HTTPServer) deleteAddress(c *gin.Context) {
	// Implementation for deleting address
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Address deleted successfully"})
}

func (s *HTTPServer) searchAddresses(c *gin.Context) {
	// Implementation for searching addresses
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Addresses searched successfully"})
}

// Role handlers
func (s *HTTPServer) createRole(c *gin.Context) {
	// Implementation for creating role
	s.utils.Responder.SendSuccess(c, http.StatusCreated, gin.H{"message": "Role created successfully"})
}

func (s *HTTPServer) getRole(c *gin.Context) {
	// Implementation for getting role
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Role retrieved successfully"})
}

func (s *HTTPServer) updateRole(c *gin.Context) {
	// Implementation for updating role
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Role updated successfully"})
}

func (s *HTTPServer) deleteRole(c *gin.Context) {
	// Implementation for deleting role
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

func (s *HTTPServer) listRoles(c *gin.Context) {
	// Implementation for listing roles
	s.utils.Responder.SendSuccess(c, http.StatusOK, gin.H{"message": "Roles listed successfully"})
}

// swaggerHandler handles Swagger documentation
func (s *HTTPServer) swaggerHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Swagger documentation endpoint",
		"url":     "/swagger/index.html",
	})
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}

// GetRouter returns the underlying router for testing
func (s *HTTPServer) GetRouter() *gin.Engine {
	return s.router
}

// GetServer returns the underlying HTTP server
func (s *HTTPServer) GetServer() *http.Server {
	return s.server
}
