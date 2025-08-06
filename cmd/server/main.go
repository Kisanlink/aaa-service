package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"sync"
	"syscall"
	"time"

	"github.com/Kisanlink/aaa-service/config"
	_ "github.com/Kisanlink/aaa-service/docs"
	"github.com/Kisanlink/aaa-service/grpc_server"
	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/middleware"
	addressRepo "github.com/Kisanlink/aaa-service/repositories/addresses"
	roleRepo "github.com/Kisanlink/aaa-service/repositories/roles"
	userRepo "github.com/Kisanlink/aaa-service/repositories/users"
	"github.com/Kisanlink/aaa-service/routes"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title AAA Service API
// @version 2.0
// @description Authentication, Authorization, and Accounting Service with SpiceDB integration
// @host localhost:8080
// @BasePath /api/v2
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// Server manages both HTTP and gRPC servers
type Server struct {
	httpServer *HTTPServer
	grpcServer *grpc_server.GRPCServer
	logger     *zap.Logger
}

// HTTPServer wraps the gin router with middleware
type HTTPServer struct {
	router *gin.Engine
	server *http.Server
	port   string
	logger *zap.Logger
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}()

	// Load configuration from environment
	httpPort := getEnv("HTTP_PORT", "8080")
	grpcPort := getEnv("GRPC_PORT", "50051")
	jwtSecret := getEnv("JWT_SECRET", "default-secret-key-change-in-production")
	spiceDBAddr := getEnv("SPICEDB_ADDR", "localhost:50051")
	spiceDBToken := getEnv("SPICEDB_TOKEN", "")

	if spiceDBToken == "" {
		logger.Warn("SpiceDB token not set, using empty token")
	}

	// Initialize database manager
	dbManager, err := config.NewDatabaseManager(logger)
	if err != nil {
		logger.Fatal("Failed to initialize database manager", zap.Error(err))
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			log.Printf("Failed to close database manager: %v", err)
		}
	}()

	// Initialize services and repositories
	server, err := initializeServer(
		httpPort, grpcPort, jwtSecret, spiceDBAddr, spiceDBToken,
		dbManager, logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	// Start servers
	logger.Info("Starting AAA service",
		zap.String("http_port", httpPort),
		zap.String("grpc_port", grpcPort))

	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start servers", zap.Error(err))
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")
	server.Stop()
}

// initializeServer initializes all services and creates the server
func initializeServer(
	httpPort, grpcPort, jwtSecret, spiceDBAddr, spiceDBToken string,
	dbManager *db.DatabaseManager,
	logger *zap.Logger,
) (*Server, error) {
	// Initialize logger adapter and utilities
	loggerAdapter := utils.NewLoggerAdapter(logger)
	validator := utils.NewValidator()
	responder := utils.NewResponder(loggerAdapter)

	// Get primary database manager
	primaryDBManager := dbManager.GetManager(dbManager.GetPostgresManager().GetBackendType())
	if primaryDBManager == nil {
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

	// Initialize business services
	_ = services.NewAddressService(addressRepository, cacheService, loggerAdapter, validator) // Available for future use
	roleService := services.NewRoleService(roleRepository, userRoleRepository, cacheService, loggerAdapter, validator)
	userService := services.NewUserService(userRepository, roleRepository, userRoleRepository, cacheService, logger, validator)

	// Initialize HTTP server
	httpServer, err := initializeHTTPServer(
		httpPort, jwtSecret, spiceDBAddr, spiceDBToken,
		primaryDBManager, userService, roleService, userRepository, userRoleRepository,
		cacheService, validator, responder, maintenanceService, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// Initialize gRPC server
	grpcConfig := &grpc_server.GRPCServerConfig{
		Port:             grpcPort,
		JWTSecret:        jwtSecret,
		TokenExpiry:      24 * time.Hour,
		RefreshExpiry:    7 * 24 * time.Hour,
		SpiceDBAddr:      spiceDBAddr,
		SpiceDBToken:     spiceDBToken,
		EnableReflection: true,
	}

	grpcServer, err := grpc_server.NewGRPCServer(
		grpcConfig, primaryDBManager, userService, roleService,
		userRoleRepository, userRepository, cacheService, logger, validator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	return &Server{
		httpServer: httpServer,
		grpcServer: grpcServer,
		logger:     logger,
	}, nil
}

// initializeHTTPServer creates and configures the HTTP server with middleware
func initializeHTTPServer(
	port, jwtSecret, spiceDBAddr, spiceDBToken string,
	dbManager db.DBManager,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	userRepository interfaces.UserRepository,
	userRoleRepository interfaces.UserRoleRepository,
	cacheService interfaces.CacheService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	maintenanceService interfaces.MaintenanceService,
	logger *zap.Logger,
) (*HTTPServer, error) {
	// Initialize enhanced services for middleware
	auditService := services.NewAuditService(dbManager, cacheService, logger)

	authzConfig := &services.AuthorizationServiceConfig{
		SpiceDBAddr:  spiceDBAddr,
		SpiceDBToken: spiceDBToken,
	}
	authzService, err := services.NewAuthorizationService(cacheService, auditService, authzConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization service: %w", err)
	}

	authConfig := &services.AuthServiceConfig{
		JWTSecret:     jwtSecret,
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		SpiceDBAddr:   spiceDBAddr,
		SpiceDBToken:  spiceDBToken,
	}
	authService, err := services.NewAuthService(
		userRepository, roleService, userRoleRepository,
		cacheService, authzService, auditService,
		authConfig, logger, validator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication service: %w", err)
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, authzService, auditService, logger)
	auditMiddleware := middleware.NewAuditMiddleware(auditService, logger)

	// Create gin router
	router := gin.New()

	// Setup middleware stack
	loggerAdapter := utils.NewLoggerAdapter(logger)
	routes.SetupMiddleware(router,
		cors.Default(),
		middleware.RequestID(),
		middleware.Logger(loggerAdapter),
		auditMiddleware.HTTPAuditMiddleware(),
		authMiddleware.HTTPAuthMiddleware(),
		authMiddleware.HTTPAuthzMiddleware(),
		middleware.ErrorHandler,
		middleware.PanicRecoveryHandler(loggerAdapter),
		middleware.MaintenanceMode(maintenanceService, responder, loggerAdapter),
	)

	// Setup routes using the routes package
	routes.SetupAAAWrapper(router, authService, authzService, auditService, authMiddleware, logger)

	// Add swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})

	return &HTTPServer{
		router: router,
		port:   port,
		logger: logger,
	}, nil
}

// Start starts both HTTP and gRPC servers
func (s *Server) Start() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.grpcServer.Start(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Check for immediate startup errors
	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		// Servers started successfully
		s.logger.Info("Both HTTP and gRPC servers started successfully")
		return nil
	}
}

// Stop gracefully stops both servers
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Stop HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.httpServer.Stop(ctx)
	}()

	// Stop gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.grpcServer.Stop()
	}()

	wg.Wait()
	s.logger.Info("All servers stopped gracefully")
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}

	s.logger.Info("Starting HTTP server", zap.String("port", s.port))
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server forced to shutdown", zap.Error(err))
		} else {
			s.logger.Info("HTTP server stopped gracefully")
		}
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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
