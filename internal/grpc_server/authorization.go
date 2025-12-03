package grpc_server

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/authorization"
	"github.com/Kisanlink/aaa-service/v2/internal/config"
	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationChecker provides authorization checking for gRPC handlers
type AuthorizationChecker struct {
	authzService      *services.AuthorizationService
	serviceAuthorizer *authorization.ServiceAuthorizer
	logger            *zap.Logger
}

// NewAuthorizationChecker creates a new authorization checker
func NewAuthorizationChecker(
	authzService *services.AuthorizationService,
	logger *zap.Logger,
) *AuthorizationChecker {
	// Load service authorization configuration
	serviceAuthConfig, err := config.LoadServiceAuthorizationConfig()
	if err != nil {
		logger.Warn("Failed to load service authorization config, using defaults",
			zap.Error(err))
		// Create default config that allows all
		serviceAuthConfig = &config.ServiceAuthorizationConfig{
			ServiceAuthorization: config.ServiceAuthSection{
				Enabled:  false,
				Services: make(map[string]config.ServicePermission),
			},
			DefaultBehavior: config.DefaultBehavior{
				WhenDisabled:            "allow_all",
				LogUnauthorizedAttempts: true,
			},
		}
	}

	// Create service authorizer
	serviceAuthorizer := authorization.NewServiceAuthorizer(serviceAuthConfig, logger)

	return &AuthorizationChecker{
		authzService:      authzService,
		serviceAuthorizer: serviceAuthorizer,
		logger:            logger,
	}
}

// CheckSeedPermission validates that the caller has permission to seed roles
//
// ENHANCED IMPLEMENTATION - Configuration-Based Authorization:
// 1. Services use configuration-based permission checking via ServiceAuthorizer
// 2. User principals still use RBAC permission checks
// 3. Service principals must have catalog:seed_roles permission in config
// 4. Backward compatible - falls back to old behavior if config not loaded
//
// Authorization Rules:
// SERVICE PRINCIPALS (authenticated via API key):
//   - Must be configured in service_permissions.yaml
//   - Must have catalog:seed_roles or catalog:* permission
//   - API key validation if api_key_required is true
//   - Can only seed their own roles (service_name must match targetServiceID)
//
// USER PRINCIPALS (authenticated via JWT):
//   - Must have catalog:seed permission
//   - Can seed default/farmers-module with basic permission
//   - Need admin:* permission for service-specific seeding
func (ac *AuthorizationChecker) CheckSeedPermission(ctx context.Context, targetServiceID string) error {
	// Extract principal information from context
	principalID, principalType, err := ac.extractPrincipal(ctx)
	if err != nil {
		ac.logger.Error("Failed to extract principal from context", zap.Error(err))
		return status.Errorf(codes.Unauthenticated, "authentication required")
	}

	ac.logger.Debug("Checking seed permission",
		zap.String("principal_id", principalID),
		zap.String("principal_type", principalType),
		zap.String("target_service_id", targetServiceID))

	// Identify if this request is from a service (either via principal_type OR explicit context values)
	serviceName := ac.getContextValue(ctx, "service_name")
	isServicePrincipal := principalType == "service" || serviceName != ""

	// Services use configuration-based authorization
	if isServicePrincipal {
		if serviceName == "" {
			// Fallback: try to resolve service name from context if not already available
			serviceName = ac.getContextValue(ctx, "service_id")
		}

		if serviceName == "" {
			ac.logger.Error("Service name missing in context (service principal detected)",
				zap.String("principal_id", principalID),
				zap.String("principal_type", principalType),
				zap.String("target_service_id", targetServiceID))
			return status.Errorf(codes.Unauthenticated,
				"service authentication incomplete: service_name missing")
		}

		// Use configuration-based authorization
		// Check if service has catalog:seed_roles permission
		if err := ac.serviceAuthorizer.Authorize(ctx, serviceName, "catalog:seed_roles"); err != nil {
			ac.logger.Warn("Service authorization failed",
				zap.String("service_id", principalID),
				zap.String("service_name", serviceName),
				zap.String("target_service_id", targetServiceID),
				zap.Error(err))
			return status.Errorf(codes.PermissionDenied,
				"service '%s' is not authorized to seed roles: %v", serviceName, err)
		}

		// Check ownership: service can only seed its own roles
		// Normalize names for comparison (handle "Farmers Module" vs "farmers-module")
		if !normalizeServiceName(serviceName, targetServiceID) {
			ac.logger.Warn("Service attempting to seed another service's roles",
				zap.String("service_id", principalID),
				zap.String("caller_service_name", serviceName),
				zap.String("target_service_id", targetServiceID))
			return status.Errorf(codes.PermissionDenied,
				"service '%s' cannot seed roles for service '%s'",
				serviceName, targetServiceID)
		}

		// Service is authorized to seed its own roles
		ac.logger.Info("Service authorized to seed own roles via configuration",
			zap.String("service_id", principalID),
			zap.String("service_name", serviceName),
			zap.String("target_service_id", targetServiceID))
		return nil
	}

	// For users, check catalog:seed permission
	permission := &services.Permission{
		UserID:     principalID,
		Resource:   "catalog",
		ResourceID: "catalog", // Use resource type as resource ID for general permissions
		Action:     "seed",
	}

	result, err := ac.authzService.CheckPermission(ctx, permission)
	if err != nil {
		ac.logger.Error("Permission check failed",
			zap.String("principal_id", principalID),
			zap.String("resource", "catalog"),
			zap.String("action", "seed"),
			zap.Error(err))
		return status.Errorf(codes.Internal, "authorization check failed: %v", err)
	}

	if !result.Allowed {
		ac.logger.Warn("Seed permission denied - insufficient permissions",
			zap.String("principal_id", principalID),
			zap.String("principal_type", principalType),
			zap.String("reason", result.Reason))
		return status.Errorf(codes.PermissionDenied,
			"insufficient permissions to seed roles: %s", result.Reason)
	}

	// If no specific service is targeted (empty service_id defaults to farmers-module),
	// allow if they have basic seed permission
	if targetServiceID == "" || targetServiceID == "farmers-module" {
		ac.logger.Debug("Seed permission granted for default/farmers-module",
			zap.String("principal_id", principalID))
		return nil
	}

	// For service-specific seeding, users need admin permission
	if err := ac.checkServiceOwnership(ctx, principalID, principalType, targetServiceID); err != nil {
		return err
	}

	ac.logger.Info("Seed permission granted",
		zap.String("principal_id", principalID),
		zap.String("principal_type", principalType),
		zap.String("target_service_id", targetServiceID))

	return nil
}

// checkServiceOwnership validates that users have admin permissions for service-specific seeding
//
// IMPORTANT: This function is ONLY called for user principals attempting to seed service-specific roles.
// Service principals are validated earlier in CheckSeedPermission (lines 62-98) and never reach this function.
//
// Rule: Users must have admin:* permission to seed service-specific roles (not farmers-module)
func (ac *AuthorizationChecker) checkServiceOwnership(
	ctx context.Context,
	principalID string,
	principalType string,
	targetServiceID string,
) error {
	// NOTE: Service principals never reach this function
	// They are validated earlier in CheckSeedPermission (lines 62-98) and return early on line 98.
	// This function is ONLY called for user principals attempting service-specific seeding.
	// If a service somehow reaches here, it's a programming error that should be caught.
	if principalType == "service" {
		ac.logger.Error("UNEXPECTED: Service principal reached checkServiceOwnership",
			zap.String("service_id", principalID),
			zap.String("target_service_id", targetServiceID),
			zap.String("note", "Services should return early in CheckSeedPermission"))
		return status.Errorf(codes.Internal,
			"internal error: service principal should not reach this code path")
	}

	// If the principal is a user, check for admin permissions
	adminPermission := &services.Permission{
		UserID:     principalID,
		Resource:   "admin",
		ResourceID: "admin",
		Action:     "*", // Wildcard for all admin actions
	}

	result, err := ac.authzService.CheckPermission(ctx, adminPermission)
	if err != nil {
		ac.logger.Error("Admin permission check failed",
			zap.String("user_id", principalID),
			zap.Error(err))
		return status.Errorf(codes.Internal, "admin authorization check failed: %v", err)
	}

	if !result.Allowed {
		ac.logger.Warn("User attempting to seed service roles without admin permission",
			zap.String("user_id", principalID),
			zap.String("target_service_id", targetServiceID),
			zap.String("reason", result.Reason))
		return status.Errorf(codes.PermissionDenied,
			"only administrators can seed roles for services: %s", result.Reason)
	}

	ac.logger.Debug("Admin permission validated for cross-service seeding",
		zap.String("user_id", principalID),
		zap.String("target_service_id", targetServiceID))

	return nil
}

// extractPrincipal extracts the authenticated principal from context
//
// ENHANCED IMPLEMENTATION - Additional Validation:
// - Validates service_name is present for service principals
// - service_name is critical for authorization decisions
//
// Returns (principalID, principalType, error)
// principalType is either "service" or "user"
func (ac *AuthorizationChecker) extractPrincipal(ctx context.Context) (string, string, error) {
	// Check if this is a service principal
	principalType := ac.getContextValue(ctx, "principal_type")
	if principalType == "service" {
		serviceID := ac.getContextValue(ctx, "service_id")
		if serviceID == "" {
			return "", "", fmt.Errorf("service principal_type set but service_id missing")
		}

		// ENHANCED: Also validate service_name is present
		// This is critical for authorization decisions
		serviceName := ac.getContextValue(ctx, "service_name")
		if serviceName == "" {
			ac.logger.Error("Service authentication incomplete",
				zap.String("service_id", serviceID),
				zap.String("principal_type", principalType))
			return "", "", fmt.Errorf("service principal_type set but service_name missing")
		}

		ac.logger.Debug("Service principal extracted",
			zap.String("service_id", serviceID),
			zap.String("service_name", serviceName))
		return serviceID, "service", nil
	}

	// Check if this is a user principal
	userID := ac.getContextValue(ctx, "user_id")
	if userID == "" {
		return "", "", fmt.Errorf("no authenticated principal found in context")
	}

	ac.logger.Debug("User principal extracted",
		zap.String("user_id", userID))
	return userID, "user", nil
}

// getContextValue safely extracts a string value from context
func (ac *AuthorizationChecker) getContextValue(ctx context.Context, key string) string {
	if val := ctx.Value(key); val != nil {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// normalizeServiceName compares two service names after normalization
// Handles cases like "Farmers Module" vs "farmers-module"
func normalizeServiceName(name1, name2 string) bool {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, " ", "-")
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}
	return normalize(name1) == normalize(name2)
}
