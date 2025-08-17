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

	"github.com/Kisanlink/aaa-service/internal/config"
	"github.com/Kisanlink/aaa-service/internal/entities/models"
	"github.com/Kisanlink/aaa-service/internal/grpc_server"
	"github.com/Kisanlink/aaa-service/internal/handlers/admin"
	"github.com/Kisanlink/aaa-service/internal/handlers/permissions"
	"github.com/Kisanlink/aaa-service/internal/handlers/roles"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/middleware"
	addressRepo "github.com/Kisanlink/aaa-service/internal/repositories/addresses"
	contactRepo "github.com/Kisanlink/aaa-service/internal/repositories/contacts"
	roleRepo "github.com/Kisanlink/aaa-service/internal/repositories/roles"
	userRepo "github.com/Kisanlink/aaa-service/internal/repositories/users"
	"github.com/Kisanlink/aaa-service/internal/routes"
	"github.com/Kisanlink/aaa-service/internal/services"
	contactService "github.com/Kisanlink/aaa-service/internal/services/contacts"
	"github.com/Kisanlink/aaa-service/internal/services/user"
	"github.com/Kisanlink/aaa-service/migrations"
	"github.com/Kisanlink/aaa-service/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// @title AAA Service API
// @version 2.0
// @description Authentication, Authorization, and Accounting Service with PostgreSQL-based RBAC
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// runSeedScripts runs all seeding scripts to initialize default data
func runSeedScripts(ctx context.Context, dbManager *db.DatabaseManager, logger *zap.Logger) error {
	logger.Info("🌱 Starting database seeding...")

	// 1. Seed static actions
	if postgresManager := dbManager.GetPostgresManager(); postgresManager != nil {
		if _, err := postgresManager.GetDB(ctx, false); err != nil {
			return fmt.Errorf("failed to get postgres DB for seeding: %w", err)
		}

		if err := migrations.SeedStaticActionsWithDBManager(ctx, dbManager, logger); err != nil {
			return fmt.Errorf("failed to seed static actions: %w", err)
		}

		// Seed core resources, roles, and permissions using PostgreSQL
		if err := migrations.SeedCoreResourcesRolesPermissionsWithDBManager(ctx, dbManager, logger); err != nil {
			return fmt.Errorf("failed to seed core roles/permissions: %w", err)
		}
	}

	// 2. Seed default roles
	logger.Info("🔧 Creating default roles...")
	defaultRoles := []struct {
		name        string
		description string
		scope       models.RoleScope
	}{
		{"super_admin", "Super Administrator with global access", models.RoleScopeGlobal},
		{"admin", "Administrator with organization-level access", models.RoleScopeOrg},
		{"user", "Regular user with basic access", models.RoleScopeOrg},
		{"viewer", "Read-only access user", models.RoleScopeOrg},
		{"aaa_admin", "AAA service administrator", models.RoleScopeGlobal},
		{"module_admin", "Module administrator for service management", models.RoleScopeOrg},
	}

	// Get the primary backend manager (explicitly request GORM backend)
	primaryManager := dbManager.GetManager(db.BackendGorm)
	if primaryManager == nil {
		logger.Warn("Primary database manager (GORM) not available; skipping role seeding")
		return nil
	}

	for _, roleData := range defaultRoles {
		// Check if role already exists by name
		filters := []base.FilterCondition{{Field: "name", Operator: base.OpEqual, Value: roleData.name}}
		var existing []models.Role
		filter := &base.Filter{
			Group: base.FilterGroup{
				Conditions: filters,
				Logic:      base.LogicAnd,
			},
		}
		if err := primaryManager.List(ctx, filter, &existing); err != nil {
			logger.Error("Failed to check existing role", zap.String("role", roleData.name), zap.Error(err))
			continue
		}
		if len(existing) > 0 {
			logger.Info("Role already exists", zap.String("role", roleData.name))
			continue
		}

		// Create new role
		role := models.NewRole(roleData.name, roleData.description, roleData.scope)
		if err := primaryManager.Create(ctx, role); err != nil {
			logger.Error("Failed to create role", zap.String("role", roleData.name), zap.Error(err))
		} else {
			logger.Info("Created role", zap.String("role", roleData.name), zap.String("id", role.ID))
		}
	}

	logger.Info("✅ Database seeding completed!")
	return nil
}

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

	// Optionally run database seeding scripts
	if getEnv("AAA_RUN_SEED", "true") == "true" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := runSeedScripts(ctx, dbManager, logger); err != nil {
			logger.Fatal("Failed to run database seeding scripts", zap.Error(err))
		}
	} else {
		logger.Info("Skipping database seeding; AAA_RUN_SEED is not true")
	}

	// Initialize ID counters from database to prevent duplicate key violations
	logger.Info("Initializing ID counters from database")
	primaryDBManager := dbManager.GetManager(db.BackendGorm)
	if primaryDBManager != nil {
		// Get the underlying GORM DB for counter initialization
		if postgresManager, ok := primaryDBManager.(*db.PostgresManager); ok {
			// Get the GORM database connection
			gormDB, err := postgresManager.GetDB(context.Background(), false)
			if err != nil {
				logger.Warn("Failed to get GORM DB connection, skipping counter initialization", zap.Error(err))
			} else {
				counterService := services.NewCounterInitializationService(gormDB, logger)
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := counterService.InitializeAllCounters(ctx); err != nil {
					logger.Warn("Failed to initialize ID counters, continuing with default values", zap.Error(err))
				} else {
					logger.Info("ID counters initialized successfully from database")
				}
			}
		} else {
			logger.Warn("Database manager is not a PostgresManager, skipping counter initialization")
		}
	} else {
		logger.Warn("No database manager available, skipping counter initialization")
	}

	// Initialize services and repositories
	server, err := initializeServer(
		httpPort, grpcPort, jwtSecret,
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
	httpPort, grpcPort, jwtSecret string,
	dbManager *db.DatabaseManager,
	logger *zap.Logger,
) (*Server, error) {
	// Initialize logger adapter and utilities
	loggerAdapter := utils.NewLoggerAdapter(logger)
	validator := utils.NewValidator()
	responder := utils.NewResponder(loggerAdapter)

	// Get primary database manager
	primaryDBManager := dbManager.GetManager(db.BackendGorm)
	if primaryDBManager == nil {
		logger.Fatal("No database manager available, exiting")
	}

	// Initialize repositories with the DatabaseManager for advanced operations
	userRepository := userRepo.NewUserRepository(primaryDBManager)
	addressRepository := addressRepo.NewAddressRepository(primaryDBManager)
	roleRepository := roleRepo.NewRoleRepository(primaryDBManager)
	userRoleRepository := roleRepo.NewUserRoleRepository(primaryDBManager)

	// Initialize cache service
	// FIX: Check CACHE_DISABLED environment variable to optionally disable Redis
	var cacheService interfaces.CacheService
	if getEnv("CACHE_DISABLED", "false") == "true" {
		logger.Info("Cache disabled by CACHE_DISABLED=true, using no-op cache service")
		cacheService = services.NewNoOpCacheService(loggerAdapter)
	} else {
		// Initialize cache service from environment variables (fallback to defaults)
		redisHost := getEnv("REDIS_HOST", "localhost")
		redisPort := getEnv("REDIS_PORT", "6379")
		redisPassword := getEnv("REDIS_PASSWORD", "")
		redisDB := 0 // Could be made configurable via REDIS_DB env var
		logger.Info("Initializing Redis cache service",
			zap.String("host", redisHost),
			zap.String("port", redisPort))
		cacheService = services.NewCacheService(redisHost+":"+redisPort, redisPassword, redisDB, loggerAdapter)
	}

	// Initialize maintenance service
	maintenanceService := services.NewMaintenanceService(cacheService, loggerAdapter)

	// Initialize business services
	_ = services.NewAddressService(addressRepository, cacheService, loggerAdapter, validator) // Available for future use
	roleService := services.NewRoleService(roleRepository, userRoleRepository, cacheService, loggerAdapter, validator)
	userService := user.NewService(userRepository, roleRepository, userRoleRepository, cacheService, logger, validator)

	// Initialize contact service
	contactRepository := contactRepo.NewContactRepository(primaryDBManager)
	contactServiceInstance := contactService.NewContactService(contactRepository, cacheService, loggerAdapter, validator)

	// Initialize handlers
	roleHandler := roles.NewRoleHandler(roleService, validator, responder, logger)
	permissionHandler := permissions.NewPermissionHandler(primaryDBManager, validator, responder, logger)

	// Initialize HTTP server
	httpServer, err := initializeHTTPServer(
		httpPort, jwtSecret,
		primaryDBManager, userService, roleService, userRepository, userRoleRepository,
		cacheService, validator, responder, maintenanceService, logger, roleHandler, permissionHandler, contactServiceInstance,
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
	port, jwtSecret string,
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
	roleHandler *roles.RoleHandler,
	permissionHandler *permissions.PermissionHandler,
	contactServiceInstance *contactService.ContactService,
) (*HTTPServer, error) {
	// Build auth, authorization, and audit stack
	auditService, authzService, authService, authMiddleware, auditMiddleware, err := setupAuthStack(
		context.Background(),
		dbManager,
		cacheService,
		userRepository,
		roleService,
		userRoleRepository,
		jwtSecret,
		logger,
		validator,
	)
	if err != nil {
		return nil, err
	}

	// Create gin router
	router := gin.New()

	// Setup middleware stack
	setupHTTPMiddleware(router, authMiddleware, auditMiddleware, maintenanceService, responder, logger)

	// Setup routes and docs
	setupRoutesAndDocs(router, authService, authzService, auditService, authMiddleware, maintenanceService, validator, responder, logger, roleHandler, permissionHandler, userService, roleService, contactServiceInstance)

	return &HTTPServer{
		router: router,
		port:   port,
		logger: logger,
	}, nil
}

// setupAuthStack initializes audit, authorization, and authentication services, and their middlewares
func setupAuthStack(
	ctx context.Context,
	dbManager db.DBManager,
	cacheService interfaces.CacheService,
	userRepository interfaces.UserRepository,
	roleService interfaces.RoleService,
	userRoleRepository interfaces.UserRoleRepository,
	jwtSecret string,
	logger *zap.Logger,
	validator interfaces.Validator,
) (*services.AuditService, *services.AuthorizationService, *services.AuthService, *middleware.AuthMiddleware, *middleware.AuditMiddleware, error) {
	auditService := services.NewAuditService(dbManager, cacheService, logger)

	// Get database connection for authorization service
	// Use the dbManager interface directly instead of type assertion
	var gormDB *gorm.DB
	var err error

	// Get GORM DB from the manager if it's available
	if pm, ok := dbManager.(*db.PostgresManager); ok {
		gormDB, err = pm.GetDB(ctx, false)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to get database connection: %w", err)
		}
	} else {
		// Fallback: create a simple in-memory connection if not PostgreSQL
		return nil, nil, nil, nil, nil, fmt.Errorf("database manager is not PostgreSQL-based")
	}

	authzConfig := &services.AuthorizationServiceConfig{
		DB: gormDB,
	}
	authzService, err := services.NewAuthorizationService(cacheService, auditService, authzConfig, logger)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create authorization service: %w", err)
	}

	authConfig := &services.AuthServiceConfig{
		JWTSecret:     jwtSecret,
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	jwtCfg := config.LoadJWTConfigFromEnv()
	// Ensure secret matches legacy env fallbacks
	if jwtCfg.Secret == "" {
		jwtCfg.Secret = jwtSecret
	}
	authService, err := services.NewAuthService(
		userRepository,
		roleService,
		userRoleRepository,
		cacheService,
		authzService,
		auditService,
		authConfig,
		logger,
		validator,
		jwtCfg,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create authentication service: %w", err)
	}

	authMiddleware := middleware.NewAuthMiddleware(authService, authzService, auditService, logger, middleware.NewHS256Verifier(), jwtCfg)
	auditMiddleware := middleware.NewAuditMiddleware(auditService, logger)
	return auditService, authzService, authService, authMiddleware, auditMiddleware, nil
}

// setupHTTPMiddleware configures the middleware stack on the provided router
func setupHTTPMiddleware(
	router *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	auditMiddleware *middleware.AuditMiddleware,
	maintenanceService interfaces.MaintenanceService,
	responder interfaces.Responder,
	logger *zap.Logger,
) {
	loggerAdapter := utils.NewLoggerAdapter(logger)
	routes.SetupMiddleware(router,
		cors.Default(),
		middleware.RequestID(),
		middleware.Logger(loggerAdapter),
		middleware.ResponseContextHeaders(),
		auditMiddleware.HTTPAuditMiddleware(),
		authMiddleware.HTTPAuthMiddleware(),
		authMiddleware.HTTPAuthzMiddleware(),
		middleware.ErrorHandler,
		middleware.PanicRecoveryHandler(loggerAdapter),
		middleware.MaintenanceMode(maintenanceService, responder, loggerAdapter),
	)
}

// setupRoutesAndDocs registers API routes and documentation endpoints
func setupRoutesAndDocs(
	router *gin.Engine,
	authService *services.AuthService,
	authzService *services.AuthorizationService,
	auditService *services.AuditService,
	authMiddleware *middleware.AuthMiddleware,
	maintenanceService interfaces.MaintenanceService,
	validator interfaces.Validator,
	responder interfaces.Responder,
	logger *zap.Logger,
	roleHandler *roles.RoleHandler,
	permissionHandler *permissions.PermissionHandler,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	contactServiceInstance *contactService.ContactService,
) {
	// Create AdminHandler for v2 admin routes
	adminHandler := admin.NewAdminHandler(maintenanceService, validator, responder, logger)

	// Setup routes using the routes package with AdminHandler
	routes.SetupAAAWithAdmin(router, authService, authzService, auditService, authMiddleware, adminHandler, roleHandler, permissionHandler, userService, roleService, contactServiceInstance, validator, responder, logger)

	// Serve OpenAPI and Scalar-powered docs UI (gated by env)
	if getEnv("AAA_ENABLE_DOCS", "true") == "true" {
		router.StaticFile("/docs/swagger.json", "docs/swagger.json")
		router.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")
		router.GET("/docs", func(c *gin.Context) {
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
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/docs")
		})
	} else {
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "aaa-service running")
		})
	}
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
func (m *InMemoryDBManager) List(ctx context.Context, filter *base.Filter, model interface{}) error {
	return nil
}

func (m *InMemoryDBManager) Count(ctx context.Context, filter *base.Filter, model interface{}) (int64, error) {
	return 0, nil
}

func (m *InMemoryDBManager) AutoMigrateModels(ctx context.Context, models ...interface{}) error {
	// In-memory database doesn't support schema migration
	return nil
}
