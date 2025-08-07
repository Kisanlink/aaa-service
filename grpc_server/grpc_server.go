package grpc_server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Kisanlink/aaa-service/interfaces"
	"github.com/Kisanlink/aaa-service/middleware"
	pb "github.com/Kisanlink/aaa-service/proto"
	"github.com/Kisanlink/aaa-service/services"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer represents the gRPC server with authentication and authorization
type GRPCServer struct {
	server       *grpc.Server
	logger       *zap.Logger
	authService  *services.AuthService
	authzService *services.AuthorizationService
	auditService *services.AuditService
	userService  interfaces.UserService
	roleService  interfaces.RoleService
	cacheService interfaces.CacheService
	dbManager    db.DBManager
	port         string
	listener     net.Listener
}

// GRPCServerConfig contains configuration for the gRPC server
type GRPCServerConfig struct {
	Port             string
	JWTSecret        string
	TokenExpiry      time.Duration
	RefreshExpiry    time.Duration
	SpiceDBAddr      string
	SpiceDBToken     string
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
	logger *zap.Logger,
	validator interfaces.Validator,
) (*GRPCServer, error) {
	if config.Port == "" {
		config.Port = "50051"
	}

	// Create audit service
	auditService := services.NewAuditService(dbManager, cacheService, logger)

	// Create authorization service
	authzConfig := &services.AuthorizationServiceConfig{
		SpiceDBAddr:  config.SpiceDBAddr,
		SpiceDBToken: config.SpiceDBToken,
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
		SpiceDBAddr:   config.SpiceDBAddr,
		SpiceDBToken:  config.SpiceDBToken,
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
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication service: %w", err)
	}

	return &GRPCServer{
		logger:       logger,
		authService:  authService,
		authzService: authzService,
		auditService: auditService,
		userService:  userService,
		roleService:  roleService,
		cacheService: cacheService,
		dbManager:    dbManager,
		port:         config.Port,
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
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			s.loggingInterceptor,
			middleware.AuthInterceptor,
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
		s.listener.Close()
	}
}

// registerServices registers all gRPC services
func (s *GRPCServer) registerServices() {
	// Register authentication service (UserServiceV2)
	authHandler := NewAuthHandler(s.authService, s.logger)
	pb.RegisterUserServiceV2Server(s.server, authHandler)

	// Authorization and audit handlers are available but not registered
	// until the corresponding proto services are defined
	//
	// For now, they can be accessed through the gRPC interceptors
	// or as separate service methods called from the auth handler

	_ = NewAuthorizationHandler(s.authzService, s.logger) // Available for future use
	_ = NewAuditHandler(s.auditService, s.logger)         // Available for future use

	s.logger.Info("gRPC services registered successfully")
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
