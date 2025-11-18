package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/grpc_server"
	actionHandlers "github.com/Kisanlink/aaa-service/v2/internal/handlers/actions"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/admin"
	kycHandlers "github.com/Kisanlink/aaa-service/v2/internal/handlers/kyc"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/permissions"
	principalHandlers "github.com/Kisanlink/aaa-service/v2/internal/handlers/principals"
	resourceHandlers "github.com/Kisanlink/aaa-service/v2/internal/handlers/resources"
	"github.com/Kisanlink/aaa-service/v2/internal/handlers/roles"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	actionRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/actions"
	repositoryAdapters "github.com/Kisanlink/aaa-service/v2/internal/repositories/adapters"
	addressRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/addresses"
	auditRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/audit"
	contactRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/contacts"
	groupRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/groups"
	kycRepositories "github.com/Kisanlink/aaa-service/v2/internal/repositories/kyc"
	organizationRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/organizations"
	permissionRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/permissions"
	principalRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/principals"
	resourcePermRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/resource_permissions"
	resourceRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/resources"
	rolePermRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/role_permissions"
	roleRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/roles"
	userRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/users"
	"github.com/Kisanlink/aaa-service/v2/internal/routes"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	actionService "github.com/Kisanlink/aaa-service/v2/internal/services/actions"
	serviceAdapters "github.com/Kisanlink/aaa-service/v2/internal/services/adapters"
	contactService "github.com/Kisanlink/aaa-service/v2/internal/services/contacts"
	groupService "github.com/Kisanlink/aaa-service/v2/internal/services/groups"
	kycServices "github.com/Kisanlink/aaa-service/v2/internal/services/kyc"
	organizationService "github.com/Kisanlink/aaa-service/v2/internal/services/organizations"
	permissionService "github.com/Kisanlink/aaa-service/v2/internal/services/permissions"
	principalService "github.com/Kisanlink/aaa-service/v2/internal/services/principals"
	resourceService "github.com/Kisanlink/aaa-service/v2/internal/services/resources"
	roleAssignmentService "github.com/Kisanlink/aaa-service/v2/internal/services/role_assignments"
	"github.com/Kisanlink/aaa-service/v2/internal/services/user"
	"github.com/Kisanlink/aaa-service/v2/migrations"
	"github.com/Kisanlink/aaa-service/v2/utils"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//	@title						AAA Service API
//	@version					2.0
//	@description				Authentication, Authorization, and Accounting Service with PostgreSQL-based RBAC
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				API key for service-to-service authentication

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

		// Seed comprehensive RBAC resources and permissions
		if err := migrations.SeedComprehensiveRBACWithDBManager(ctx, dbManager, logger); err != nil {
			return fmt.Errorf("failed to seed comprehensive RBAC: %w", err)
		}

		// Seed wildcard permissions for super_admin role
		// This ensures super_admin has explicit wildcard resource_permissions for all resource types
		if err := migrations.SeedSuperAdminWildcardPermissionsWithDBManager(ctx, dbManager, logger); err != nil {
			return fmt.Errorf("failed to seed super_admin wildcard permissions: %w", err)
		}

		// Add performance indexes after all migrations complete
		logger.Info("🔧 Creating performance indexes for optimal query performance...")
		pm := dbManager.GetPostgresManager()
		if pm != nil {
			gormDB, err := pm.GetDB(ctx, false)
			if err != nil {
				logger.Warn("Failed to get GORM DB for index creation", zap.Error(err))
			} else {
				if err := migrations.AddPerformanceIndexes(ctx, gormDB, logger); err != nil {
					logger.Warn("Failed to create some performance indexes", zap.Error(err))
					// Don't fail startup - indexes can be added manually
				} else {
					logger.Info("✅ Performance indexes created successfully")
				}

				// Run hierarchy fields migration
				if err := migrations.AddHierarchyFields(ctx, gormDB, logger); err != nil {
					logger.Warn("Failed to add hierarchy fields and indexes", zap.Error(err))
					// Don't fail startup - migration can be run manually
				} else {
					logger.Info("✅ Hierarchy fields and indexes created successfully")
				}

				// Validate hierarchy migration
				if err := migrations.ValidateHierarchyMigration(ctx, gormDB, logger); err != nil {
					logger.Warn("Hierarchy migration validation issues", zap.Error(err))
				}
			}
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
	router                      *gin.Engine
	server                      *http.Server
	port                        string
	logger                      *zap.Logger
	organizationServiceInstance interfaces.OrganizationService
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
	userProfileRepository := userRepo.NewUserProfileRepository(primaryDBManager)
	addressRepository := addressRepo.NewAddressRepository(primaryDBManager)
	roleRepository := roleRepo.NewRoleRepository(primaryDBManager)
	userRoleRepository := roleRepo.NewUserRoleRepository(primaryDBManager)

	// Initialize organization and group repositories
	organizationRepository := organizationRepo.NewOrganizationRepository(primaryDBManager)
	groupRepository := groupRepo.NewGroupRepository(primaryDBManager)
	groupRoleRepository := groupRepo.NewGroupRoleRepository(primaryDBManager)
	groupMembershipRepository := groupRepo.NewGroupMembershipRepository(primaryDBManager)

	// Initialize principal repositories for service authentication
	serviceRepository := principalRepo.NewServiceRepository(primaryDBManager)

	// Initialize RBAC repositories for fine-grained permissions
	resourceRepository := resourceRepo.NewResourceRepository(primaryDBManager)
	actionRepository := actionRepo.NewActionRepository(primaryDBManager)
	permissionRepository := permissionRepo.NewPermissionRepository(primaryDBManager)
	rolePermissionRepository := rolePermRepo.NewRolePermissionRepository(primaryDBManager)
	resourcePermissionRepository := resourcePermRepo.NewResourcePermissionRepository(primaryDBManager)

	// Initialize S3Manager for photo/document storage (optional, configured via AWS env vars)
	// The S3Manager from kisanlink-db handles AWS S3 operations
	// AWS credentials are picked up from environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
	var s3Manager *db.S3Manager
	awsBucket := getEnv("AWS_S3_BUCKET", "")
	if awsBucket != "" {
		s3Config := &db.Config{
			S3Bucket: awsBucket,
			S3Region: getEnv("AWS_REGION", "us-east-1"),
		}
		s3Manager = db.NewS3Manager(s3Config, logger)
		if err := s3Manager.Connect(context.Background()); err != nil {
			logger.Warn("Failed to connect to S3, photo upload features will not work",
				zap.Error(err),
				zap.String("bucket", awsBucket))
			s3Manager = nil
		} else {
			logger.Info("S3Manager initialized successfully",
				zap.String("bucket", awsBucket),
				zap.String("region", getEnv("AWS_REGION", "us-east-1")))
		}
	} else {
		logger.Info("AWS_S3_BUCKET not configured, photo upload features will be disabled")
	}

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
		redisTLSEnabled := getEnv("REDIS_TLS_ENABLED", "false") == "true"
		redisDB := 0 // Could be made configurable via REDIS_DB env var

		// Parse timeout values from environment (with sensible defaults)
		dialTimeout := parseDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second)
		readTimeout := parseDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second)
		writeTimeout := parseDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second)
		poolSize := parseIntEnv("REDIS_POOL_SIZE", 10)
		minIdleConns := parseIntEnv("REDIS_MIN_IDLE_CONNS", 2)

		config := services.RedisConfig{
			Addr:         redisHost + ":" + redisPort,
			Password:     redisPassword,
			DB:           redisDB,
			TLSEnabled:   redisTLSEnabled,
			DialTimeout:  dialTimeout,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		}

		logger.Info("Initializing Redis cache service",
			zap.String("host", redisHost),
			zap.String("port", redisPort),
			zap.Bool("tls_enabled", redisTLSEnabled),
			zap.Duration("dial_timeout", dialTimeout),
			zap.Duration("read_timeout", readTimeout),
			zap.Duration("write_timeout", writeTimeout))
		cacheService = services.NewCacheService(config, loggerAdapter)
	}

	// Initialize maintenance service
	maintenanceService := services.NewMaintenanceService(cacheService, loggerAdapter)

	// Initialize business services
	addressService := services.NewAddressService(addressRepository, cacheService, loggerAdapter, validator)
	roleService := services.NewRoleService(roleRepository, userRoleRepository, cacheService, loggerAdapter, validator)
	userServiceInstance := user.NewService(userRepository, roleRepository, userRoleRepository, cacheService, logger, validator)

	// Inject organizational repositories for JWT context enhancement
	if svc, ok := userServiceInstance.(*user.Service); ok {
		svc.SetOrganizationalRepositories(groupMembershipRepository, groupRepository, organizationRepository)
	}

	// Initialize role inheritance engine for group-based role inheritance
	roleInheritanceEngine := groupService.NewRoleInheritanceEngineWithRepos(
		groupRepository,
		groupRoleRepository,
		roleRepository,
		groupMembershipRepository,
		cacheService,
		logger,
	)

	// Inject role inheritance engine into user service
	if svc, ok := userServiceInstance.(*user.Service); ok {
		svc.SetRoleInheritanceEngine(roleInheritanceEngine)
	}

	userService := userServiceInstance

	// Initialize contact service
	contactRepository := contactRepo.NewContactRepository(primaryDBManager)
	contactServiceInstance := contactService.NewContactService(contactRepository, cacheService, loggerAdapter, validator)

	// Initialize principal service
	principalRepository := principalRepo.NewPrincipalRepository(primaryDBManager)
	principalService := principalService.NewPrincipalService(
		organizationRepository,
		principalRepository,
		serviceRepository,
		validator,
		logger,
	)

	// Initialize audit service early for RBAC services to use
	auditRepository := auditRepo.NewAuditRepository(primaryDBManager)
	auditServiceConcrete := services.NewAuditService(primaryDBManager, auditRepository, cacheService, logger)
	auditServiceAdapter := serviceAdapters.NewAuditServiceAdapter(auditServiceConcrete)

	// Initialize RBAC services with proper dependencies
	resourceService := resourceService.NewService(
		resourceRepository,
		cacheService,
		logger,
	)

	actionService := actionService.NewActionService(
		actionRepository,
		cacheService,
		loggerAdapter,
		validator,
	)

	permissionService := permissionService.NewService(
		permissionRepository,
		rolePermissionRepository,
		resourcePermissionRepository,
		roleRepository,
		cacheService,
		auditServiceAdapter,
		loggerAdapter,
	)

	roleAssignmentService := roleAssignmentService.NewService(
		roleRepository,
		rolePermissionRepository,
		resourcePermissionRepository,
		permissionRepository,
		auditRepository,
		cacheService,
		auditServiceAdapter,
		loggerAdapter,
	)

	// Initialize KYC service and dependencies
	// Sandbox API client for Aadhaar verification
	sandboxBaseURL := getEnv("AADHAAR_SANDBOX_URL", "")
	sandboxAPIKey := getEnv("AADHAAR_SANDBOX_API_KEY", "")
	sandboxAPISecret := getEnv("AADHAAR_SANDBOX_API_SECRET", "")

	if sandboxBaseURL == "" {
		logger.Warn("AADHAAR_SANDBOX_URL not configured, Aadhaar verification will not work")
	}

	sandboxClient := kycServices.NewSandboxClient(sandboxBaseURL, sandboxAPIKey, sandboxAPISecret, logger)

	// Create Aadhaar verification repository
	aadhaarRepo := kycRepositories.NewAadhaarVerificationRepository(primaryDBManager, s3Manager, logger)

	// Create user service adapter for KYC operations
	kycUserServiceAdapter := kycServices.NewUserServiceAdapter(userRepository, userProfileRepository, logger)

	// Create address service adapter (wrapping existing address service)
	kycAddressServiceAdapter := serviceAdapters.NewAddressServiceAdapter(addressService)

	// KYC service configuration
	kycConfig := &kycServices.Config{
		OTPExpirationSeconds: parseIntEnv("OTP_EXPIRATION_SECONDS", 300),
		OTPMaxAttempts:       parseIntEnv("OTP_MAX_ATTEMPTS", 3),
		OTPCooldownSeconds:   parseIntEnv("OTP_COOLDOWN_SECONDS", 60),
		PhotoMaxSizeMB:       parseIntEnv("PHOTO_MAX_SIZE_MB", 5),
	}

	// Create KYC service with all dependencies
	kycService := kycServices.NewService(
		aadhaarRepo,
		kycUserServiceAdapter,
		kycAddressServiceAdapter,
		sandboxClient,
		auditServiceAdapter,
		logger,
		kycConfig,
	)

	// Initialize handlers
	permissionHandler := permissions.NewPermissionHandler(permissionService, roleAssignmentService, validator, responder, logger)
	resourceHandler := resourceHandlers.NewResourceHandler(resourceService, validator, responder, logger)
	actionHandler := actionHandlers.NewActionHandler(actionService, validator, responder, logger)
	principalHandler := principalHandlers.NewPrincipalHandler(principalService, responder, logger)
	kycHandler := kycHandlers.NewHandler(kycService, validator, responder, logger)

	// Initialize HTTP server
	httpServer, err := initializeHTTPServer(
		httpPort, jwtSecret,
		primaryDBManager, userService, roleService, userRepository, userRoleRepository,
		cacheService, validator, responder, maintenanceService, logger, permissionHandler, resourceHandler, actionHandler, principalHandler, kycHandler, contactServiceInstance, addressService,
		organizationRepository, groupRepository, groupRoleRepository, groupMembershipRepository, roleRepository,
		serviceRepository,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// Initialize gRPC server using organizationServiceInstance from httpServer
	grpcConfig := &grpc_server.GRPCServerConfig{
		Port:             grpcPort,
		JWTSecret:        jwtSecret,
		TokenExpiry:      24 * time.Hour,
		RefreshExpiry:    7 * 24 * time.Hour,
		EnableReflection: true,
	}

	grpcServer, err := grpc_server.NewGRPCServer(
		grpcConfig, primaryDBManager, userService, roleService,
		userRoleRepository, userRepository, cacheService, httpServer.organizationServiceInstance,
		addressService, serviceRepository, logger, validator,
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
	permissionHandler *permissions.PermissionHandler,
	resourceHandler *resourceHandlers.ResourceHandler,
	actionHandler *actionHandlers.ActionHandler,
	principalHandler *principalHandlers.Handler,
	kycHandler *kycHandlers.Handler,
	contactServiceInstance *contactService.ContactService,
	addressService interfaces.AddressService,
	organizationRepository *organizationRepo.OrganizationRepository,
	groupRepository *groupRepo.GroupRepository,
	groupRoleRepository *groupRepo.GroupRoleRepository,
	groupMembershipRepository *groupRepo.GroupMembershipRepository,
	roleRepository *roleRepo.RoleRepository,
	serviceRepository interfaces.ServiceRepository,
) (*HTTPServer, error) {
	// Build auth, authorization, and audit stack
	auditService, authzService, authService, authMiddleware, auditMiddleware, err := setupAuthStack(
		context.Background(),
		dbManager,
		cacheService,
		userRepository,
		roleService,
		userRoleRepository,
		serviceRepository,
		jwtSecret,
		logger,
		validator,
	)
	if err != nil {
		return nil, err
	}

	// Initialize roleHandler now that auditService is available
	roleHandler := roles.NewRoleHandler(roleService, validator, responder, auditService, logger)

	// Create repository adapters to handle interface mismatches
	userRepositoryAdapter := repositoryAdapters.NewUserRepositoryAdapter(userRepository.(*userRepo.UserRepository))
	groupRepositoryAdapter := repositoryAdapters.NewGroupRepositoryAdapter(groupRepository)

	// Create audit service adapter
	auditRepository := auditRepo.NewAuditRepository(dbManager)
	auditServiceConcrete := services.NewAuditService(dbManager, auditRepository, cacheService, logger)
	auditServiceAdapter := serviceAdapters.NewAuditServiceAdapter(auditServiceConcrete)

	// Initialize organization service with adapters
	organizationServiceConcrete := organizationService.NewOrganizationService(
		organizationRepository,
		userRepositoryAdapter,
		groupRepositoryAdapter,
		nil, // groupService will be set after to avoid circular dependency
		validator,
		cacheService,
		auditServiceAdapter,
		logger,
	)
	organizationServiceInstance := organizationService.NewServiceAdapter(organizationServiceConcrete, logger)

	// Initialize group service with adapters
	groupServiceConcrete := groupService.NewGroupService(
		groupRepository,
		groupRoleRepository,
		groupMembershipRepository,
		organizationRepository,
		roleRepository,
		userRoleRepository,
		validator,
		cacheService,
		auditServiceAdapter,
		logger,
	)
	// Inject user service for cache invalidation
	groupServiceConcrete.SetUserService(userService)
	groupServiceInstance := groupServiceConcrete

	// Inject group service into organization service to resolve circular dependency
	organizationServiceConcrete.SetGroupService(groupServiceInstance)

	// Create gin router
	router := gin.New()

	// Setup middleware stack
	setupHTTPMiddleware(router, authMiddleware, auditMiddleware, maintenanceService, responder, logger)

	// Setup routes and docs
	setupRoutesAndDocs(router, authService, authzService, auditService, authMiddleware, maintenanceService, validator, responder, logger, roleHandler, permissionHandler, resourceHandler, actionHandler, principalHandler, kycHandler, userService, roleService, contactServiceInstance, addressService, organizationServiceInstance, groupServiceInstance)

	return &HTTPServer{
		router:                      router,
		port:                        port,
		logger:                      logger,
		organizationServiceInstance: organizationServiceInstance,
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
	serviceRepository interfaces.ServiceRepository,
	jwtSecret string,
	logger *zap.Logger,
	validator interfaces.Validator,
) (*services.AuditService, *services.AuthorizationService, *services.AuthService, *middleware.AuthMiddleware, *middleware.AuditMiddleware, error) {
	// Initialize audit repository
	auditRepository := auditRepo.NewAuditRepository(dbManager)
	auditService := services.NewAuditService(dbManager, auditRepository, cacheService, logger)

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

	authMiddleware := middleware.NewAuthMiddleware(authService, authzService, auditService, serviceRepository, logger, middleware.NewHS256Verifier(), jwtCfg)
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
		middleware.CORS(), // Use our custom CORS middleware instead of cors.Default()
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
	resourceHandler *resourceHandlers.ResourceHandler,
	actionHandler *actionHandlers.ActionHandler,
	principalHandler *principalHandlers.Handler,
	kycHandler *kycHandlers.Handler,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	contactServiceInstance *contactService.ContactService,
	addressService interfaces.AddressService,
	organizationServiceInstance interfaces.OrganizationService,
	groupServiceInstance interfaces.GroupService,
) {
	// Create AdminHandler for v2 admin routes
	adminHandler := admin.NewAdminHandler(maintenanceService, validator, responder, logger)

	// Setup routes using the enhanced wrapper that supports organization services
	routes.SetupAAAWithOrganizations(
		router,
		authService,
		authzService,
		auditService,
		authMiddleware,
		adminHandler,
		roleHandler,
		permissionHandler,
		userService,
		roleService,
		contactServiceInstance,
		addressService,
		organizationServiceInstance,
		groupServiceInstance,
		validator,
		responder,
		logger,
	)

	// Register principal and service management routes
	routes.RegisterPrincipalRoutes(router, principalHandler, authMiddleware)

	// Register RBAC resource routes
	routes.RegisterResourceRoutes(router, resourceHandler, authMiddleware)

	// Register RBAC action routes
	routes.RegisterActionRoutes(router.Group("/api/v1"), actionHandler)

	// Register KYC routes
	kycHandlers.RegisterRoutes(router, kycHandler, authMiddleware.HTTPAuthMiddleware())

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

		// Add favicon route to avoid 401 errors
		router.GET("/favicon.ico", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	} else {
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "aaa-service running")
		})
		logger.Info("Documentation endpoints disabled (AAA_ENABLE_DOCS=false)")
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

// parseDurationEnv parses a duration from environment variable (in seconds) with fallback
func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// parseIntEnv parses an integer from environment variable with fallback
func parseIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
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
