package grpc_server

import (
	"context"
	"fmt"

	"github.com/Kisanlink/aaa-service/v2/internal/services"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthorizationChecker provides authorization checking for gRPC handlers
type AuthorizationChecker struct {
	authzService *services.AuthorizationService
	logger       *zap.Logger
}

// NewAuthorizationChecker creates a new authorization checker
func NewAuthorizationChecker(
	authzService *services.AuthorizationService,
	logger *zap.Logger,
) *AuthorizationChecker {
	return &AuthorizationChecker{
		authzService: authzService,
		logger:       logger,
	}
}

// CheckSeedPermission validates that the caller has permission to seed roles
// Authorization rules:
//  1. Caller must be authenticated (user or service)
//  2. Caller must have "catalog:seed" or "admin:*" permission
//  3. If serviceID is provided, caller must either:
//     a) Have super_admin role, OR
//     b) Be the service that matches serviceID (services can only seed their own roles)
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

	// Check basic permission: catalog:seed
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

	// For service-specific seeding, enforce ownership rules
	if err := ac.checkServiceOwnership(ctx, principalID, principalType, targetServiceID); err != nil {
		return err
	}

	ac.logger.Info("Seed permission granted",
		zap.String("principal_id", principalID),
		zap.String("principal_type", principalType),
		zap.String("target_service_id", targetServiceID))

	return nil
}

// checkServiceOwnership validates service-specific seed authorization
// Rules:
// 1. If principal is a service, it can only seed its own roles (service_id must match)
// 2. If principal is a user, check for super_admin or admin:* permission
func (ac *AuthorizationChecker) checkServiceOwnership(
	ctx context.Context,
	principalID string,
	principalType string,
	targetServiceID string,
) error {
	// If the principal is a service, it can only seed its own roles
	if principalType == "service" {
		// Extract service_id from context
		serviceID := ac.getContextValue(ctx, "service_id")
		if serviceID != targetServiceID {
			ac.logger.Warn("Service attempting to seed another service's roles",
				zap.String("caller_service_id", serviceID),
				zap.String("target_service_id", targetServiceID))
			return status.Errorf(codes.PermissionDenied,
				"services can only seed their own roles (attempted to seed %s, but caller is %s)",
				targetServiceID, serviceID)
		}

		ac.logger.Debug("Service ownership validated",
			zap.String("service_id", serviceID),
			zap.String("target_service_id", targetServiceID))
		return nil
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
		return serviceID, "service", nil
	}

	// Check if this is a user principal
	userID := ac.getContextValue(ctx, "user_id")
	if userID == "" {
		return "", "", fmt.Errorf("no authenticated principal found in context")
	}

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
