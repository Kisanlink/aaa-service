package grpc_server

import (
	"context"
	"fmt"
	"net"
	"time"

	cfg "github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"github.com/Kisanlink/aaa-service/v2/internal/middleware"
	auditRepo "github.com/Kisanlink/aaa-service/v2/internal/repositories/audit"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"github.com/Kisanlink/aaa-service/v2/internal/services/catalog"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

// GRPCServer represents the gRPC server with authentication and authorization
type GRPCServer struct {
	server              *grpc.Server
	logger              *zap.Logger
	authService         *services.AuthService
	authzService        *services.AuthorizationService
	auditService        *services.AuditService
	userService         interfaces.UserService
	roleService         interfaces.RoleService
	cacheService        interfaces.CacheService
	userRepository      interfaces.UserRepository
	organizationService interfaces.OrganizationService
	addressService      interfaces.AddressService
	serviceRepository   interfaces.ServiceRepository
	dbManager           db.DBManager
	port                string
	listener            net.Listener
}

// GRPCServerConfig contains configuration for the gRPC server
type GRPCServerConfig struct {
	Port             string
	JWTSecret        string
	TokenExpiry      time.Duration
	RefreshExpiry    time.Duration
	EnableReflection bool
	EnableTLS        bool
	CertFile         string
	KeyFile          string
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(
	config *GRPCServerConfig,
	dbManager db.DBManager,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	userRoleRepository interfaces.UserRoleRepository,
	userRepository interfaces.UserRepository,
	cacheService interfaces.CacheService,
	organizationService interfaces.OrganizationService,
	addressService interfaces.AddressService,
	serviceRepository interfaces.ServiceRepository,
	logger *zap.Logger,
	validator interfaces.Validator,
) (*GRPCServer, error) {
	if config.Port == "" {
		config.Port = "50051"
	}

	// Create audit repository and service
	auditRepository := auditRepo.NewAuditRepository(dbManager)
	auditService := services.NewAuditService(dbManager, auditRepository, cacheService, logger)

	// Create authorization service
	// Get database connection for authorization service
	// For gRPC server, we need to get the GORM DB directly from the manager
	// The dbManager here is the interface, so we need to type assert to get the actual DB
	var gormDB *gorm.DB
	var err error

	// Try to get GORM DB from the manager if it's a PostgresManager
	if pm, ok := dbManager.(*db.PostgresManager); ok {
		gormDB, err = pm.GetDB(context.Background(), false)
		if err != nil {
			return nil, fmt.Errorf("failed to get database connection: %w", err)
		}
	} else {
		// Fallback: create a simple in-memory connection if not PostgreSQL
		return nil, fmt.Errorf("database manager is not PostgreSQL-based")
	}

	authzConfig := &services.AuthorizationServiceConfig{
		DB: gormDB,
	}
	authzService, err := services.NewAuthorizationService(cacheService, auditService, authzConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization service: %w", err)
	}

	// Create authentication service
	authConfig := &services.AuthServiceConfig{
		JWTSecret:     config.JWTSecret,
		TokenExpiry:   config.TokenExpiry,
		RefreshExpiry: config.RefreshExpiry,
	}
	jwtCfg := cfg.LoadJWTConfigFromEnv()
	if jwtCfg.Secret == "" {
		jwtCfg.Secret = config.JWTSecret
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
		return nil, fmt.Errorf("failed to create authentication service: %w", err)
	}

	return &GRPCServer{
		logger:              logger,
		authService:         authService,
		authzService:        authzService,
		auditService:        auditService,
		userService:         userService,
		roleService:         roleService,
		cacheService:        cacheService,
		userRepository:      userRepository,
		organizationService: organizationService,
		addressService:      addressService,
		serviceRepository:   serviceRepository,
		dbManager:           dbManager,
		port:                config.Port,
	}, nil
}

// Start starts the enhanced gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}
	s.listener = lis

	// Create gRPC server with interceptors
	// Build unified auth middleware for gRPC
	jwtCfg := cfg.LoadJWTConfigFromEnv()
	authMW := middleware.NewAuthMiddleware(s.authService, s.authzService, s.auditService, s.serviceRepository, s.logger, middleware.NewHS256Verifier(), jwtCfg)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			s.loggingInterceptor,
			authMW.GRPCAuthInterceptor(),
			s.auditInterceptor,
		),
	}

	s.server = grpc.NewServer(opts...)

	// Register services
	s.registerServices()

	// Enable reflection for development
	reflection.Register(s.server)

	s.logger.Info("Starting enhanced gRPC server", zap.String("port", s.port))

	// Start server
	go func() {
		if err := s.server.Serve(lis); err != nil {
			s.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() {
	s.logger.Info("Stopping enhanced gRPC server")
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Warn("Failed to close listener", zap.Error(err))
		}
	}
}

// registerServices registers all gRPC services
func (s *GRPCServer) registerServices() {
	// Register the unified AAA service handler
	// 	aaaHandler := NewAAAHandler(
	// 		s.authService,
	// 		s.authzService,
	// 		s.auditService,
	// 		s.userService,
	// 		s.roleService,
	// 		s.cacheService,
	// 		s.logger,
	// 	)
	// 	pb.RegisterAAAServiceServer(s.server, aaaHandler)

	// Keep legacy services for backward compatibility
	combinedHandler := NewCombinedUserHandler(s.authService, s.userService, s.logger)
	pb.RegisterUserServiceServer(s.server, combinedHandler)

	authzHandler := NewAuthorizationHandler(s.authzService, s.logger)
	pb.RegisterAuthorizationServiceServer(s.server, authzHandler)

	// Register TokenService for remote token validation
	tokenHandler := NewTokenHandler(
		s.authService,
		s.userService,
		s.authzService,
		s.cacheService,
		s.userRepository,
		s.logger,
	)
	pb.RegisterTokenServiceServer(s.server, tokenHandler)

	// Register OrganizationService for organization management
	orgHandler := NewOrganizationHandler(
		s.organizationService,
		s.logger,
	)
	pb.RegisterOrganizationServiceServer(s.server, orgHandler)

	// Register CatalogService for catalog management
	catalogService := catalog.NewCatalogService(s.dbManager, s.logger)

	// Create authorization checker for catalog operations
	catalogAuthChecker := NewAuthorizationChecker(s.authzService, s.logger)

	catalogHandler := NewCatalogHandler(catalogService, catalogAuthChecker, s.logger)
	pb.RegisterCatalogServiceServer(s.server, catalogHandler)

	// Register AddressService for address management (v1 - backward compatibility)
	addressHandler := NewAddressHandler(s.addressService, s.logger)
	pb.RegisterAddressServiceServer(s.server, addressHandler)

	// Register AddressService V2 for address management (v2 - Indian format, matches HTTP)
	addressHandlerV2 := NewAddressHandlerV2(s.addressService, s.logger)
	pbv2.RegisterAddressServiceServer(s.server, addressHandlerV2)

	s.logger.Info("gRPC services registered successfully",
		zap.String("primary_service", "AAAService"),
		zap.Strings("services", []string{"UserServiceV2", "AuthorizationService", "TokenService", "OrganizationService", "CatalogService", "AddressService", "AddressServiceV2"}))
}

// loggingInterceptor logs gRPC requests
func (s *GRPCServer) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// Call the handler
	resp, err := handler(ctx, req)

	// Log the request
	duration := time.Since(start)
	if err != nil {
		s.logger.Error("gRPC request failed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		s.logger.Info("gRPC request completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		)
	}

	return resp, err
}

// auditInterceptor logs gRPC requests for audit purposes
func (s *GRPCServer) auditInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Extract user ID from context (set by auth interceptor)
	userID := "unknown"
	if userIDValue := ctx.Value("user_id"); userIDValue != nil {
		if uid, ok := userIDValue.(string); ok {
			userID = uid
		}
	}

	// Call the handler
	resp, err := handler(ctx, req)

	// Log to audit service
	success := err == nil
	if s.auditService != nil {
		if err != nil {
			s.auditService.LogUserActionWithError(ctx, userID, "grpc_call", "grpc", info.FullMethod, err, map[string]interface{}{
				"method": info.FullMethod,
			})
		} else {
			s.auditService.LogUserAction(ctx, userID, "grpc_call", "grpc", info.FullMethod, map[string]interface{}{
				"method": info.FullMethod,
			})
		}
	}

	_ = success // Mark as used
	return resp, err
}

// GetServer returns the underlying gRPC server
func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}

// GetPort returns the server port
func (s *GRPCServer) GetPort() string {
	return s.port
}

// Health check method
func (s *GRPCServer) HealthCheck() bool {
	return s.server != nil && s.listener != nil
}
