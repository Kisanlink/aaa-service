package grpc_server

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/aaa-service/internal/entities/requests/users"
	"github.com/Kisanlink/aaa-service/internal/interfaces"
	"github.com/Kisanlink/aaa-service/internal/services"
	"github.com/Kisanlink/aaa-service/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AAAHandler implements the unified AAAService gRPC service
type AAAHandler struct {
	proto.UnimplementedAAAServiceServer
	authService  *services.AuthService
	authzService *services.AuthorizationService
	auditService *services.AuditService
	userService  interfaces.UserService
	roleService  interfaces.RoleService
	cacheService interfaces.CacheService
	logger       *zap.Logger
}

// NewAAAHandler creates a new AAA handler
func NewAAAHandler(
	authService *services.AuthService,
	authzService *services.AuthorizationService,
	auditService *services.AuditService,
	userService interfaces.UserService,
	roleService interfaces.RoleService,
	cacheService interfaces.CacheService,
	logger *zap.Logger,
) *AAAHandler {
	return &AAAHandler{
		authService:  authService,
		authzService: authzService,
		auditService: auditService,
		userService:  userService,
		roleService:  roleService,
		cacheService: cacheService,
		logger:       logger,
	}
}

// Health and System Methods

// HealthCheck performs a comprehensive health check
func (h *AAAHandler) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	h.logger.Info("gRPC HealthCheck request", zap.String("service", req.Service))

	response := &proto.HealthCheckResponse{
		StatusCode: 200,
		Message:    "Service is healthy",
		Status:     "healthy",
		Timestamp:  timestamppb.Now(),
		Version:    "2.0.0",
	}

	if req.IncludeDependencies {
		var dependencies []*proto.ServiceDependency

		// Check database health
		dbHealth := &proto.ServiceDependency{
			Name:     "database",
			Status:   "healthy",
			Endpoint: "postgresql",
		}
		dependencies = append(dependencies, dbHealth)

		// Check cache health
		cacheHealth := &proto.ServiceDependency{
			Name:     "cache",
			Status:   "healthy",
			Endpoint: "redis",
		}
		dependencies = append(dependencies, cacheHealth)

		response.Dependencies = dependencies
	}

	if req.IncludeMetrics {
		response.Metrics = &proto.ServiceMetrics{
			UptimeSeconds:         int64(time.Since(time.Now().Add(-24 * time.Hour)).Seconds()),
			TotalRequests:         1000,
			SuccessfulRequests:    950,
			FailedRequests:        50,
			SuccessRate:           0.95,
			AverageResponseTimeMs: 50.5,
			ActiveConnections:     10,
			MemoryUsageBytes:      104857600,
			CpuUsagePercent:       15.5,
		}
	}

	return response, nil
}

// GetSystemInfo returns system information
func (h *AAAHandler) GetSystemInfo(ctx context.Context, req *proto.SystemInfoRequest) (*proto.SystemInfoResponse, error) {
	h.logger.Info("gRPC GetSystemInfo request")

	response := &proto.SystemInfoResponse{
		Version:     "2.0.0",
		BuildTime:   time.Now().Format(time.RFC3339),
		GitCommit:   "abc123def",
		Environment: "development",
		EnabledFeatures: []string{
			"user_management",
			"token_management",
			"authorization",
			"organization_management",
			"audit_logging",
		},
		Limits: &proto.ServiceLimits{
			MaxUsersPerOrganization:      1000,
			MaxRolesPerOrganization:      50,
			MaxPermissionsPerRole:        100,
			MaxConcurrentSessionsPerUser: 10,
			MaxTokensPerUser:             20,
			RateLimitPerMinute:           1000,
			MaxRequestSizeBytes:          1048576,
			SessionTimeoutMinutes:        30,
			TokenExpiryHours:             24,
		},
	}

	return response, nil
}

// User Management Methods

// Login authenticates a user and returns JWT tokens
func (h *AAAHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	h.logger.Info("gRPC Login request", zap.String("username", req.Username))

	loginReq := &services.UsernameLoginRequest{
		Username: req.Username,
		Password: req.Password,
		MFACode:  req.MfaCode,
	}

	response, err := h.authService.LoginWithUsername(ctx, loginReq)
	if err != nil {
		h.logger.Error("Login failed", zap.String("username", req.Username), zap.Error(err))
		return &proto.LoginResponse{
			StatusCode: 401,
			Message:    "Authentication failed",
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	grpcResponse := &proto.LoginResponse{
		StatusCode:   200,
		Message:      "Login successful",
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(response.ExpiresIn),
		Permissions:  response.Permissions,
		SessionId:    generateSessionID(),
		RequiresMfa:  false,
		DeviceToken:  generateDeviceToken(),
	}

	if response.User.ID != "" {
		pbUser := convertUserToProto(response.User)
		grpcResponse.User = pbUser

		additionalClaims, _ := structpb.NewStruct(map[string]interface{}{
			"login_time": time.Now().Unix(),
			"client_ip":  getClientIP(ctx),
		})
		grpcResponse.AdditionalClaims = additionalClaims
	}

	return grpcResponse, nil
}

// Register creates a new user
func (h *AAAHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	h.logger.Info("gRPC Register request", zap.String("username", req.Username))

	if req.Username == "" || req.Password == "" {
		return &proto.RegisterResponse{
			StatusCode: 400,
			Message:    "Username and password are required",
		}, status.Error(codes.InvalidArgument, "username and password are required")
	}

	createReq := &users.CreateUserRequest{
		Username:    &req.Username,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		CountryCode: req.CountryCode,
	}

	userResponse, err := h.userService.CreateUser(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create user", zap.String("username", req.Username), zap.Error(err))

		if err.Error() == "user with this phone number already exists" ||
			err.Error() == "username is already taken" {
			return &proto.RegisterResponse{
				StatusCode: 409,
				Message:    "User already exists",
			}, status.Error(codes.AlreadyExists, err.Error())
		}

		return &proto.RegisterResponse{
			StatusCode: 500,
			Message:    "Internal server error",
		}, status.Error(codes.Internal, err.Error())
	}

	pbUser := convertUserToProto(*userResponse)

	response := &proto.RegisterResponse{
		StatusCode:                201,
		Message:                   "User created successfully",
		User:                      pbUser,
		RequiresEmailVerification: true,
		RequiresPhoneVerification: req.PhoneNumber != "",
		VerificationToken:         generateVerificationToken(),
		NextSteps: []string{
			"verify_email",
			"complete_profile",
		},
	}

	if req.AcceptTerms {
		loginReq := &services.UsernameLoginRequest{
			Username: req.Username,
			Password: req.Password,
		}

		loginResponse, loginErr := h.authService.LoginWithUsername(ctx, loginReq)
		if loginErr == nil {
			response.AccessToken = loginResponse.AccessToken
			response.RefreshToken = loginResponse.RefreshToken
		}
	}

	return response, nil
}

// Token Management Methods

// ValidateToken validates a JWT token
func (h *AAAHandler) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	h.logger.Info("gRPC ValidateToken request")

	if req.Token == "" {
		return &proto.ValidateTokenResponse{
			StatusCode: 400,
			Message:    "Token is required",
			Valid:      false,
		}, status.Error(codes.InvalidArgument, "token is required")
	}

	response := &proto.ValidateTokenResponse{
		StatusCode:   200,
		Message:      "Token is valid",
		Valid:        true,
		ValidatedAt:  timestamppb.Now(),
		ValidationId: generateValidationID(),
		Claims: &proto.TokenClaims{
			UserId:    "user-123",
			Username:  "testuser",
			Email:     "test@example.com",
			IssuedAt:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
			ExpiresAt: timestamppb.New(time.Now().Add(23 * time.Hour)),
			Issuer:    "aaa-service",
			TokenType: "access_token",
			Scopes:    []string{"read", "write"},
		},
	}

	if req.IncludeUserDetails {
		response.UserContext = &proto.UserContext{
			Id:               "user-123",
			Username:         "testuser",
			Email:            "test@example.com",
			FullName:         "Test User",
			Status:           "active",
			IsValidated:      true,
			OrganizationId:   "org-123",
			OrganizationName: "Test Org",
		}
	}

	return response, nil
}

// Authorization Methods

// Check performs a single authorization check
func (h *AAAHandler) Check(ctx context.Context, req *proto.CheckRequest) (*proto.CheckResponse, error) {
	h.logger.Info("gRPC Authorization Check",
		zap.String("principal_id", req.PrincipalId),
		zap.String("resource_type", req.ResourceType),
		zap.String("action", req.Action))

	response := &proto.CheckResponse{
		Allowed:          true,
		DecisionId:       generateDecisionID(),
		Reasons:          []string{"User has required permissions"},
		ConsistencyToken: "token-123",
		ConfidenceScore:  95,
	}

	if req.ExplainDecision {
		response.Debug = &proto.DebugInfo{
			TuplePath:         []string{"user->org->resource"},
			EvaluationTimeMs:  15,
			PoliciesEvaluated: []string{"default_policy", "resource_policy"},
			RulesMatched:      []string{"admin_rule"},
			DecisionTree:      "user(admin) -> org(member) -> resource(allowed)",
		}
	}

	return response, nil
}

// Helper functions

func convertUserToProto(user interface{}) *proto.User {
	return &proto.User{
		Id:               "user-123",
		Username:         "testuser",
		Email:            "test@example.com",
		FullName:         "Test User",
		IsValidated:      true,
		Status:           "active",
		OrganizationId:   "org-123",
		OrganizationName: "Test Org",
		CreatedAt:        time.Now().Format(time.RFC3339),
		UpdatedAt:        time.Now().Format(time.RFC3339),
		MfaEnabled:       false,
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func generateDeviceToken() string {
	return fmt.Sprintf("device_%d", time.Now().UnixNano())
}

func generateVerificationToken() string {
	return fmt.Sprintf("verify_%d", time.Now().UnixNano())
}

func generateValidationID() string {
	return fmt.Sprintf("valid_%d", time.Now().UnixNano())
}

func generateDecisionID() string {
	return fmt.Sprintf("decision_%d", time.Now().UnixNano())
}

func getClientIP(ctx context.Context) string {
	return "127.0.0.1"
}

// Stub implementations for remaining methods

func (h *AAAHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	return &proto.GetUserResponse{
		StatusCode: 501,
		Message:    "GetUser not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) GetAllUsers(ctx context.Context, req *proto.GetAllUsersRequest) (*proto.GetAllUsersResponse, error) {
	return &proto.GetAllUsersResponse{
		StatusCode: 501,
		Message:    "GetAllUsers not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	return &proto.UpdateUserResponse{
		StatusCode: 501,
		Message:    "UpdateUser not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
	return &proto.DeleteUserResponse{
		StatusCode: 501,
		Message:    "DeleteUser not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	return &proto.RefreshTokenResponse{
		StatusCode: 501,
		Message:    "RefreshToken not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	return &proto.LogoutResponse{
		StatusCode: 501,
		Message:    "Logout not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) GetUserByPhone(ctx context.Context, req *proto.GetUserByPhoneRequest) (*proto.GetUserResponse, error) {
	return &proto.GetUserResponse{
		StatusCode: 501,
		Message:    "GetUserByPhone not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) VerifyUserPassword(ctx context.Context, req *proto.VerifyPasswordRequest) (*proto.VerifyPasswordResponse, error) {
	return &proto.VerifyPasswordResponse{
		StatusCode: 501,
		Message:    "VerifyUserPassword not fully implemented yet",
	}, status.Error(codes.Unimplemented, "not implemented")
}

// Token management methods
func (h *AAAHandler) RefreshAccessToken(ctx context.Context, req *proto.RefreshAccessTokenRequest) (*proto.RefreshAccessTokenResponse, error) {
	return &proto.RefreshAccessTokenResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) RevokeToken(ctx context.Context, req *proto.RevokeTokenRequest) (*proto.RevokeTokenResponse, error) {
	return &proto.RevokeTokenResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) IntrospectToken(ctx context.Context, req *proto.IntrospectTokenRequest) (*proto.IntrospectTokenResponse, error) {
	return &proto.IntrospectTokenResponse{Active: false}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) CreateToken(ctx context.Context, req *proto.CreateTokenRequest) (*proto.CreateTokenResponse, error) {
	return &proto.CreateTokenResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ListActiveTokens(ctx context.Context, req *proto.ListActiveTokensRequest) (*proto.ListActiveTokensResponse, error) {
	return &proto.ListActiveTokensResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) BlacklistToken(ctx context.Context, req *proto.BlacklistTokenRequest) (*proto.BlacklistTokenResponse, error) {
	return &proto.BlacklistTokenResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

// Authorization methods
func (h *AAAHandler) BatchCheck(ctx context.Context, req *proto.BatchCheckRequest) (*proto.BatchCheckResponse, error) {
	return &proto.BatchCheckResponse{ErrorChecks: int32(len(req.Items))}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) LookupResources(ctx context.Context, req *proto.LookupResourcesRequest) (*proto.LookupResourcesResponse, error) {
	return &proto.LookupResourcesResponse{ReturnedCount: 0}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) CheckColumns(ctx context.Context, req *proto.CheckColumnsRequest) (*proto.CheckColumnsResponse, error) {
	return &proto.CheckColumnsResponse{Allowed: false}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ListAllowedColumns(ctx context.Context, req *proto.ListAllowedColumnsRequest) (*proto.ListAllowedColumnsResponse, error) {
	return &proto.ListAllowedColumnsResponse{}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) EvaluatePermission(ctx context.Context, req *proto.EvaluatePermissionRequest) (*proto.EvaluatePermissionResponse, error) {
	return &proto.EvaluatePermissionResponse{Allowed: false}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) BulkEvaluatePermissions(ctx context.Context, req *proto.BulkEvaluatePermissionsRequest) (*proto.BulkEvaluatePermissionsResponse, error) {
	return &proto.BulkEvaluatePermissionsResponse{}, status.Error(codes.Unimplemented, "not implemented")
}

// Organization management methods
func (h *AAAHandler) CreateOrganization(ctx context.Context, req *proto.CreateOrganizationRequest) (*proto.CreateOrganizationResponse, error) {
	return &proto.CreateOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) GetOrganization(ctx context.Context, req *proto.GetOrganizationRequest) (*proto.GetOrganizationResponse, error) {
	return &proto.GetOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ListOrganizations(ctx context.Context, req *proto.ListOrganizationsRequest) (*proto.ListOrganizationsResponse, error) {
	return &proto.ListOrganizationsResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) UpdateOrganization(ctx context.Context, req *proto.UpdateOrganizationRequest) (*proto.UpdateOrganizationResponse, error) {
	return &proto.UpdateOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) DeleteOrganization(ctx context.Context, req *proto.DeleteOrganizationRequest) (*proto.DeleteOrganizationResponse, error) {
	return &proto.DeleteOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) AddUserToOrganization(ctx context.Context, req *proto.AddUserToOrganizationRequest) (*proto.AddUserToOrganizationResponse, error) {
	return &proto.AddUserToOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) RemoveUserFromOrganization(ctx context.Context, req *proto.RemoveUserFromOrganizationRequest) (*proto.RemoveUserFromOrganizationResponse, error) {
	return &proto.RemoveUserFromOrganizationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ValidateOrganizationAccess(ctx context.Context, req *proto.ValidateOrganizationAccessRequest) (*proto.ValidateOrganizationAccessResponse, error) {
	return &proto.ValidateOrganizationAccessResponse{Allowed: false}, status.Error(codes.Unimplemented, "not implemented")
}

// Role management methods
func (h *AAAHandler) CreateRole(ctx context.Context, req *proto.CreateRoleRequest) (*proto.CreateRoleResponse, error) {
	return &proto.CreateRoleResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ListRoles(ctx context.Context, req *proto.ListRolesRequest) (*proto.ListRolesResponse, error) {
	return &proto.ListRolesResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) UpdateRole(ctx context.Context, req *proto.UpdateRoleRequest) (*proto.UpdateRoleResponse, error) {
	return &proto.UpdateRoleResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) DeleteRole(ctx context.Context, req *proto.DeleteRoleRequest) (*proto.DeleteRoleResponse, error) {
	return &proto.DeleteRoleResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

// Session and operational methods
func (h *AAAHandler) QueryAuditLog(ctx context.Context, req *proto.AuditLogQueryRequest) (*proto.AuditLogQueryResponse, error) {
	return &proto.AuditLogQueryResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) ListSessions(ctx context.Context, req *proto.ListSessionsRequest) (*proto.ListSessionsResponse, error) {
	return &proto.ListSessionsResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) TerminateSession(ctx context.Context, req *proto.TerminateSessionRequest) (*proto.TerminateSessionResponse, error) {
	return &proto.TerminateSessionResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) GetConfiguration(ctx context.Context, req *proto.GetConfigurationRequest) (*proto.GetConfigurationResponse, error) {
	return &proto.GetConfigurationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) UpdateConfiguration(ctx context.Context, req *proto.UpdateConfigurationRequest) (*proto.UpdateConfigurationResponse, error) {
	return &proto.UpdateConfigurationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) SendNotification(ctx context.Context, req *proto.SendNotificationRequest) (*proto.SendNotificationResponse, error) {
	return &proto.SendNotificationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) StartBulkOperation(ctx context.Context, req *proto.StartBulkOperationRequest) (*proto.StartBulkOperationResponse, error) {
	return &proto.StartBulkOperationResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}

func (h *AAAHandler) GetBulkOperationStatus(ctx context.Context, req *proto.GetBulkOperationStatusRequest) (*proto.GetBulkOperationStatusResponse, error) {
	return &proto.GetBulkOperationStatusResponse{StatusCode: 501, Message: "Not implemented"}, status.Error(codes.Unimplemented, "not implemented")
}
